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

## Adding a new command

1. Create `cmd/<group>.go` with constructor function and init()
2. Add client method in `internal/client/<group>.go`
3. Add GraphQL query in `queries/<operation>.graphql`
4. Add model structs in `internal/models/` if needed
5. Add tests in `cmd/<group>_test.go`
6. Update skill files in `skills/marketsurge-agent/`
