# marketsurge-agent review instructions

Review this repository as a Go CLI that gives agents structured access to MarketSurge stock research data. The CLI is JSON-first, self-documenting through structcli, and authenticated by exchanging browser cookies for a JWT.

Focus on correctness, credential safety, broken command contracts, data loss, GraphQL/API behavior, and repository conventions. Do not nitpick formatting or style that gofmt or golangci-lint already handles.

## Project invariants

- Command output must use JSON envelopes through `internal/output` helpers.
- Errors must use the `MarketSurgeError` hierarchy and constructor functions.
- JWT and Cookie headers must be added per request in `client.Execute()`, not in base headers.
- GraphQL queries are embedded files and loaded through `queries.Load("name")`, not hardcoded strings.
- `--jsonschema`, `--jsonschema=tree`, `--mcp`, help topics, and generated `SKILL.md` should stay aligned with command behavior.
- Keep README, `AGENTS.md`, `SKILL.md`, CodeRabbit, and Copilot review guidance updated when command behavior or review priorities change.

## Security and API checks

- Flag any leak of cookies, JWTs, browser profile paths that expose secrets, GraphQL auth headers, or account data.
- Cookie database access should handle missing files and permission errors gracefully.
- Verify GraphQL query variables match the Go code that calls them.
- Concurrency should follow the existing `sync.WaitGroup` and `sync.Mutex` patterns for parallel stock analysis.
