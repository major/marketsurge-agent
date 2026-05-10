package cmd_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "marketsurge-agent-test-*")
	if err != nil {
		panic(err)
	}

	testBinary = filepath.Join(dir, "marketsurge-agent")
	out, err := exec.Command("go", "build", "-o", testBinary, "./marketsurge-agent").CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(dir)
		panic("go build ./marketsurge-agent failed: " + string(out))
	}

	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

func TestHelpShowsGlobalFlagsAndCommands(t *testing.T) {
	t.Parallel()

	cmd := exec.Command(testBinary, "--help")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "--help should succeed")

	output := string(out)
	assert.Contains(t, output, "compare")
	assert.Contains(t, output, "overview")
	assert.Contains(t, output, "reports")
	assert.Contains(t, output, "cookie-db")
	assert.Contains(t, output, "verbose")
}

func TestVersionPrintsDev(t *testing.T) {
	t.Parallel()

	cmd := exec.Command(testBinary, "--version")
	out, err := cmd.Output()
	require.NoError(t, err, "--version should succeed")

	assert.Equal(t, "dev", strings.TrimSpace(string(out)))
}

func TestMissingSubcommandFailsBeforeAuth(t *testing.T) {
	t.Parallel()

	cmd := exec.Command(testBinary)
	out, err := cmd.CombinedOutput()
	require.Error(t, err, "missing subcommand should fail")

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr, "error should be *exec.ExitError")
	assert.Equal(t, 80, exitErr.ExitCode(), "kong missing-subcommand exit code")
	assert.Contains(t, string(out), `expected one of "chart", "coach", "compare", "industry", "overview", ...`)
}

func TestAuthErrorWritesJSONAndExits32(t *testing.T) {
	t.Parallel()

	cmd := exec.Command(testBinary, "--cookie-db", "/nonexistent/path/cookies.sqlite", "reports", "list")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	require.Error(t, err, "nonexistent cookie-db should fail")

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr, "error should be *exec.ExitError")
	assert.Equal(t, 32, exitErr.ExitCode(), "auth error exit code")

	var payload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(stderr.String()), &payload), "stderr should be valid JSON")
	assert.Equal(t, "AUTH_FAILED", payload.Code)
	assert.NotEmpty(t, payload.Message, "auth error message should not be empty")
}
