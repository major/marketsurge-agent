# Fundamental Skill

Use `fundamental get` when the question is only about earnings, sales, margins, estimates, or cash flow. Use `stock analyze` instead when price, ratings, or ownership are also needed.

## Command

```bash
marketsurge-agent fundamental get AAPL
```

Required arg: one ticker symbol.

## Output focus

- Historical EPS and sales with year-over-year changes.
- Future EPS and sales estimates.
- Quarterly earnings, sales, and margin breakdowns.
- Cash flow per share.
- Standard JSON envelope with symbol metadata.

## Agent guidance

- Ask for this before discussing growth consistency, estimate acceleration, or margin trends.
- Do not call this after `stock analyze` unless the prior analysis was run in `--summary` mode and omitted fundamentals.
- Pair with `rs-history get` when separating business improvement from price-relative strength.
