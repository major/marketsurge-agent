package cmd

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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
	assert.Equal(t, "dev", initResult.ServerInfo.Version)

	var listResult mcpToolsListResult
	require.NoError(t, json.Unmarshal(responses[1].Result, &listResult))
	toolNames := make([]string, 0, len(listResult.Tools))
	for _, tool := range listResult.Tools {
		toolNames = append(toolNames, tool.Name)
	}
	assert.ElementsMatch(t, []string{
		"catalog-list",
		"catalog-run",
		"chart-history",
		"chart-markups",
		"fundamental-get",
		"ownership-get",
		"rs-history-get",
		"stock-analyze",
		"stock-get",
	}, toolNames)
	assert.NotContains(t, toolNames, "completion-bash")
	assert.NotContains(t, toolNames, "completion-fish")
	assert.NotContains(t, toolNames, "completion-powershell")
	assert.NotContains(t, toolNames, "completion-zsh")
	assert.NotContains(t, string(output), "Firefox")
	assert.NotContains(t, string(output), "cookie")
}

func TestRootMCPToolsCallUsesAPIAuth(t *testing.T) {
	cmd := rootExecCommand(t, "--mcp")
	cmd.Stdin = strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"stock-analyze","arguments":{"tickers":"AAPL"}}}` + "\n")
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
		Name string `json:"name"`
	} `json:"tools"`
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
