---
applyTo: "cmd/**/*.go"
---

# Command review instructions

- Commands are kong structs with `Run(client *marketsurge.Client) error`.
- Auth runs in `main.go` before `ctx.Run`; commands receive an already-authenticated client.
- Success output is a raw JSON array written to stdout (or an `io.Writer` for testability).
- Keep porcelain commands like `overview <symbol>` inside the same raw-array contract, even when returning one logical object.
- Error output uses `mserrors.WriteJSON(os.Stderr, err)` in `main.go`; commands return errors, they don't write them.
- Preserve request context propagation for cancellation.
