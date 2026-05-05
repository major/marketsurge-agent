---
applyTo: "**/*_test.go"
---

# Test review instructions

- Use `testify/assert` and `testify/require`, not bare if checks.
- Mock HTTP with `httptest.NewServer` and request capture.
- Use shared helpers such as `testClient()`, `jsonServer()`, and fixture builders where they exist.
- Prefer table-driven subtests with `t.Run()`.
- Check typed errors with `assert.ErrorAs()` or `require.ErrorAs()`.
- CLI tests should suppress exits via `ExitErrHandler` and capture output with `bytes.Buffer`.
