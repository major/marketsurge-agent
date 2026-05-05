---
applyTo: "internal/auth/**/*.go"
---

# Auth review instructions

- Cookie database precedence is explicit `--cookie-db`, then Firefox auto-discovery.
- The token exchange happens at investors.com and the JWT is used for GraphQL at dowjones.io.
- Verify no cookies, JWTs, or credentials are logged or exposed in errors.
- Cookie database access should handle missing files and permission errors gracefully.
