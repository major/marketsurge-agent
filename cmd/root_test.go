package cmd

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootJSONSchema(t *testing.T) {
	cmd := rootExecCommand(t, "--jsonschema")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	schemas := parseJSONSchemas(t, output)
	require.Greater(t, len(schemas), 1, "bare --jsonschema should return the full command tree")

	rootSchema := schemaJSONByTitle(t, schemas, "marketsurge-agent")
	assert.Contains(t, rootSchema, `"x-structcli-env-prefix": "MARKETSURGE_AGENT"`)
	assert.Contains(t, rootSchema, `"MARKETSURGE_AGENT_COOKIE_DB"`)
	assert.Contains(t, rootSchema, `"MARKETSURGE_AGENT_VERBOSE"`)
	assert.Contains(t, rootSchema, `"x-structcli-config-flag": "config"`)

	schemaMapByTitle(t, schemas, "marketsurge-agent stock analyze")
	schemaMapByTitle(t, schemas, "marketsurge-agent chart history")
	schemaMapByTitle(t, schemas, "marketsurge-agent catalog run")
}

func TestRootJSONSchemaTreeMatchesDefault(t *testing.T) {
	defaultCmd := rootExecCommand(t, "--jsonschema")
	defaultOutput, err := defaultCmd.CombinedOutput()
	require.NoError(t, err, string(defaultOutput))

	treeCmd := rootExecCommand(t, "--jsonschema=tree")
	treeOutput, err := treeCmd.CombinedOutput()
	require.NoError(t, err, string(treeOutput))

	assert.JSONEq(t, string(defaultOutput), string(treeOutput))
}

func TestRootJSONSchemaIncludesEnumDescriptions(t *testing.T) {
	cmd := rootExecCommand(t, "--jsonschema")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	schemas := parseJSONSchemas(t, output)
	chartMarkups := schemaMapByTitle(t, schemas, "marketsurge-agent chart markups")
	frequency := schemaProperty(t, chartMarkups, "frequency")
	assert.Equal(t, []any{"DAILY", "WEEKLY"}, frequency["enum"])
	assert.Contains(t, frequency["description"], "Chart candle frequency for markup lookup")
	assert.Contains(t, frequency["description"], "{DAILY,WEEKLY}")
	sortDir := schemaProperty(t, chartMarkups, "sort-dir")
	assert.Equal(t, []any{"ASC", "DESC"}, sortDir["enum"])
	assert.Contains(t, sortDir["description"], "Sort direction for markup annotations")
	assert.Contains(t, sortDir["description"], "{ASC,DESC}")

	chartHistory := schemaMapByTitle(t, schemas, "marketsurge-agent chart history")
	lookback := schemaProperty(t, chartHistory, "lookback")
	assert.Contains(t, lookback["description"], "Relative lookback period")
	assert.Contains(t, lookback["description"], "1W, 1M, 3M, 6M, 1Y, YTD")
	period := schemaProperty(t, chartHistory, "period")
	assert.Equal(t, []any{"daily", "weekly"}, period["enum"])
	assert.Contains(t, period["description"], "Data period granularity")
	assert.Contains(t, period["description"], "{daily,weekly}")
	assert.Equal(t, "daily", period["default"])
}

func TestRootJSONSchemaIncludesAgentMetadata(t *testing.T) {
	cmd := rootExecCommand(t, "--jsonschema")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	schemas := parseJSONSchemas(t, output)
	stockAnalyze := schemaJSONByTitle(t, schemas, "marketsurge-agent stock analyze")
	assert.Contains(t, stockAnalyze, "one or more")
	assert.Contains(t, stockAnalyze, `"x-structcli-group": "Input"`)
	assert.Contains(t, stockAnalyze, `"x-structcli-group": "Output Format"`)
	assert.Contains(t, stockAnalyze, "symbols to analyze")

	catalogRun := schemaJSONByTitle(t, schemas, "marketsurge-agent catalog run")
	assert.Contains(t, catalogRun, `"x-structcli-group": "Catalog Selection"`)
	assert.Contains(t, catalogRun, `"x-structcli-group": "Kind-Specific IDs"`)
	assert.Contains(t, catalogRun, "required when kind=report")
	assert.Contains(t, catalogRun, "required when kind=watchlist")
	catalogRunSchema := schemaMapByTitle(t, schemas, "marketsurge-agent catalog run")
	kind := schemaProperty(t, catalogRunSchema, "kind")
	assert.Contains(t, kind["description"], "Required catalog kind to run")
	assert.Contains(t, kind["description"], "watchlist uses --watchlist-id")
	assert.Contains(t, kind["description"], "coach_screen uses --coach-screen-id")
	assert.Contains(t, kind["description"], "screens are list-only")
}

func TestRootJSONSchemaTreeIncludesComplexExamples(t *testing.T) {
	cmd := rootExecCommand(t, "--jsonschema=tree")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	var schemas []map[string]any
	require.NoError(t, json.Unmarshal(output, &schemas))

	chartHistory := schemaMapByTitle(t, schemas, "marketsurge-agent chart history")
	assert.Contains(t, chartHistory["description"], "chart history AAPL --lookback 3M")
	assert.Contains(t, schemaProperty(t, chartHistory, "start-date")["description"], "--start-date 2024-01-01 --end-date 2024-06-30")
	assert.Contains(t, schemaProperty(t, chartHistory, "lookback")["description"], "--lookback 3M")

	catalogRun := schemaMapByTitle(t, schemas, "marketsurge-agent catalog run")
	assert.Contains(t, catalogRun["description"], "catalog run --kind report --report-id")
	assert.Contains(t, schemaProperty(t, catalogRun, "kind")["description"], "report uses --report-id")
	assert.Contains(t, schemaProperty(t, catalogRun, "fields")["description"], "--fields symbol,price,composite_rating")

	stockAnalyze := schemaMapByTitle(t, schemas, "marketsurge-agent stock analyze")
	assert.Contains(t, stockAnalyze["description"], "stock analyze --summary")
	assert.Contains(t, schemaProperty(t, stockAnalyze, "setup")["description"], "base_length_weeks")
	assert.Contains(t, schemaProperty(t, stockAnalyze, "setup")["description"], "quarterly_funds")
	assert.Contains(t, schemaProperty(t, stockAnalyze, "flat")["description"], "pricing_market_cap")
}

func TestRootOptionsStructTags(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeFor[RootOptions]()
	tests := []struct {
		field string
		flag  string
		group string
		descr string
	}{
		{field: "CookieDB", flag: "cookie-db", group: "Authentication & Logging", descr: "Path to Firefox cookies.sqlite file; omit to auto-discover Firefox profiles"},
		{field: "Verbose", flag: "verbose", group: "Authentication & Logging", descr: "Enable verbose logging for auth and API requests"},
	}
	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			f, ok := typ.FieldByName(tt.field)
			require.True(t, ok, "field %s should exist", tt.field)
			assert.Equal(t, tt.flag, f.Tag.Get("flag"), "flag tag")
			assert.Equal(t, tt.group, f.Tag.Get("flaggroup"), "flaggroup tag")
			assert.Equal(t, tt.descr, f.Tag.Get("flagdescr"), "flagdescr tag")
			assert.Equal(t, "true", f.Tag.Get("flagenv"), "flagenv tag")
		})
	}
}

func TestRootDebugOptions(t *testing.T) {
	cmd := rootExecCommand(t, "--debug-options")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}

func TestRootHelpShowsConfigFlag(t *testing.T) {
	cmd := rootExecCommand(t, "--help")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	help := string(output)
	assert.Contains(t, help, "--config string")
	assert.Contains(t, help, "config file")
	assert.Contains(t, help, "--mcp")
	assert.Contains(t, help, "Sign into MarketSurge in Firefox")
}

func TestRootMCPInitializeAndToolsList(t *testing.T) {
	cmd := rootExecCommand(t, "--mcp")
	cmd.Stdin = strings.NewReader(
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"clientInfo":{"name":"test"}}}` + "\n" +
			`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
	)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	responses := decodeMCPResponses(t, output)
	require.Len(t, responses, 2)

	var initResult mcpInitializeResult
	require.NoError(t, json.Unmarshal(responses[0].Result, &initResult))
	assert.Equal(t, "marketsurge-agent", initResult.ServerInfo.Name)
	assert.True(t, strings.HasPrefix(initResult.ServerInfo.Version, "dev"), "MCP version = %q, want dev build version", initResult.ServerInfo.Version)

	var listResult mcpToolsListResult
	require.NoError(t, json.Unmarshal(responses[1].Result, &listResult))
	toolNames := mcpToolNames(listResult)
	assert.ElementsMatch(t, []string{
		"catalog_list",
		"catalog_run",
		"chart_history",
		"chart_markups",
		"fundamental_get",
		"ownership_get",
		"rs-history_get",
		"stock_analyze",
		"stock_get",
	}, toolNames)
	assert.NotContains(t, toolNames, "completion_bash")
	assert.NotContains(t, toolNames, "completion_fish")
	assert.NotContains(t, toolNames, "completion_powershell")
	assert.NotContains(t, toolNames, "completion_zsh")
	assert.NotContains(t, toolNames, "stock")
	assert.NotContains(t, toolNames, "chart")
	assert.NotContains(t, toolNames, "catalog")
	for _, tool := range listResult.Tools {
		assert.NotEmpty(t, tool.Description, "tool %s should have a description", tool.Name)
		require.NotEmpty(t, tool.InputSchema, "tool %s should expose an input schema", tool.Name)
		var schema map[string]any
		require.NoError(t, json.Unmarshal(tool.InputSchema, &schema), "tool %s input schema should be JSON", tool.Name)
		assert.Equal(t, "object", schema["type"], "tool %s schema should be an object", tool.Name)
		assert.Equal(t, "MARKETSURGE_AGENT", schema["x-structcli-env-prefix"], "tool %s schema should expose env metadata", tool.Name)
		assert.Equal(t, "config", schema["x-structcli-config-flag"], "tool %s schema should expose config metadata", tool.Name)
	}
	assert.NotContains(t, string(output), "Firefox")
	assert.NotContains(t, string(output), "cookie")
}

func TestRootMCPToolSchemasExposeFlagMetadata(t *testing.T) {
	cmd := rootExecCommand(t, "--mcp")
	cmd.Stdin = strings.NewReader(
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"clientInfo":{"name":"test"}}}` + "\n" +
			`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
	)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	responses := decodeMCPResponses(t, output)
	require.Len(t, responses, 2)

	var listResult mcpToolsListResult
	require.NoError(t, json.Unmarshal(responses[1].Result, &listResult))

	stockAnalyze := mcpToolByName(t, listResult, "stock_analyze")
	stockAnalyzeSchema := mcpInputSchema(t, stockAnalyze)
	stockAnalyzeProps := schemaProperties(t, stockAnalyzeSchema)
	assert.Contains(t, stockAnalyze.Description, "Fetches stock")
	assert.Contains(t, stockAnalyzeProps, "tickers")
	assert.Contains(t, stockAnalyzeProps, "compact")
	assert.Equal(t, "Input", stockAnalyzeProps["tickers"].(map[string]any)["x-structcli-group"])

	chartHistory := mcpToolByName(t, listResult, "chart_history")
	chartHistorySchema := mcpInputSchema(t, chartHistory)
	chartHistoryProps := schemaProperties(t, chartHistorySchema)
	assert.Contains(t, chartHistory.Description, "Fetches price history")
	assert.Contains(t, chartHistoryProps, "lookback")
	assert.Contains(t, chartHistoryProps, "start-date")
	assert.Contains(t, chartHistoryProps, "end-date")
	assert.Equal(t, "Date Range", chartHistoryProps["lookback"].(map[string]any)["x-structcli-group"])

	catalogRun := mcpToolByName(t, listResult, "catalog_run")
	catalogRunSchema := mcpInputSchema(t, catalogRun)
	catalogRunProps := schemaProperties(t, catalogRunSchema)
	assert.Contains(t, catalogRun.Description, "Runs a catalog entry")
	assert.Contains(t, catalogRunProps, "kind")
	assert.Contains(t, catalogRunProps, "report-id")
	assert.Contains(t, catalogRunProps, "watchlist-id")
	assert.Equal(t, "Catalog Selection", catalogRunProps["kind"].(map[string]any)["x-structcli-group"])
}

func TestRootMCPToolsCallUsesAPIAuth(t *testing.T) {
	cmd := rootExecCommand(t, "--mcp")
	cmd.Stdin = strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"stock_analyze","arguments":{"tickers":"AAPL"}}}` + "\n")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	responses := decodeMCPResponses(t, output)
	require.Len(t, responses, 1)
	require.Nil(t, responses[0].Error)

	var callResult mcpToolCallResult
	require.NoError(t, json.Unmarshal(responses[0].Result, &callResult))
	assert.True(t, callResult.IsError)
	require.Len(t, callResult.Content, 1)
	assert.Contains(t, callResult.Content[0].Text, "no JWT available")
	assert.Contains(t, callResult.Content[0].Text, "no Firefox profiles found")
	assert.Contains(t, callResult.Content[0].Text, "marketsurge-agent stock analyze")
}

func TestRootConfigFileLoading(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"
	cookiePath := tmpDir + "/cookies.sqlite"
	require.NoError(t, os.WriteFile(configPath, []byte("cookie-db: "+cookiePath+"\nverbose: true\n"), 0o600))

	cmd := rootExecCommand(t, "--config", configPath, "--debug-options=json")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	debugOutput := string(output)
	assert.Contains(t, debugOutput, `"name": "cookie-db"`)
	assert.Contains(t, debugOutput, `"value": "`+cookiePath+`"`)
	assert.Contains(t, debugOutput, `"source": "config"`)
	assert.Contains(t, debugOutput, `"verbose": true`)
	assert.NotContains(t, debugOutput, `"jwt"`)
}

func TestRootConfigKeysListCookieAuthOnly(t *testing.T) {
	cmd := rootExecCommand(t, "config-keys")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	configKeys := string(output)
	assert.Contains(t, configKeys, "cookie-db")
	assert.Contains(t, configKeys, "verbose")
	assert.NotContains(t, configKeys, "jwt")
}

func TestRootEnvVarLoading(t *testing.T) {
	cmd := rootExecCommand(t, "--debug-options=json")
	cmd.Env = append(cmd.Env,
		"MARKETSURGE_AGENT_COOKIE_DB=/tmp/marketsurge-agent-env-cookies.sqlite",
		"MARKETSURGE_AGENT_VERBOSE=true",
	)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	debugOutput := string(output)
	assert.Contains(t, debugOutput, `"value": "/tmp/marketsurge-agent-env-cookies.sqlite"`)
	assert.Contains(t, debugOutput, `"source": "env"`)
	assert.Contains(t, debugOutput, `"verbose": "true"`)
}

func TestRootRejectsJWTFlag(t *testing.T) {
	cmd := rootExecCommand(t, "--jwt", "token", "stock", "get", "AAPL")
	output, err := cmd.CombinedOutput()
	require.Error(t, err, string(output))

	assert.Contains(t, string(output), "unknown flag: --jwt")
}

func TestDisplayVersionFromSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rawVersion string
		settings   []debug.BuildSetting
		want       string
	}{
		{
			name:       "release version ignores vcs revision",
			rawVersion: "1.2.3",
			settings: []debug.BuildSetting{
				{Key: vcsRevisionKey, Value: "09dd191abc123"},
			},
			want: "1.2.3",
		},
		{
			name:       "dev version uses short vcs revision",
			rawVersion: devVersion,
			settings: []debug.BuildSetting{
				{Key: vcsRevisionKey, Value: "09dd191abc123"},
			},
			want: "dev-09dd191",
		},
		{
			name:       "dev version keeps short revision",
			rawVersion: devVersion,
			settings: []debug.BuildSetting{
				{Key: vcsRevisionKey, Value: "abc123"},
			},
			want: "dev-abc123",
		},
		{
			name:       "dev version without vcs revision stays dev",
			rawVersion: devVersion,
			settings: []debug.BuildSetting{
				{Key: "vcs.modified", Value: "true"},
			},
			want: devVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := displayVersionFromSettings(tt.rawVersion, tt.settings)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDisplayVersionFromCommit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rawVersion string
		revision   string
		want       string
	}{
		{
			name:       "release version ignores commit",
			rawVersion: "1.2.3",
			revision:   "09dd191abc123",
			want:       "1.2.3",
		},
		{
			name:       "dev version uses short commit",
			rawVersion: devVersion,
			revision:   "09dd191abc123",
			want:       "dev-09dd191",
		},
		{
			name:       "dev version keeps already short commit",
			rawVersion: devVersion,
			revision:   "abc123",
			want:       "dev-abc123",
		},
		{
			name:       "dev version without commit stays dev",
			rawVersion: devVersion,
			revision:   "",
			want:       devVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := displayVersionFromCommit(tt.rawVersion, tt.revision)
			assert.Equal(t, tt.want, got)
		})
	}
}

func rootExecCommand(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()

	tmpDir := t.TempDir()
	cmdArgs := append([]string{"run", "./marketsurge-agent"}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Env = append(sanitizedRootCommandEnv(),
		"HOME="+tmpDir,
		"GOPATH="+filepath.Join(userHomeDir(t), "go"),
		"XDG_CONFIG_HOME="+tmpDir+"/xdg",
	)
	return cmd
}

func userHomeDir(t *testing.T) string {
	t.Helper()

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	return homeDir
}

func sanitizedRootCommandEnv() []string {
	env := make([]string, 0, len(os.Environ()))
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "MARKETSURGE_AGENT_") ||
			strings.HasPrefix(entry, "HOME=") ||
			strings.HasPrefix(entry, "XDG_CONFIG_HOME=") {
			continue
		}
		env = append(env, entry)
	}
	return env
}

type mcpResponse struct {
	Result json.RawMessage `json:"result"`
	Error  any             `json:"error"`
}

type mcpInitializeResult struct {
	ServerInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

type mcpToolsListResult struct {
	Tools []struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		InputSchema json.RawMessage `json:"inputSchema"`
	} `json:"tools"`
}

func mcpToolNames(result mcpToolsListResult) []string {
	toolNames := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		toolNames = append(toolNames, tool.Name)
	}
	return toolNames
}

func mcpToolByName(t *testing.T, result mcpToolsListResult, name string) struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
} {
	t.Helper()
	for _, tool := range result.Tools {
		if tool.Name == name {
			return tool
		}
	}
	require.Failf(t, "MCP tool not found", "tool %q not found in %v", name, mcpToolNames(result))
	return struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		InputSchema json.RawMessage `json:"inputSchema"`
	}{}
}

func mcpInputSchema(t *testing.T, tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}) map[string]any {
	t.Helper()
	var schema map[string]any
	require.NoError(t, json.Unmarshal(tool.InputSchema, &schema), "tool %s input schema should be JSON", tool.Name)
	return schema
}

func schemaProperties(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()
	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok, "schema should include object properties")
	return props
}

type mcpToolCallResult struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	IsError bool `json:"isError"`
}

func decodeMCPResponses(t *testing.T, output []byte) []mcpResponse {
	t.Helper()

	dec := json.NewDecoder(strings.NewReader(string(output)))
	responses := []mcpResponse{}
	for {
		var response mcpResponse
		err := dec.Decode(&response)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		responses = append(responses, response)
	}
	return responses
}

func parseJSONSchemas(t *testing.T, output []byte) []map[string]any {
	t.Helper()

	var schemas []map[string]any
	require.NoError(t, json.Unmarshal(output, &schemas), string(output))
	require.NotEmpty(t, schemas)
	return schemas
}

func schemaJSONByTitle(t *testing.T, schemas []map[string]any, title string) string {
	t.Helper()

	schema := schemaMapByTitle(t, schemas, title)
	encoded, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err)
	return string(encoded)
}

func schemaMapByTitle(t *testing.T, schemas []map[string]any, title string) map[string]any {
	t.Helper()

	for _, schema := range schemas {
		if schema["title"] == title {
			return schema
		}
	}
	require.Failf(t, "schema title not found", "missing %q", title)
	return nil
}

func schemaProperty(t *testing.T, schema map[string]any, name string) map[string]any {
	t.Helper()

	properties, ok := schema["properties"].(map[string]any)
	require.True(t, ok, "schema should have properties")
	property, ok := properties[name].(map[string]any)
	require.True(t, ok, "schema should have property %q", name)
	return property
}
