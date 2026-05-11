# marketsurge-agent review instructions

Review this repository as a Go CLI that gives agents structured access to MarketSurge stock research data. The CLI is JSON-first and authenticated by exchanging browser cookies for a JWT.

Focus on correctness, credential safety, broken command contracts, data loss, API behavior, and repository conventions. Do not nitpick formatting or style that gofmt or golangci-lint already handles.

## Project invariants

- Commands write raw JSON arrays to stdout; no envelope wrapper.
- `compare <symbols...>` compares short symbol lists through `MarketDataAdhocScreen`, preserves requested data under `columns`, and also emits LLM-oriented grouped keys.
- `industry <symbols...>` emits `[{ticker, industryGroupRS}]` from 6-month RS data; nil RS values are JSON `null`.
- `overview <symbol>` is a single-symbol porcelain command, but it still returns a one-element JSON array and uses compact LLM-oriented keys.
- `watchlist list` emits `[]WatchlistSummary`; empty result is `[]`, not `null`.
- `watchlist get <id>` emits a one-element array with LLM-oriented keys including a `symbols` list extracted from watchlist items.
- Errors must use the `MarketSurgeError` hierarchy and constructor functions.
- Error output goes to stderr as `{"code":"...","message":"..."}` via `mserrors.WriteJSON`.
- Auth is lazy via `kong.BindSingletonProvider`; only commands whose `Run` accepts `*marketsurge.Client` trigger it. Commands with `Run() error` (columns, reports catalog) skip auth.
- Keep README, `AGENTS.md`, CodeRabbit, and Copilot review guidance updated when command behavior or review priorities change.

## Security and API checks

- Flag any leak of cookies, JWTs, or browser profile paths that expose secrets.
- Cookie database access should handle missing files and permission errors gracefully.
- Verify API request parameters match the Go code that calls them.
