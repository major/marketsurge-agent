package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
		"chart", "catalog", "completion",
	} {
		assert.Contains(t, helpText, name, "help should list %q", name)
	}
}

// TestStaticSkillDocsStayWellFormed checks that all 8 expected skill files
// exist, are non-empty, have at least one heading, and close all code fences.
func TestStaticSkillDocsStayWellFormed(t *testing.T) {
	skillDir := filepath.Join("..", "..", "skills", "marketsurge-agent")
	expected := []string{
		"SKILL.md", "index.md", "stock.md", "fundamental.md",
		"ownership.md", "rs-history.md", "chart.md", "catalog.md",
	}

	entries, err := os.ReadDir(skillDir)
	require.NoError(t, err)

	var found []string
	for _, e := range entries {
		if !e.IsDir() {
			found = append(found, e.Name())
		}
	}
	assert.ElementsMatch(t, expected, found, "exactly 8 skill files expected")

	for _, name := range expected {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(skillDir, name))
			require.NoError(t, err)
			content := string(data)

			assert.NotEmpty(t, content)
			assert.Contains(t, content, "## ")

			fences := strings.Count(content, "```")
			assert.Equal(t, 0, fences%2,
				"unclosed code fence (count=%d)", fences)
		})
	}
}

// TestUnknownCommandReturnsError verifies that an unrecognized subcommand
// produces a non-zero exit code and mentions "unknown command".
func TestUnknownCommandReturnsError(t *testing.T) {
	out, err := exec.Command(testBinary, "nonexistent").CombinedOutput()
	require.Error(t, err, "unknown command should exit non-zero")
	assert.Contains(t, strings.ToLower(string(out)), "unknown command")
}

// TestErrorOutputIsValidJSON runs a command with no valid auth
// (HOME=/nonexistent, no JWT) and confirms the error output is a JSON
// envelope with an "error" key.
func TestErrorOutputIsValidJSON(t *testing.T) {
	cmd := exec.Command(testBinary, "stock", "get", "AAPL")

	// Strip HOME and MARKETSURGE_JWT so the auth chain fails entirely.
	var env []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "HOME=") ||
			strings.HasPrefix(e, "MARKETSURGE_JWT=") {
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
