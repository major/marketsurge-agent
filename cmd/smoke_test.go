//go:build smoke

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	smokeCommandTimeout = 90 * time.Second
	smokeSchemaTimeout  = 30 * time.Second
)

// smokeCase describes one live CLI invocation that should return a JSON envelope.
type smokeCase struct {
	Name       string
	Title      string
	Args       []string
	SymbolMeta string
	Skip       string
}

// TestSmokeCommandSchemaCoverage fails when a runnable API command lacks a smoke case.
func TestSmokeCommandSchemaCoverage(t *testing.T) {
	schemas := smokeJSONSchemas(t)
	actual := smokeAPILeafTitles(schemas)
	expected := smokeCoveredTitles()

	assert.Equal(t, expected, actual, "live smoke cases should cover every API leaf command exposed by --jsonschema")
}

// TestSmokeCommandsAgainstLiveData runs curated command invocations against MarketSurge.
func TestSmokeCommandsAgainstLiveData(t *testing.T) {
	for _, tc := range smokeCases() {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Skip != "" {
				t.Skip(tc.Skip)
			}

			output, err := smokeRunCommand(t, tc.Args...)
			if smokeShouldSkipAuthFailure(output, err) {
				t.Skip("MarketSurge auth is not configured or the local Firefox session is not usable; sign in with Firefox or set MARKETSURGE_AGENT_COOKIE_DB")
			}
			if smokeShouldSkipRateLimit(output, err) {
				t.Skipf("MarketSurge rate limited the smoke command: %s", strings.TrimSpace(string(output)))
			}
			require.NoError(t, err, string(output))

			envelope := smokeParseEnvelope(t, output)
			assert.Contains(t, envelope, "data")
			assert.Contains(t, envelope, "metadata")

			if tc.SymbolMeta != "" {
				metadata, ok := envelope["metadata"].(map[string]any)
				require.True(t, ok, "metadata should be an object")
				assert.Equal(t, tc.SymbolMeta, metadata["symbol"])
			}
		})
	}
}

// smokeCases returns the curated local smoke test matrix.
func smokeCases() []smokeCase {
	return []smokeCase{
		{Name: "stock_get", Title: "marketsurge-agent stock get", Args: []string{"stock", "get", "AAPL"}, SymbolMeta: "AAPL"},
		{Name: "stock_analyze", Title: "marketsurge-agent stock analyze", Args: []string{"stock", "analyze", "--summary", "AAPL", "MSFT"}},
		{Name: "fundamental_get", Title: "marketsurge-agent fundamental get", Args: []string{"fundamental", "get", "AAPL"}, SymbolMeta: "AAPL"},
		{Name: "ownership_get", Title: "marketsurge-agent ownership get", Args: []string{"ownership", "get", "AAPL"}, SymbolMeta: "AAPL"},
		{Name: "rs_history_get", Title: "marketsurge-agent rs-history get", Args: []string{"rs-history", "get", "AAPL"}, SymbolMeta: "AAPL"},
		{Name: "chart_history", Title: "marketsurge-agent chart history", Args: []string{"chart", "history", "AAPL", "--lookback", "1M"}, SymbolMeta: "AAPL"},
		{Name: "chart_markups", Title: "marketsurge-agent chart markups", Args: []string{"chart", "markups", "AAPL"}, SymbolMeta: "AAPL"},
		{Name: "catalog_list", Title: "marketsurge-agent catalog list", Args: []string{"catalog", "list"}},
		smokeCatalogRunCase(),
	}
}

// smokeCatalogRunCase builds the account-specific catalog run smoke case.
func smokeCatalogRunCase() smokeCase {
	if value := os.Getenv("MARKETSURGE_SMOKE_WATCHLIST_ID"); value != "" {
		return smokeCase{
			Name:  "catalog_run_watchlist",
			Title: "marketsurge-agent catalog run",
			Args:  []string{"catalog", "run", "--kind", "watchlist", "--watchlist-id", value, "--limit", "5"},
		}
	}
	if value := os.Getenv("MARKETSURGE_SMOKE_REPORT_ID"); value != "" {
		return smokeCase{
			Name:  "catalog_run_report",
			Title: "marketsurge-agent catalog run",
			Args:  []string{"catalog", "run", "--kind", "report", "--report-id", value, "--limit", "5"},
		}
	}
	if value := os.Getenv("MARKETSURGE_SMOKE_COACH_SCREEN_ID"); value != "" {
		return smokeCase{
			Name:  "catalog_run_coach_screen",
			Title: "marketsurge-agent catalog run",
			Args:  []string{"catalog", "run", "--kind", "coach_screen", "--coach-screen-id", value, "--limit", "5"},
		}
	}

	return smokeCase{
		Name:  "catalog_run",
		Title: "marketsurge-agent catalog run",
		Skip:  "catalog run needs one account-specific ID: set MARKETSURGE_SMOKE_WATCHLIST_ID, MARKETSURGE_SMOKE_REPORT_ID, or MARKETSURGE_SMOKE_COACH_SCREEN_ID",
	}
}

// smokeCoveredTitles returns sorted schema titles that have smoke case coverage.
func smokeCoveredTitles() []string {
	titles := make([]string, 0, len(smokeCases()))
	for _, tc := range smokeCases() {
		titles = append(titles, tc.Title)
	}
	slices.Sort(titles)
	return slices.Compact(titles)
}

// smokeJSONSchemas returns the full command schema from the local CLI.
func smokeJSONSchemas(t *testing.T) []map[string]any {
	t.Helper()

	output, err := smokeRunCommandWithTimeout(t, smokeSchemaTimeout, "--jsonschema")
	require.NoError(t, err, string(output))

	var schemas []map[string]any
	require.NoError(t, json.Unmarshal(output, &schemas), string(output))
	require.NotEmpty(t, schemas)
	return schemas
}

// smokeAPILeafTitles extracts runnable API leaf command titles from schema output.
func smokeAPILeafTitles(schemas []map[string]any) []string {
	titles := []string{}
	for _, schema := range schemas {
		title, _ := schema["title"].(string)
		if title == "" || !strings.HasPrefix(title, "marketsurge-agent ") {
			continue
		}
		if _, hasSubcommands := schema["x-structcli-subcommands"]; hasSubcommands {
			continue
		}
		if smokeIsNonAPICommandTitle(title) {
			continue
		}
		titles = append(titles, title)
	}
	slices.Sort(titles)
	return titles
}

// smokeIsNonAPICommandTitle reports whether a schema title belongs to a helper command.
func smokeIsNonAPICommandTitle(title string) bool {
	for _, prefix := range []string{
		"marketsurge-agent completion ",
		"marketsurge-agent config-keys",
		"marketsurge-agent env-vars",
		"marketsurge-agent help",
	} {
		if strings.HasPrefix(title, prefix) {
			return true
		}
	}
	return false
}

// smokeRunCommand executes a live CLI command with the default smoke timeout.
func smokeRunCommand(t *testing.T, args ...string) ([]byte, error) {
	t.Helper()
	return smokeRunCommandWithTimeout(t, smokeCommandTimeout, args...)
}

// smokeRunCommandWithTimeout executes a CLI command as a subprocess.
func smokeRunCommandWithTimeout(t *testing.T, timeout time.Duration, args ...string) ([]byte, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	cmdArgs := append([]string{"run", "./marketsurge-agent"}, args...)
	cmd := exec.CommandContext(ctx, "go", cmdArgs...)
	cmd.Env = os.Environ()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() != nil {
		return append(stdout.Bytes(), stderr.Bytes()...), fmt.Errorf("smoke command timed out after %s: %w", timeout, ctx.Err())
	}
	if err != nil && stderr.Len() > 0 {
		return stderr.Bytes(), err
	}
	return stdout.Bytes(), err
}

// smokeParseEnvelope decodes a command response as a JSON object.
func smokeParseEnvelope(t *testing.T, output []byte) map[string]any {
	t.Helper()

	var envelope map[string]any
	require.NoError(t, json.Unmarshal(output, &envelope), string(output))
	require.NotEmpty(t, envelope)
	return envelope
}

// smokeShouldSkipAuthFailure reports whether an auth failure should skip the local suite.
func smokeShouldSkipAuthFailure(output []byte, err error) bool {
	if err == nil || smokeAuthConfigured() {
		return false
	}
	return smokeErrorCode(output) == "AUTH_FAILED"
}

// smokeAuthConfigured reports whether the user explicitly configured local auth.
func smokeAuthConfigured() bool {
	return os.Getenv("MARKETSURGE_AGENT_COOKIE_DB") != "" || os.Getenv("MARKETSURGE_AGENT_COOKIEDB") != "" || os.Getenv("MARKETSURGE_AGENT_CONFIG") != ""
}

// smokeShouldSkipRateLimit reports whether MarketSurge rate limited the live smoke run.
func smokeShouldSkipRateLimit(output []byte, err error) bool {
	if err == nil {
		return false
	}
	if smokeErrorCode(output) != "HTTP_ERROR" {
		return false
	}
	return strings.Contains(strings.ToLower(string(output)), "rate") || strings.Contains(string(output), "429")
}

// smokeErrorCode extracts an error envelope code from command output.
func smokeErrorCode(output []byte) string {
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(output, &envelope); err != nil {
		return ""
	}
	return envelope.Error.Code
}

// TestSmokeCasesHaveUniqueNames protects subtest output from ambiguous duplicate names.
func TestSmokeCasesHaveUniqueNames(t *testing.T) {
	seen := map[string]bool{}
	for _, tc := range smokeCases() {
		if seen[tc.Name] {
			t.Fatalf("duplicate smoke case name %q", tc.Name)
		}
		seen[tc.Name] = true
	}
}

// TestSmokeCoverageExpectationStaysSorted verifies deterministic comparison inputs.
func TestSmokeCoverageExpectationStaysSorted(t *testing.T) {
	titles := smokeCoveredTitles()
	sorted := slices.Clone(titles)
	slices.Sort(sorted)
	assert.Equal(t, sorted, titles, "smoke covered titles should stay sorted")
}
