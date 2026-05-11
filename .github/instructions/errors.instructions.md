---
applyTo: "internal/errors/**/*.go"
---

# Error review instructions

- All project errors should embed `MarketSurgeError`.
- Use constructor functions, not raw struct literals.
- Keep exit codes aligned with `AGENTS.md`.
- Import alias convention: `mserrors` in commands and main.
