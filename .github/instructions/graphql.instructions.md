---
applyTo: "queries/**/*.graphql"
---

# GraphQL review instructions

- This project no longer uses embedded GraphQL queries. The marketsurge-go client handles API communication.
- If GraphQL files are added back, verify query variables match the Go caller and check for unused fields.
