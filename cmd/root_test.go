package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootJSONSchema(t *testing.T) {
	cmd := rootExecCommand(t, "--jsonschema")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	schema := string(output)
	assert.True(t, strings.Contains(schema, `"$schema"`) || strings.Contains(schema, `"properties"`))
	assert.Contains(t, schema, `"x-structcli-env-prefix": "MARKETSURGE_AGENT"`)
	assert.Contains(t, schema, `"MARKETSURGE_AGENT_COOKIE_DB"`)
	assert.Contains(t, schema, `"MARKETSURGE_AGENT_VERBOSE"`)
	assert.Contains(t, schema, `"x-structcli-config-flag": "config"`)
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
	assert.Contains(t, help, "Sign into MarketSurge in Firefox")
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
