---
applyTo: "queries/**/*.graphql"
---

# GraphQL review instructions

- Verify query variables match the Go caller.
- Check for unused fields that add unnecessary payload.
- Queries are embedded at build time with go:embed and should not be duplicated as string literals.
