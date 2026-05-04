# marketsurge-agent

Go CLI tool that lets AI agents query the MarketSurge stock research API. Single binary, JSON-first output, self-documenting via `--jsonschema`, generated `SKILL.md`, and `--help`.

This project is unofficial and is not affiliated with, approved by, or endorsed by MarketSurge or Investor's Business Daily.

## Architecture

```text
cmd/
  generate-docs/main.go          Generates root SKILL.md from the command tree
  marketsurge-agent/main.go      Entry point, calls cmd.Execute()
  root.go                        Root command, PersistentPreRunE (auth), Execute()
  symbol.go                      Shared symbol-fetcher pattern
  <group>.go                     One file per command group (stock, chart, etc.)
internal/
  auth/                          Cookie-based JWT exchange
  client/                        GraphQL client + domain methods
  constants/                     API endpoints, columns, report IDs
  cookies/                       Firefox cookie extraction
  errors/                        Custom error hierarchy
  models/                        Data structures (includes enums.go)
  output/                        JSON envelope formatting
queries/                         Embedded .graphql files (go:embed)
```

### Request flow

1. `main.go` calls `cmd.Execute()` which runs the root cobra command
2. `PersistentPreRunE` exchanges Firefox cookies for a JWT, injects `client.Client` into context
3. Command `RunE` retrieves client via `ClientFromContext(cmd.Context())`, validates args, calls client method
4. Client loads embedded `.graphql` query, executes HTTP POST to GraphQL endpoint
5. Response parsed into typed model, wrapped in JSON envelope via `output.WriteSuccess`
6. `PersistentPostRunE` closes the client

### Auth chain (`internal/auth/chain.go`)

Cookie database precedence:

1. `--cookie-db` explicit path to Firefox cookie DB
2. Firefox profile auto-discovery

Cookies are exchanged at `investors.com` using the DylanToken constant, then the resulting JWT is used for GraphQL requests at `dowjones.io`. The CLI does not accept JWT injection through arguments, environment variables, or config files.

### Error hierarchy (`internal/errors/errors.go`)

All errors embed `MarketSurgeError` base type. Use the constructor functions, not raw structs.

Exit code ranges: 0 = success, 10-23 = reserved by structcli, 30+ = domain errors.

| Type | Exit Code | When |
|---|---|---|
| `ValidationError` | 30 | Bad args, missing fields |
| `SymbolNotFoundError` | 31 | Ticker not recognized |
| `AuthenticationError` | 32 | 401/403, missing token |
| `TokenExpiredError` | 32 | 401 specifically |
| `CookieExtractionError` | 32 | Cookie DB read failures |
| `APIError` | 33 | GraphQL-level errors |
| `HTTPError` | 33 | 429, 5xx |

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

Envelope shape: `{ data, metadata }` for success, `{ data, errors, metadata }` for partial success, and `{ error: { code, message, details } }` for errors.

### structcli integration (`cmd/root.go`)

The CLI uses [structcli](https://github.com/leodido/structcli) v0.17.0 for flag binding, validation, and self-documentation.

Root setup in `init()`:

```go
structcli.Bind(rootCmd, rootOpts)
structcli.Setup(rootCmd,
    structcli.WithAppName("marketsurge-agent"),
    structcli.WithConfig(config.Options{ValidateKeys: true}),
    structcli.WithJSONSchema(jsonschema.Options{
        SchemaOpts: []jsonschema.Opt{
            jsonschema.WithFullTree(),
            jsonschema.WithEnumInDescription(),
        },
    }),
    structcli.WithFlagErrors(),
    structcli.WithHelpTopics(helptopics.Options{ReferenceSection: true}),
    structcli.WithDebug(debug.Options{Exit: true}),
    structcli.WithMCP(mcp.Options{
        Name:      "marketsurge-agent",
        Version:   version,
        Separator: "_",
        Exclude: []string{
            "marketsurge-agent completion bash",
            "marketsurge-agent completion fish",
            "marketsurge-agent completion powershell",
            "marketsurge-agent completion zsh",
        },
    }),
)
```

Key behaviors:
- `WithAppName("marketsurge-agent")` sets the structcli env prefix to `MARKETSURGE_AGENT` and propagates the app name to config/debug helpers
- `WithConfig()` adds the persistent `--config` flag, auto-loads config before structcli unmarshals options, and honors `MARKETSURGE_AGENT_CONFIG`
- `--jsonschema` prints the full command tree JSON schema and exits without auth
- `--jsonschema=tree` remains supported and returns the same full-tree schema for scripts that already request it
- `--mcp` runs a stdio Model Context Protocol server; `initialize` and `tools/list` discovery do not require auth, while `tools/call` for API commands uses the normal Firefox cookie auth chain; shell completion subcommands are excluded from the MCP tool list
- MCP tool names use `_` between command path segments, for example `stock_analyze`, `chart_history`, and `catalog_run`; command segments that already contain a dash keep it, for example `rs-history_get`
- MCP exposes only runnable leaf API commands. Do not set `AllCommands: true` unless parent command tools become intentionally useful.
- `--debug-options` prints resolved flag values and exits (requires `Exit: true` in debug options)
- `env-vars` and `config-keys` are built-in help topics (e.g., `marketsurge-agent help env-vars`)
- `structcli.ExecuteC(rootCmd)` replaces `rootCmd.Execute()` in `Execute()`
- `rootCmd.TraverseChildren = true` is required for root-bound flags to work on subcommands

Root options use structcli-managed configuration:
- `--cookie-db` and `--verbose` use `flagenv:"true"`, so structcli binds them to `MARKETSURGE_AGENT_COOKIE_DB` and `MARKETSURGE_AGENT_VERBOSE`
- Config files can set non-secret root keys such as `cookie-db` and `verbose`

`isNonAPICommand()` skips auth for: `completion`, `help`, `env-vars`, `config-keys`. structcli intercepts `--mcp` before normal hooks for discovery requests, so MCP discovery also skips auth.

### Typed enums (`internal/models/enums.go`)

Typed string enums are registered with structcli so flag validation and schema generation know the allowed values.

| Type | Values | Used by |
|---|---|---|
| `Frequency` | `DAILY`, `WEEKLY` | chart markups, chart history |
| `SortDirection` | `ASC`, `DESC` | chart markups |
| `Lookback` | `1W`, `1M`, `3M`, `6M`, `1Y`, `YTD` | chart history |
| `Period` | `daily`, `weekly` | chart history |
| `CatalogKind` | `watchlist`, `screen`, `report`, `coach_screen` | catalog run |

`CatalogKind` is defined in `catalog.go` but registered in `enums.go`'s `init()`.

**Optional enum fields**: fields with no `default:` struct tag must stay `string` type. structcli rejects an empty string for a registered enum during unmarshal, before `Validate()` can return a typed `ValidationError`. Only use typed enum fields when a non-empty default is always present.

### Schema fidelity for LLM agents

LLM agents should use the full command tree schema:

```bash
marketsurge-agent --jsonschema=tree
```

Bare `--jsonschema` returns the same full-tree JSON array by default. The explicit `=tree` form is kept for compatibility with scripts and prompts that already request it.

Schema tags should make command selection and flag filling obvious:
- Use `flaggroup:` for logical groups such as `Date Range`, `Output Format`, `Pagination`, and `Filtering & Projection`
- Use `flagdescr:` to document conditional requirements and concrete examples, especially when a flag is only required for one mode
- Registered enum values appear in both machine-readable `enum` arrays and the preserved `{value1,value2}` text in descriptions because the root setup enables `jsonschema.WithEnumInDescription()`
- Keep complete invocation examples in complex command `Long` descriptions and per-flag examples in `flagdescr`, because structcli's JSON Schema output carries those descriptions
- Keep conditional and domain-specific requirements in `Validate()` when they need the CLI's JSON error envelope and MarketSurge exit codes
- Do not mark chart history date fields or catalog ID fields with `flagrequired`, because their requirements depend on other flags

The schema should describe enough for agents to choose flags without scraping prose help, while runtime validation remains the source of truth for mutually exclusive and conditional rules.

### Generated skill file

Run `make docs` after command, flag, default, example, or help text changes. It refreshes the repository-root `SKILL.md` from the live Cobra and structcli command tree so Claude Code and other skill-aware tools have the same primary skill-file location as the sibling agent repositories.

## Conventions

### Code style

- **`init()` for cobra wiring**: Each command file uses `func init() { rootCmd.AddCommand(newXxxCmd()) }` to register commands
- **Command constructors**: `newXxxCmd() *cobra.Command` (unexported, called by init and tests)
- **Error wrapping**: Always use `fmt.Errorf("context: %w", err)` with `%w`
- **Typed errors**: Use `errors.As()` to match, constructor functions to create
- **Concurrency**: `sync.WaitGroup` + `sync.Mutex` for parallel ops (see `stock_analyze.go`)
- **GraphQL queries**: Embedded via `queries/embed.go`, loaded with `queries.Load("name")`

### Critical constraints

- JWT and Cookie HTTP headers must be added per-request in `client.Execute()`, not in base/default headers
- Chart history has mutually exclusive date params: explicit start/end dates XOR lookback period
- `kind` is required for catalog commands; each kind requires its own ID flag (report-id, watchlist-id, coach-screen-id)
- `structcli.Bind(cmd, opts)` sets a scope context on the command. This blocks cobra's parent-to-child context propagation. Test helpers must layer the client onto each subcommand's context explicitly (see `executeStockAnalyze` in `stock_test.go` for the pattern).

### Adding a new command

1. Create `cmd/<group>.go` with constructor function and `init()` wiring
2. Define an options struct with `flag:`, `flagdescr:`, and `default:` struct tags
3. Call `structcli.Bind(cmd, opts)` in the constructor (replaces manual `BindFlags`)
4. Add client method in `internal/client/<group>.go`
5. Add GraphQL query in `queries/<operation>.graphql`
6. Add model structs in `internal/models/` if needed
7. Add tests in `cmd/<group>_test.go`

Follow `fundamental.go` (35 lines) as the canonical simple command template (no options struct needed for symbol-only commands using `newSymbolCmd`).

## Testing

- Framework: Go stdlib `testing` + `testify/assert` + `testify/require`
- CI runs: `go test -v -race -coverprofile=coverage.out ./...`
- Mock pattern: `httptest.NewServer` with request capture (no external mock libraries)
- Shared helpers in `cmd/helpers_test.go`: `testClient()`, `jsonServer()`, fixture builders
- Table-driven subtests with `t.Run()`, typed error checks with `assert.ErrorAs()`
- CLI tests call constructors directly, inject client via context, capture output to `bytes.Buffer`
- `viper.Reset()` called before and after each test execution (structcli uses viper internally)
- `commandExecutionMu sync.Mutex` in helpers serializes command executions to prevent viper races under `t.Parallel()`

## Build

Requires Go 1.26+.

```bash
make build     # Build binary
make test      # go test -v -race -coverprofile
make smoke     # Local-only live API smoke tests with -tags=smoke
make lint      # golangci-lint run
make docs      # Refresh root SKILL.md
make clean     # Remove binary + coverage
```

Local smoke tests live in `cmd/smoke_test.go` behind `//go:build smoke` and run with `make smoke`. They execute curated live invocations for every API leaf command discovered by `--jsonschema`. They require a local Firefox MarketSurge session or `MARKETSURGE_AGENT_COOKIE_DB`; `catalog run` is included only when `MARKETSURGE_SMOKE_WATCHLIST_ID`, `MARKETSURGE_SMOKE_REPORT_ID`, or `MARKETSURGE_SMOKE_COACH_SCREEN_ID` is set.

Linting uses [golangci-lint](https://golangci-lint.run/) v2 with config in `.golangci.yml`. The standard linter set is enabled plus `bodyclose`, `errorlint`, `gocritic`, `misspell`, `modernize`, `nolintlint`, `revive`, `unconvert`, and `unparam`.

CI pipeline: `golangci-lint` (separate job) + `go test -v -race` -> `go build`

Release: push `v*` tag -> goreleaser v2 -> multi-platform binaries (linux/darwin, amd64/arm64, CGO disabled) -> GitHub Releases

## Maintenance

- **Keep this file updated**: When adding, removing, or changing commands, error types, conventions, or architecture, update this file and subdirectory AGENTS.md files to match.
- **Keep README.md updated**: When changing commands, flags, output format, install instructions, or development workflow, update README.md as well.
- **Keep SKILL.md updated**: Run `make docs` when command metadata changes so the generated root skill stays in sync.

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/leodido/structcli` - Flag binding, validation, self-documentation
- `github.com/browserutils/kooky` - Firefox cookie extraction
- `github.com/stretchr/testify` - Test assertions
