package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/major/marketsurge-agent/internal/output"
)

// TestHelpContainsAllCommands verifies that --help exits 0 and lists every
// top-level command group.
func TestHelpContainsAllCommands(t *testing.T) {
	var jsonBuf bytes.Buffer
	app := buildApp(&jsonBuf)

	var helpBuf bytes.Buffer
	app.Writer = &helpBuf

	err := app.Run(context.Background(), []string{"marketsurge-agent", "--help"})
	require.NoError(t, err)

	helpText := helpBuf.String()
	for _, name := range []string{
		"stock", "fundamental", "ownership",
		"rs-history", "chart", "catalog", "skills",
	} {
		assert.Contains(t, helpText, name, "help output should list %q command", name)
	}
}

// TestUnknownCommandReturnsError verifies that an unrecognized subcommand
// produces a non-nil error from Run.
func TestUnknownCommandReturnsError(t *testing.T) {
	var buf bytes.Buffer
	app := buildApp(&buf)
	app.Writer = io.Discard

	err := app.Run(context.Background(), []string{"marketsurge-agent", "nonexistent"})
	assert.Error(t, err)
}

// TestErrorOutputIsValidJSON verifies that error responses are valid JSON
// with the expected envelope structure.
func TestErrorOutputIsValidJSON(t *testing.T) {
	// Force auth failure by clearing all JWT sources.
	t.Setenv("MARKETSURGE_JWT", "")
	t.Setenv("TICKERSCOPE_JWT", "")

	var jsonBuf bytes.Buffer
	app := buildApp(&jsonBuf)
	app.Writer = io.Discard

	// Running a real command without auth triggers an AuthenticationError
	// from the Before handler.
	err := app.Run(context.Background(), []string{"marketsurge-agent", "stock", "get", "AAPL"})
	require.Error(t, err)

	// Simulate main()'s error handler writing JSON to stdout.
	var errBuf bytes.Buffer
	writeErr := output.WriteError(&errBuf, err)
	require.NoError(t, writeErr)

	var envelope output.ErrorEnvelope
	decodeErr := json.NewDecoder(&errBuf).Decode(&envelope)
	require.NoError(t, decodeErr, "error output must be valid JSON")
	assert.NotEmpty(t, envelope.Error.Code)
	assert.NotEmpty(t, envelope.Error.Message)
}
