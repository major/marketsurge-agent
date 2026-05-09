# marketsurge-agent

Go CLI tool that lets AI agents query the MarketSurge stock research API. Single binary, JSON-first output, and `--help`.

This project is unofficial and is not affiliated with, approved by, or endorsed by MarketSurge or Investor's Business Daily.

Keep `.coderabbit.yaml` and `.github/copilot-instructions.md` plus `.github/instructions/*.instructions.md` aligned with current repo conventions when review-relevant behavior changes.

## Architecture

```text
cmd/
  marketsurge-agent/main.go   Entry point: kong.Parse, auth, ctx.Run(client)
  root.go                     CLI struct (CLI, ReportsCmd), kong flag tags
  reports_list.go             ReportsListCmd.Run(client) - calls Screens API
  reports_get.go              ReportsGetCmd.Run(client) - calls RunScreen API
  root_test.go                Binary-level tests (help, version, auth error, missing subcommand)
  reports_list_test.go        Unit tests for reports list
  reports_get_test.go         Unit tests for reports get
internal/
  auth/                       Cookie-based JWT exchange
  cookies/                    Firefox cookie extraction
  errors/                     Custom error hierarchy + WriteJSON(w, err) + ExitCodeFor(err)
```

### Request flow

1. `main.go` calls `kong.Parse(&cli, ...)` which parses flags and selects the matched command
2. `newClient(cli.CookieDB)` calls `auth.ResolveJWT` then `marketsurge.NewClient(marketsurge.WithJWT(jwt))`
3. On auth failure, `mserrors.WriteJSON(os.Stderr, err)` writes a JSON error and `os.Exit` uses `mserrors.ExitCodeFor(err)`
4. `ctx.Run(client)` dispatches to the matched command's `Run(client *marketsurge.Client) error` method
5. The command calls the marketsurge-go client, marshals the result as JSON to stdout
6. On command error, `mserrors.WriteJSON(os.Stderr, err)` and `os.Exit(mserrors.ExitCodeFor(err))`

### Auth chain (`internal/auth/chain.go`)

Cookie database precedence:

1. `--cookie-db` explicit path to Firefox cookie DB
2. Firefox profile auto-discovery

Cookies are exchanged at `investors.com` using the JWT exchange URL constant, then the resulting JWT is used for API requests at `dowjones.io`. The CLI does not accept JWT injection through arguments, environment variables, or config files.

Auth is wired in `main.go` before `ctx.Run`. Kong's `BindToProvider` pattern is not used; instead, `newClient` runs unconditionally and the client is passed directly to `ctx.Run`.

### Error hierarchy (`internal/errors/errors.go`)

All errors embed `MarketSurgeError` base type. Use the constructor functions, not raw structs.

Exit code ranges: 0 = success, 30+ = domain errors.

| Type | Exit Code | When |
|---|---|---|
| `ValidationError` | 30 | Bad args, missing fields |
| `SymbolNotFoundError` | 31 | Ticker not recognized |
| `AuthenticationError` | 32 | 401/403, missing token |
| `TokenExpiredError` | 32 | 401 specifically |
| `CookieExtractionError` | 32 | Cookie DB read failures |
| `APIError` | 33 | API-level errors |
| `HTTPError` | 33 | 429, 5xx |

Import alias convention: `mserrors` in commands and main.

Error output uses `mserrors.WriteJSON(w, err)` which writes `{"code":"AUTH_FAILED","message":"..."}` to stderr. Success output goes to stdout as a raw JSON array (no envelope wrapper).

### Kong CLI struct (`cmd/root.go`)

The CLI uses [kong](https://github.com/alecthomas/kong) for flag parsing and command dispatch.

```go
type CLI struct {
    CookieDB string           `help:"..." env:"MARKETSURGE_AGENT_COOKIE_DB" name:"cookie-db"`
    Verbose  bool             `help:"..." env:"MARKETSURGE_AGENT_VERBOSE"`
    Version  kong.VersionFlag `help:"..." short:"V"`
    Reports ReportsCmd `cmd:"" help:"..."`
}

type ReportsCmd struct {
    List ReportsListCmd `cmd:"" help:"..."`
    Get  ReportsGetCmd  `cmd:"" help:"..."`
}
```

Key behaviors:
- `--cookie-db` and `--verbose` bind to `MARKETSURGE_AGENT_COOKIE_DB` and `MARKETSURGE_AGENT_VERBOSE` via `env:` tags
- `--version` / `-V` prints the version set via ldflags and exits
- Kong exits with code 80 when a required subcommand is missing
- `[]string` flags use `sep:","` to accept comma-separated values

## Conventions

### Code style

- **Kong struct pattern**: Commands are structs with kong struct tags; no `init()` wiring needed
- **Command dispatch**: Each command struct implements `Run(client *marketsurge.Client) error`; kong calls it via `ctx.Run(client)`
- **Error wrapping**: Always use `fmt.Errorf("context: %w", err)` with `%w`
- **Typed errors**: Use `errors.As()` to match, constructor functions to create
- **Output**: Commands write JSON directly to `os.Stdout` (or an `io.Writer` for testability); no envelope wrapper

### Critical constraints

- Auth runs in `main.go` before `ctx.Run`; there is no per-command auth hook
- `reports get` accepts `--columns` as a comma-separated list with `sep:","` and a default of 23 column names
- `reports list` emits a JSON array of `marketsurge.ScreenEntry` objects; empty result is `[]`, not `null`
- `reports get` reshapes `[][]RunScreenCell` into `[]map[string]any` keyed by `MDItem.Name`

### Adding a new command

1. Add a new struct to `cmd/root.go` (or a new file for larger commands) with kong struct tags
2. Embed the struct in the appropriate parent (`ReportsCmd` or a new group on `CLI`)
3. Implement `Run(client *marketsurge.Client) error` on the struct
4. Write JSON output to `os.Stdout` (or accept `io.Writer` for testability)
5. Add tests in `cmd/<command>_test.go` using the external `cmd_test` package

## Testing

- Framework: Go stdlib `testing` + `testify/assert` + `testify/require`
- CI runs: `go test -v -race -coverprofile=coverage.out ./...`
- Mock pattern: `marketsurge.WithHTTPClient()` + `httptest.NewServer` (no external mock libraries)
- Tests live in the external `cmd_test` package and construct command structs directly
- Table-driven subtests with `t.Run()`, typed error checks with `assert.ErrorAs()`
- Commands that write to `os.Stdout` accept an `io.Writer` parameter in their internal `run` method for output capture in tests

## Build

Requires Go 1.26+.

```bash
make build     # Build binary
make test      # go test -v -race -coverprofile
make smoke     # Local-only live API smoke tests with -tags=smoke
make lint      # golangci-lint run
make clean     # Remove binary + coverage
```

Local smoke tests live in `cmd/smoke_test.go` behind `//go:build smoke` and run with `make smoke`. They require a local Firefox MarketSurge session or `MARKETSURGE_AGENT_COOKIE_DB`.

Linting uses [golangci-lint](https://golangci-lint.run/) v2 with config in `.golangci.yml`. The standard linter set is enabled plus `bodyclose`, `errorlint`, `gocritic`, `misspell`, `modernize`, `nolintlint`, `revive`, `unconvert`, and `unparam`.

CI pipeline: `golangci-lint` (separate job) + `go test -v -race` -> `go build`

Release: push `v*` tag -> goreleaser v2 -> multi-platform binaries (linux/darwin, amd64/arm64, CGO disabled) -> GitHub Releases

## Maintenance

- **Keep this file updated**: When adding, removing, or changing commands, error types, conventions, or architecture, update this file to match.
- **Keep README.md updated**: When changing commands, flags, output format, install instructions, or development workflow, update README.md as well.

## Dependencies

- `github.com/alecthomas/kong` - CLI framework (flag parsing, command dispatch)
- `github.com/major/marketsurge-go/marketsurge` - MarketSurge API client
- `github.com/browserutils/kooky` - Firefox cookie extraction
- `github.com/stretchr/testify` - Test assertions
