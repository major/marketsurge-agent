# internal/client

GraphQL client and domain-specific API methods.

## Core: `client.go`

```go
type Client struct {
    JWT        string
    Endpoint   string
    HTTPClient *http.Client
}

// Functional options pattern:
NewClient(jwt string, opts ...Option)
WithBaseURL(url string)       // Override GraphQL endpoint
WithHTTPClient(c *http.Client) // Override HTTP client
```

`Execute(query string, variables map[string]interface{})` handles:

1. Marshal request body (`query` + `variables`)
2. Build HTTP POST request to `Endpoint`
3. Add headers: Content-Type, User-Agent, JWT (Authorization), Cookie
4. Execute request, read response
5. Unmarshal into `graphqlResponse`
6. Check for GraphQL-level errors via `mapGraphQLError()`
7. Return `response.Data`

**Critical**: JWT and Cookie headers are added per-request in `Execute()`. Never set them as default headers on the HTTP client.

## API methods

| File | Method | GraphQL Query | Returns |
|---|---|---|---|
| `stock.go` | `GetStock(symbol)` | `other_market_data.graphql` | `models.StockData` |
| `fundamental.go` | `GetFundamental(symbol)` | `fundamental.graphql` | `models.FundamentalData` |
| `ownership.go` | `GetOwnership(symbol)` | `ownership.graphql` | `models.OwnershipData` |
| `rs_history.go` | `GetRSHistory(symbol)` | `rs_history.graphql` | `models.RSHistoryData` |
| `chart.go` | `GetChartHistory(symbol, opts)` | `chart_history_daily/weekly.graphql` | `models.ChartHistoryData` |
| `chart.go` | `GetChartMarkups(symbol)` | `chart_markups.graphql` | `models.ChartMarkupData` |
| `catalog.go` | `ListCatalog()` | Multiple queries | `models.CatalogData` |
| `catalog.go` | `RunReport(id)` | `report_run.graphql` | `[]map[string]interface{}` |
| `catalog.go` | `RunWatchlist(id, fields)` | `watchlist_run.graphql` | `[]map[string]interface{}` |
| `catalog.go` | `RunCoachScreen(id)` | `coach_screen_run.graphql` | `[]map[string]interface{}` |

## Response parsing

The MarketSurge GraphQL API returns deeply nested, untyped JSON. The client navigates this with helper functions in `graphql_error.go`:

### Navigation helpers

```go
getNestedMap(data, "key1", "key2", ...)  // Traverse nested maps
getNestedSlice(data, "key1", "key2", ...) // Get a slice from nested path
firstMap(slice)                            // First element of []interface{} as map
firstMarketData(data)                      // Shortcut for common response shape
```

### Type-safe converters

```go
stringPtr(v)       // interface{} -> *string
intPtr(v)          // interface{} -> *int
int64Ptr(v)        // interface{} -> *int64
floatPtr(v)        // interface{} -> *float64
boolPtr(v)         // interface{} -> *bool
formattedValue(v)  // interface{} -> string (handles nested {value, format})
scalarValue(v)     // interface{} -> interface{} (unwraps {value} wrappers)
stringify(v)        // interface{} -> string (any type to string)
```

These return nil/zero on type mismatch rather than panicking. Always use them when extracting fields from GraphQL responses.

## Error mapping

### HTTP errors (`client.go: mapHTTPError`)

| Status | Error Type |
|---|---|
| 401 | `TokenExpiredError` |
| 403 | `AuthenticationError` |
| 429 | `HTTPError` (rate limited) |
| 5xx | `HTTPError` (server error) |

### GraphQL errors (`graphql_error.go: mapGraphQLError`)

Inspects the `errors` array in GraphQL responses. Maps error messages containing specific strings (e.g., "not found", "unauthorized") to typed errors.

## Adding a new API method

1. Create the `.graphql` query in `queries/`
2. Add the method to the appropriate file (or create a new one for a new domain)
3. Load the query: `query, err := queries.Load("operation_name")`
4. Call `c.Execute(query, variables)`
5. Navigate the response with `getNestedMap`/`firstMarketData` helpers
6. Map fields to model structs using type-safe converters
7. Return the typed model

### Template: simple method

```go
func (c *Client) GetFoo(symbol string) (*models.FooData, error) {
    query, err := queries.Load("foo")
    if err != nil {
        return nil, fmt.Errorf("loading query: %w", err)
    }

    variables := map[string]interface{}{
        "symbol": symbol,
    }

    data, err := c.Execute(query, variables)
    if err != nil {
        return nil, err
    }

    // Navigate response
    marketData, err := firstMarketData(data)
    if err != nil {
        return nil, err
    }

    // Build model with type-safe converters
    result := &models.FooData{
        Symbol: symbol,
        Name:   stringPtr(marketData["name"]),
    }

    return result, nil
}
```

## `catalog.go` specifics

The most complex client file (392 lines):

- `ListCatalog()` aggregates 4 separate GraphQL queries into a unified catalog
- `RunReport`, `RunWatchlist`, `RunCoachScreen` each use different GraphQL operations and response shapes
- `parseAdhocScreenResult` / `parseWatchlistEntries` / `parseRows` handle the different response formats
- `coachTreeEntries` is a recursive tree walker for coach screen nested structure
- Response parsing varies significantly by catalog kind; do not assume uniform shapes

## `stock.go` specifics

`GetStock()` does date template substitution in the GraphQL query (replaces `{{date}}` with today's date) and navigates a deeply nested response to build `models.StockData` with sub-structs: Ratings, Company, Pricing, Financials, QuarterlyFinancials.
