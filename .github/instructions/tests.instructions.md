---
applyTo: "**/*_test.go"
---

# Test review instructions

- Use `testify/assert` and `testify/require`, not bare if checks.
- Mock HTTP with `marketsurge.WithHTTPClient()` and `httptest.NewServer` (no external mock libraries).
- Tests live in the external `cmd_test` package and construct command structs directly.
- Prefer table-driven subtests with `t.Run()`.
- Check typed errors with `assert.ErrorAs()` or `require.ErrorAs()`.
- Commands that write to `os.Stdout` accept an `io.Writer` in their internal `run` method for output capture.
