---
applyTo: "internal/**/*.go"
---

# Go review instructions

- Wrap errors with `fmt.Errorf("context: %w", err)` when preserving causes.
- Use `errors.As()` for typed errors and constructor functions to create project errors.
- Avoid `init()` functions except for the existing Cobra command wiring pattern.
- Command factories should follow the existing command constructor style.
- All user-facing output must go through JSON envelope helpers.
- Use `sync.WaitGroup` and `sync.Mutex` for concurrent analysis patterns.
- Do not suggest style-only changes handled by gofmt or golangci-lint.
