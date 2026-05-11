---
applyTo: "Makefile"
---

# Makefile review instructions

- Non-file targets should have `.PHONY` declarations.
- Avoid flags that are already tool defaults.
- Keep build, test, smoke, lint, and clean targets aligned with README and `AGENTS.md`.
