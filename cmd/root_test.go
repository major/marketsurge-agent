package cmd_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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

func TestHelpShowsGlobalFlagsAndReports(t *testing.T) {
	t.Parallel()

	cmd := exec.Command(testBinary, "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s --help error = %v, output = %q, want nil error", testBinary, err, string(out))
	}

	output := string(out)
	assertContains(t, output, "reports", "--help")
	assertContains(t, output, "cookie-db", "--help")
	assertContains(t, output, "verbose", "--help")
}

func TestVersionPrintsDev(t *testing.T) {
	t.Parallel()

	cmd := exec.Command(testBinary, "--version")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s --version error = %v, output = %q, want nil error", testBinary, err, string(out))
	}

	if got, want := strings.TrimSpace(string(out)), "dev"; got != want {
		t.Errorf("%s --version = %q, want %q", testBinary, got, want)
	}
}

func TestMissingSubcommandFailsBeforeAuth(t *testing.T) {
	t.Parallel()

	cmd := exec.Command(testBinary)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("%s error = nil, output = %q, want non-zero exit", testBinary, string(out))
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("%s error type = %T, want *exec.ExitError", testBinary, err)
	}
	if got, want := exitErr.ExitCode(), 80; got != want {
		t.Errorf("%s exit code = %d, want %d", testBinary, got, want)
	}
	assertContains(t, string(out), `expected "reports"`, "missing subcommand")
}

func TestAuthErrorWritesJSONAndExits32(t *testing.T) {
	t.Parallel()

	cmd := exec.Command(testBinary, "--cookie-db", "/nonexistent/path/cookies.sqlite", "reports", "list")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatalf("%s --cookie-db /nonexistent/path/cookies.sqlite reports list error = nil, want non-zero exit", testBinary)
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("%s --cookie-db /nonexistent/path/cookies.sqlite reports list error type = %T, want *exec.ExitError", testBinary, err)
	}
	if got, want := exitErr.ExitCode(), 32; got != want {
		t.Errorf("%s --cookie-db /nonexistent/path/cookies.sqlite reports list exit code = %d, want %d", testBinary, got, want)
	}

	var payload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(stderr.String()), &payload); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v, want nil error", stderr.String(), err)
	}
	if got, want := payload.Code, "AUTH_FAILED"; got != want {
		t.Errorf("auth error JSON code = %q, want %q", got, want)
	}
	if payload.Message == "" {
		t.Errorf("auth error JSON message = %q, want non-empty", payload.Message)
	}
}

func assertContains(t *testing.T, output, substring, command string) {
	t.Helper()
	if !strings.Contains(output, substring) {
		t.Errorf("%s output = %q, want substring %q", command, output, substring)
	}
}
