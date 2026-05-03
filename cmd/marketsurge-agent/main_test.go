package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testBinary holds the path to the compiled CLI binary, built once in TestMain.
var testBinary string

// TestMain builds the binary once before running all tests and removes it after.
func TestMain(m *testing.M) {
	tmp, err := os.CreateTemp("", "marketsurge-agent-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "creating temp file: %v\n", err)
		os.Exit(1)
	}
	tmp.Close()
	testBinary = tmp.Name()

	build := exec.Command("go", "build", "-o", testBinary, ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "building test binary: %v\n", err)
		os.Remove(testBinary)
		os.Exit(1)
	}

	code := m.Run()
	os.Remove(testBinary)
	os.Exit(code)
}

// TestHelpContainsAllCommands verifies that --help lists every command group
// plus the built-in completion subcommand.
func TestHelpContainsAllCommands(t *testing.T) {
	out, err := exec.Command(testBinary, "--help").CombinedOutput()
	require.NoError(t, err, "--help should exit 0: %s", out)

	helpText := string(out)
	for _, name := range []string{
		"stock", "fundamental", "ownership", "rs-history",
		"chart", "catalog",
	} {
		assert.Contains(t, helpText, name, "help should list %q", name)
	}
}

// TestUnknownCommandReturnsError verifies that an unrecognized subcommand
// produces a non-zero exit code and mentions "unknown command".
func TestUnknownCommandReturnsError(t *testing.T) {
	cmd := exec.Command(testBinary, "nonexistent")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err, "unknown command should exit non-zero")
	assert.Contains(t, stdout.String(), "Usage:", "unknown commands still show root help on stdout")

	envelope := parseErrorEnvelope(t, stderr.Bytes())
	assert.Equal(t, "GENERAL_ERROR", envelope["code"])
	assert.Contains(t, strings.ToLower(envelope["message"].(string)), "unknown command")
}

// TestUnknownFlagReturnsJSONError verifies cobra/structcli flag errors use the
// same stderr JSON envelope as domain errors.
func TestUnknownFlagReturnsJSONError(t *testing.T) {
	cmd := exec.Command(testBinary, "--jwt", "not-supported")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err, "unknown flag should exit non-zero")
	assert.Empty(t, stdout.String(), "errors should not be written to stdout")

	envelope := parseErrorEnvelope(t, stderr.Bytes())
	assert.Equal(t, "GENERAL_ERROR", envelope["code"])
	assert.Contains(t, strings.ToLower(envelope["message"].(string)), "unknown flag")
}

// TestErrorOutputIsValidJSON runs a command with no valid Firefox profile
// (HOME=/nonexistent) and confirms the error output is a JSON
// envelope with an "error" key.
func TestErrorOutputIsValidJSON(t *testing.T) {
	cmd := exec.Command(testBinary, "stock", "get", "AAPL")

	// Strip home and structcli configuration inputs so browser-cookie auth fails
	// even on developer machines with local marketsurge-agent settings.
	var env []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "HOME=") ||
			strings.HasPrefix(e, "XDG_CONFIG_HOME=") ||
			strings.HasPrefix(e, "MARKETSURGE_AGENT_") {
			continue
		}
		env = append(env, e)
	}
	env = append(env, "HOME=/nonexistent")
	cmd.Env = env

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err, "should fail with auth error")

	var envelope map[string]any
	require.NoError(t, json.Unmarshal(stderr.Bytes(), &envelope),
		"stderr should be valid JSON: %s", stderr.String())
	assert.Contains(t, envelope, "error")
}

func parseErrorEnvelope(t *testing.T, payload []byte) map[string]any {
	t.Helper()
	var envelope map[string]any
	require.NoError(t, json.Unmarshal(payload, &envelope), "stderr should be valid JSON: %s", string(payload))
	errorBody, ok := envelope["error"].(map[string]any)
	require.True(t, ok, "stderr JSON should include an error object: %s", string(payload))
	return errorBody
}
