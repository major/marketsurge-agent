# Catalog Skill

Use catalog commands to discover and fetch MarketSurge watchlists, screens, reports, and coach screens.

## Flow

1. Run `catalog list` to discover entries and IDs.
2. Run a returned `watchlist_id`, `report_id`, or `coach_screen_id` with `catalog run`.
3. Page and project results before deeper analysis.
4. Feed selected tickers into `stock analyze --summary` or full `stock analyze`.

## `catalog list`

```bash
marketsurge-agent catalog list --kind watchlist
```

`--kind` is optional. Valid values: `watchlist`, `screen`, `report`, `coach_screen`. Omit it to aggregate all sources. Partial source failures can still return entries from working sources.

Output focus: `entries[]` with `name`, `kind`, `description`, and the relevant ID field.

## `catalog run`

```bash
marketsurge-agent catalog run --kind report --report-id 456 --limit 50 --fields symbol,composite_rating,rs_rating
```

Required by kind:

| Kind | Required flag | Runnable |
|---|---|---|
| `watchlist` | `--watchlist-id` | Yes |
| `report` | `--report-id` | Yes |
| `coach_screen` | `--coach-screen-id` | Yes |
| `screen` | None | No, list only |

Useful flags:

- `--limit` and `--offset`: page large lists, default limit is 50.
- `--fields`: project columns such as `symbol`, `price`, `composite_rating`, `eps_rating`, `rs_rating`, `acc_dis_rating`, `smr_rating`, `industry_name`, `market_cap`, `volume_dollar_avg_50d`.
- `--min-composite`, `--min-rs`, `--exclude-spacs`: filter report or watchlist entries before returning data.

Gotcha: coach screen rows are paginated, but field projection and filters do not behave like report or watchlist rows.
