---
applyTo: "**/*.go"
---

# Go review instructions

- Wrap errors with `fmt.Errorf("context: %w", err)` when preserving causes.
- Use `errors.As()` for typed errors and constructor functions to create project errors.
- Commands are kong structs; no `init()` wiring or cobra constructors.
- Each command struct implements `Run(client *marketsurge.Client) error`.
- Success output goes to stdout as a raw JSON array; error output goes to stderr via `mserrors.WriteJSON`.
- Do not suggest style-only changes handled by gofmt or golangci-lint.
