# RS Rating History Skill

Use `rs-history get` to compare relative strength trends over time. It accepts one or more symbols and avoids shell loops.

## Command

```bash
marketsurge-agent rs-history get AAPL NVDA TSLA
```

Required args: one or more ticker symbols.

## Output focus

- Single symbol: symbol metadata plus that symbol's RS history.
- Multiple symbols: `data` object keyed by ticker.
- Includes RS rating snapshots and `rs_line_new_high` when provided.
- Partial multi-symbol failures can return successful symbols plus errors.

## Agent guidance

- Use this after `stock analyze --summary` when top candidates need RS trend confirmation.
- RS line new highs can identify leadership before price breaks out.
- Compare RS trend with `chart history` candles when checking divergence or confirmation.
