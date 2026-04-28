# Ownership Skill

Use `ownership get` for institutional sponsorship questions. Use `stock analyze` instead when ownership is only one part of a broader stock review.

## Command

```bash
marketsurge-agent ownership get AAPL
```

Required arg: one ticker symbol.

## Output focus

- Quarterly fund ownership counts.
- Funds as percentage of float.
- Ownership trend data in a standard JSON envelope.

## Agent guidance

- Rising fund count suggests growing institutional interest.
- High funds-float percentage can show sponsorship but may also imply crowded ownership.
- For screening many symbols, prefer `stock analyze --summary` because it includes `funds_float_percent` with other ranking fields.
