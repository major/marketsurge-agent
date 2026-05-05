---
applyTo: "internal/client/**/*.go"
---

# Client review instructions

- JWT and Cookie headers must be set per request, not in default or base headers.
- GraphQL queries must be loaded from embedded files through `queries.Load()`.
- HTTP and GraphQL errors must be wrapped in typed project errors.
- Preserve request context propagation for cancellation.
- Avoid adding fields to GraphQL requests that increase payload size without command value.
