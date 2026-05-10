# marketsurge-agent review instructions

Review this repository as a Go CLI that gives agents structured access to MarketSurge stock research data. The CLI is JSON-first and authenticated by exchanging browser cookies for a JWT.

Focus on correctness, credential safety, broken command contracts, data loss, API behavior, and repository conventions. Do not nitpick formatting or style that gofmt or golangci-lint already handles.

## Project invariants

- Commands write raw JSON arrays to stdout; no envelope wrapper.
- `overview <symbol>` is a single-symbol porcelain command, but it still returns a one-element JSON array and uses compact LLM-oriented keys.
- Errors must use the `MarketSurgeError` hierarchy and constructor functions.
- Error output goes to stderr as `{"code":"...","message":"..."}` via `mserrors.WriteJSON`.
- Auth runs in `main.go` before `ctx.Run`; there is no per-command auth hook.
- Keep README, `AGENTS.md`, CodeRabbit, and Copilot review guidance updated when command behavior or review priorities change.

## Security and API checks

- Flag any leak of cookies, JWTs, or browser profile paths that expose secrets.
- Cookie database access should handle missing files and permission errors gracefully.
- Verify API request parameters match the Go code that calls them.
