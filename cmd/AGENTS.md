# cmd package

Cobra command tree for marketsurge-agent.

## Structure

- `root.go` - Root command, PersistentPreRunE (auth), PersistentPostRunE (cleanup), Execute()
- `symbol.go` - Shared symbol-fetcher pattern (newSymbolCmd)
- `docs.go` - Doc generation command
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

### Option struct pattern (structcli bridge)

Commands use option structs to encapsulate flags and validation logic. This pattern bridges the gap between Cobra's flag system and the upcoming structcli migration.

**Canonical 4-step pattern:**

1. Constructor creates `opts := &XxxOptions{}` before the command
2. `opts.BindFlags(cmd)` registers flags via `StringVar`, `IntVar`, etc. (called after command creation)
3. RunE calls `opts.FromCommand(cmd)` if pointer fields exist, then `opts.Validate()` if validation exists
4. RunE accesses `opts.Field` instead of `cmd.Flags().GetXxx()`

**Simplest example (ChartMarkupsOptions):**

```go
type ChartMarkupsOptions struct {
    Frequency string
    SortDir   string
}

func (o *ChartMarkupsOptions) BindFlags(cmd *cobra.Command) {
    cmd.Flags().StringVar(&o.Frequency, "frequency", "DAILY", "Chart frequency: DAILY or WEEKLY")
    cmd.Flags().StringVar(&o.SortDir, "sort-dir", "ASC", "Sort direction: ASC or DESC")
}

func newChartMarkupsCmd() *cobra.Command {
    opts := &ChartMarkupsOptions{}
    cmd := &cobra.Command{
        Use:   "markups <symbol>",
        Short: "Get chart markup data for a symbol",
        RunE: func(cmd *cobra.Command, args []string) error {
            symbol := args[0]
            c := ClientFromContext(cmd.Context())
            data, err := c.GetChartMarkups(cmd.Context(), symbol, opts.Frequency, opts.SortDir)
            if err != nil {
                return err
            }
            return output.WriteSuccess(cmd.OutOrStdout(), data, output.SymbolMeta(symbol))
        },
    }
    opts.BindFlags(cmd)
    return cmd
}
```

**Methods by command:**

- `BindFlags()` only: ChartMarkupsOptions, StockAnalyzeOptions, CatalogListOptions, RootOptions
- `BindFlags()` + `Validate()`: ChartHistoryOptions
- `BindFlags()` + `Validate()` + transformation method: ChartHistoryOptions (also has `ResolveDates()`)
- `BindFlags()` + `FromCommand()` + `Validate()` + helper method: CatalogRunOptions (also has `Filters()`)
- `BindFlags()` + transformation method: StockAnalyzeOptions (also has `MergeSymbols()`)

**Pointer fields and FromCommand():**

Only CatalogRunOptions uses pointer fields (`MinComposite *int`, `MinRS *int`) to distinguish "not set" from "set to zero". These require `FromCommand(cmd)` to detect `Changed()` and populate the pointers:

```go
func (o *CatalogRunOptions) FromCommand(cmd *cobra.Command) {
    if cmd.Flags().Changed("min-composite") {
        v, _ := cmd.Flags().GetInt("min-composite")
        o.MinComposite = &v
    }
}
```

This is the only place where `cmd.Flags().GetXxx()` is acceptable in RunE.

**Structcli migration path:**

When migrating to structcli, struct fields become `flag:"name"` tagged, `BindFlags()` is replaced by `structcli.Bind()`, and `Validate()` signature changes to `Validate(ctx context.Context) []error`. The RunE logic remains the same: call `FromCommand()` if needed, then `Validate()`, then use `opts.Field` directly.

## Adding a new command

1. Create `cmd/<group>.go` with constructor function and init()
2. Add client method in `internal/client/<group>.go`
3. Add GraphQL query in `queries/<operation>.graphql`
4. Add model structs in `internal/models/` if needed
5. Add tests in `cmd/<group>_test.go`
6. Update skill files in `skills/marketsurge-agent/`
