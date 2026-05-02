# cmd package

Cobra command tree for marketsurge-agent.

## Structure

- `root.go` - Root command, PersistentPreRunE (auth), PersistentPostRunE (cleanup), Execute()
- `symbol.go` - Shared symbol-fetcher pattern (newSymbolCmd)
- `stock.go` - stock get + stock analyze
- `fundamental.go` - fundamental get
- `ownership.go` - ownership get
- `rs_history.go` - rs-history get
- `chart.go` - chart history + chart markups
- `catalog.go` - catalog list + catalog run
- `helpers_test.go` - Test utilities

## Patterns

### Command constructor pattern

Each command file uses unexported constructors + init() for wiring:

```go
func init() { rootCmd.AddCommand(newFundamentalCmd()) }

func newFundamentalCmd() *cobra.Command {
    cmd := &cobra.Command{Use: "fundamental", Short: "..."}
    cmd.AddCommand(newSymbolCmd("get", "...", func(ctx context.Context, symbol string) (any, error) {
        return ClientFromContext(ctx).GetFundamentals(ctx, symbol)
    }))
    return cmd
}
```

### Context-based client injection

Client is stored in context by PersistentPreRunE, retrieved in RunE:

```go
c := ClientFromContext(cmd.Context())
```

### MCP behavior

The root command enables structcli's stdio MCP server with `structcli.WithMCP()`. MCP discovery requests (`initialize`, `tools/list`) are intercepted before `PersistentPreRunE`, so they must not require Firefox cookie auth. MCP `tools/call` executes the normal Cobra command path and must keep using `PersistentPreRunE` auth for API commands.

Shell completion subcommands are excluded from MCP tool discovery because they are not useful agent-callable API tools. Keep the MCP tool list focused on MarketSurge data commands.

MCP argument conversion maps tool arguments to flags, not positional arguments. Any API command that requires stock symbols must expose those symbols through schema-visible structcli flags (`--symbol` for single-symbol commands, `--symbols` for multi-symbol commands) while preserving positional arguments for shell compatibility.

### JSON schema behavior

The root command enables structcli JSON schema output with `jsonschema.WithFullTree()` and `jsonschema.WithEnumInDescription()`. Bare `--jsonschema` and explicit `--jsonschema=tree` both return the full command tree as a JSON array, which lets LLM agents discover every runnable command in one call.

Enum-backed flags should keep typed enum fields when a non-empty default exists. The schema exposes those values through machine-readable `enum` arrays and keeps the `{value1,value2}` text in descriptions for prompt readability. Optional enum-like fields with no default, such as `ChartHistoryOptions.Lookback`, must stay `string` and document valid values in `flagdescr` so command validation can return domain `ValidationError` envelopes.

### Test pattern

Tests call constructors directly (not package-level vars) and inject client via context:

```go
func TestFundamentalGet(t *testing.T) {
    t.Parallel()
    server := jsonServer(fundamentalResponseFixture())
    defer server.Close()
    c := testClient(t, server)
    output, err := executeCommandWithClient(newFundamentalCmd(), c, "get", "AAPL")
    require.NoError(t, err)
    result := parseJSONEnvelope(t, output)
    assertSymbolMeta(t, result, "AAPL")
}
```

### Option struct pattern (structcli)

Commands use option structs with structcli struct tags to encapsulate flags and validation logic.

**Canonical 4-step pattern:**

1. Constructor creates `opts := &XxxOptions{}` before the command
2. `structcli.Bind(cmd, opts)` registers flags from struct tags (called after command creation)
3. RunE calls `opts.FromCommand(cmd)` if pointer fields exist, then `opts.Validate(ctx)` if validation exists
4. RunE accesses `opts.Field` directly (structcli populates fields before RunE runs)

**Simplest example (ChartMarkupsOptions):**

```go
type ChartMarkupsOptions struct {
    Symbol    string               `flag:"symbol" flagshort:"s" flaggroup:"Input" flagdescr:"Stock symbol to fetch"`
    Frequency models.Frequency    `flag:"frequency" default:"DAILY" flagdescr:"Chart frequency"`
    SortDir   models.SortDirection `flag:"sort-dir" default:"ASC" flagdescr:"Sort direction"`
}

func newChartMarkupsCmd() *cobra.Command {
    opts := &ChartMarkupsOptions{}
    cmd := &cobra.Command{
        Use:   "markups <symbol>",
        Short: "Get chart markup data for a symbol",
        Args:  cobra.ArbitraryArgs,
        RunE: func(cmd *cobra.Command, args []string) error {
            symbol, err := resolveSingleSymbol(args, opts.Symbol)
            if err != nil {
                return err
            }
            c := ClientFromContext(cmd.Context())
            data, err := c.GetChartMarkups(cmd.Context(), symbol, string(opts.Frequency), string(opts.SortDir))
            if err != nil {
                return err
            }
            return output.WriteSuccess(cmd.OutOrStdout(), data, output.SymbolMeta(symbol))
        },
    }
    structcli.Bind(cmd, opts)
    return cmd
}
```

**Methods by command:**

- `structcli.Bind()` only: ChartMarkupsOptions, StockAnalyzeOptions, CatalogListOptions, RootOptions
- `structcli.Bind()` + `Validate(ctx) []error`: ChartHistoryOptions
- `structcli.Bind()` + `Validate(ctx) []error` + transformation method: ChartHistoryOptions (also has `ResolveDates()`)
- `structcli.Bind()` + `FromCommand()` + `Validate(ctx) []error` + helper method: CatalogRunOptions (also has `Filters()`)
- `structcli.Bind()` + transformation method: StockAnalyzeOptions (also has `MergeSymbols()`)

**Pointer fields and FromCommand():**

Only CatalogRunOptions uses pointer fields (`MinComposite *int`, `MinRS *int`) to distinguish "not set" from "set to zero". structcli cannot handle `*int` natively, so these flags are registered manually after `structcli.Bind()` and `FromCommand(cmd)` detects `Changed()` to populate the pointers:

```go
func (o *CatalogRunOptions) FromCommand(cmd *cobra.Command) {
    if cmd.Flags().Changed("min-composite") {
        v, _ := cmd.Flags().GetInt("min-composite")
        o.MinComposite = &v
    }
}
```

This is the only place where `cmd.Flags().GetXxx()` is acceptable in RunE.

**Optional enum fields:**

Fields with no `default:` struct tag must stay `string` type. structcli rejects an empty string for a registered enum during unmarshal, before `Validate()` can return a typed `ValidationError`. Only use typed enum fields when a non-empty default is always present (e.g., `Frequency`, `SortDir`, `Period`). Fields like `Lookback` stay `string` with manual validation.

### Schema-visible symbol flags

Use `newSymbolCmd()` for single-symbol commands. It exposes `--symbol/-s` through structcli and still accepts one positional `<symbol>` for shell users. Commands with their own option structs, such as `chart history` and `chart markups`, add a `Symbol string` field with the same `flag:"symbol"` tags and call `resolveSingleSymbol(args, opts.Symbol)`.

Use `--symbols/-s` for multi-symbol commands. Merge positional symbols and flag values with `mergeSymbolInputs()` so repeated flags, comma-separated values, and positional arguments deduplicate consistently.

## Adding a new command

1. Create `cmd/<group>.go` with constructor function and init()
2. Define an options struct with `flag:`, `flagdescr:`, `flaggroup:`, and `default:` struct tags
3. Call `structcli.Bind(cmd, opts)` in the constructor (replaces manual flag registration)
4. Add client method in `internal/client/<group>.go`
5. Add GraphQL query in `queries/<operation>.graphql`
6. Add model structs in `internal/models/` if needed
7. Add tests in `cmd/<group>_test.go`

Use `flaggroup:` on every non-trivial flag so `--jsonschema=tree` and MCP tool metadata are useful to LLM agents. Prefer `Validate()` over `flagrequired:"true"` for conditional requirements or domain errors that must preserve the JSON error envelope and MarketSurge exit codes.

For complex modes, put copyable examples where schema generators can see them:

- Put complete invocation examples in command `Long` descriptions for schema output; Cobra `Example` is useful for help/generation consumers but is not emitted by structcli's draft JSON Schema conversion.
- Include short examples in `flagdescr` for conditional flags, for example date range pairs and `catalog run --kind` plus matching ID flag combinations.
- Do not use presets just to document examples. Presets create real CLI alias flags, so reserve them for deliberate UX changes.
