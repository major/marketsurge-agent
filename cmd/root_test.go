package cmd

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootJSONSchema(t *testing.T) {
	cmd := exec.Command("go", "run", "./marketsurge-agent", "--jsonschema")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	schema := string(output)
	assert.True(t, strings.Contains(schema, `"$schema"`) || strings.Contains(schema, `"properties"`))
}

func TestRootDebugOptions(t *testing.T) {
	cmd := exec.Command("go", "run", "./marketsurge-agent", "--debug-options")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}
