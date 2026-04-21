# marketsurge-agent

Go CLI tool that lets AI agents query the MarketSurge stock research API. Single binary, JSON-first output, auto-generated skill files for agent consumption.

This project is unofficial and is not affiliated with, approved by, or endorsed by MarketSurge or Investor's Business Daily.

## Architecture

```text
cmd/marketsurge-agent/main.go    Entry point, buildApp() factory
internal/
  auth/                          JWT resolution (4-tier chain)
  client/                        GraphQL client + domain methods
  commands/                      CLI command implementations
  constants/                     API endpoints, columns, report IDs
  cookies/                       Firefox cookie extraction
  errors/                        Custom error hierarchy
  models/                        Data structures
  output/                        JSON envelope formatting
queries/                         Embedded .graphql files (go:embed)
skills/                          Auto-generated agent skill docs
```

### Request flow

1. `main.go` builds the CLI app with `buildApp()`
2. `Before` hook resolves JWT via the auth chain (skipped for `skills` command)
3. Command handler validates args, calls `client.Client` method
4. Client loads embedded `.graphql` query, executes HTTP POST to GraphQL endpoint
5. Response parsed into typed model, wrapped in JSON envelope via `output.WriteSuccess`

### Auth chain (`internal/auth/chain.go`)

Four-tier JWT precedence, first non-empty wins:

1. `--token` CLI flag
2. `MARKETSURGE_TOKEN` env var
3. `--cookie-db` explicit path to Firefox cookie DB
4. Firefox profile auto-discovery

The JWT is exchanged at `investors.com` using the DylanToken constant, then used for GraphQL requests at `dowjones.io`.

### Error hierarchy (`internal/errors/errors.go`)

All errors embed `MarketSurgeError` base type. Use the constructor functions, not raw structs.

| Type | Exit Code | When |
|---|---|---|
| `AuthenticationError` | 3 | 401/403, missing token |
| `TokenExpiredError` | 3 | 401 specifically |
| `CookieExtractionError` | 3 | Cookie DB read failures |
| `APIError` | 4 | GraphQL-level errors |
| `SymbolNotFoundError` | 2 | Ticker not recognized |
| `HTTPError` | 4 | 429, 5xx |
| `ValidationError` | 1 | Bad args, missing fields |

Import alias convention: `mserrors` in commands, `mserr` in the output package.

### Output contract (`internal/output/`)

Every command must produce JSON envelopes. Never write raw output to stdout.

```go
// Success
output.WriteSuccess(w, data, output.SymbolMeta(symbol))

// Error
output.WriteError(w, err)

// Partial (some symbols succeeded, some failed)
output.WritePartial(w, results, errors, metadata)
```

Envelope shape: `{ data, metadata, timestamp }` for success, `{ error, code, message, timestamp }` for errors.

## Conventions

### Code style

- **No `init()` functions**: All setup in `main()` and the `Before` hook
- **Command factories**: `FooCommand(c *client.Client, w io.Writer) *cli.Command`
- **Error wrapping**: Always use `fmt.Errorf("context: %w", err)` with `%w`
- **Typed errors**: Use `errors.As()` to match, constructor functions to create
- **Concurrency**: `sync.WaitGroup` + `sync.Mutex` for parallel ops (see `stock_analyze.go`)
- **GraphQL queries**: Embedded via `queries/embed.go`, loaded with `queries.Load("name")`

### Critical constraints

- JWT and Cookie HTTP headers must be added per-request in `client.Execute()`, not in base/default headers
- Chart history has mutually exclusive date params: explicit start/end dates XOR lookback period
- `kind` is required for catalog commands; each kind requires its own ID flag (report-id, watchlist-id, coach-screen-id)

### Adding a new command

1. Create `internal/commands/<group>_<action>.go` with factory function
2. Add client method in `internal/client/<group>.go`
3. Add GraphQL query in `queries/<operation>.graphql`
4. Add model structs in `internal/models/` if needed
5. Register in `main.go` `buildApp()` under the appropriate command group
6. Add tests in `internal/commands/<group>_<action>_test.go`

Follow `fundamental_get.go` (33 lines) as the canonical simple command template.

## Testing

- Framework: Go stdlib `testing` + `testify/assert` + `testify/require`
- CI runs: `go test -v -race -coverprofile=coverage.out ./...`
- Mock pattern: `httptest.NewServer` with request capture (no external mock libraries)
- Shared helpers in `internal/commands/helpers_test.go`: `testClient()`, `jsonServer()`, fixture builders
- Table-driven subtests with `t.Run()`, typed error checks with `assert.ErrorAs()`
- CLI tests suppress exit via `ExitErrHandler`, capture output to `bytes.Buffer`

## Build

```bash
make build     # Build binary
make test      # go test -v -race -coverprofile
make lint      # golangci-lint run
make clean     # Remove binary + coverage
```

Linting uses [golangci-lint](https://golangci-lint.run/) v2 with config in `.golangci.yml`. The standard linter set is enabled plus `bodyclose`, `errorlint`, `gocritic`, `misspell`, `nolintlint`, `revive`, `unconvert`, and `unparam`.

CI pipeline: `golangci-lint` (separate job) + `go test -v -race` -> `go build`

Release: push `v*` tag -> goreleaser v2 -> multi-platform binaries (linux/darwin, amd64/arm64, CGO disabled) -> GitHub Releases

## Maintenance

- **Keep this file updated**: When adding, removing, or changing commands, error types, conventions, or architecture, update this file and subdirectory AGENTS.md files to match.
- **Keep README.md updated**: When changing commands, flags, output format, install instructions, or development workflow, update README.md as well.

## Dependencies

- `github.com/nicholasgasior/urfave-cli-v3-docs-markdown` - CLI doc generation
- `github.com/urfave/cli/v3` - CLI framework
- `github.com/nicholasgasior/gcs` - cookie/session handling
- `github.com/browserutils/kooky` - Firefox cookie extraction
- `github.com/stretchr/testify` - Test assertions
