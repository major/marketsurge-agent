# Stock Skill

Use stock commands for current MarketSurge ratings, price context, base patterns, and signal flags.

## Commands

| Need | Command | Token guidance |
|---|---|---|
| One-symbol quote, ratings, base, signals | `marketsurge-agent stock get AAPL` | Smallest stock-only call |
| Complete research packet | `marketsurge-agent stock analyze AAPL` | Includes stock, fundamentals, ownership |
| Compare many candidates | `marketsurge-agent stock analyze --summary AAPL NVDA TSLA` | Best low-token ranking mode |
| Batch from a generated list | `marketsurge-agent stock analyze --tickers AAPL,NVDA,TSLA --compact` | Avoid shell loops |
| Parser wants one-level keys | `marketsurge-agent stock analyze AAPL --flat` | Use only when nesting is inconvenient |

## `stock get <symbol>`

Use for targeted current stock data when fundamentals and ownership are not needed.

Output focus: ratings, price, valuation ratios, risk metrics, short interest, `base_pattern`, and `signals` such as blue dot and ant signal.

## `stock analyze [symbols...]`

Fetches stock, fundamentals, and ownership concurrently. Accepts positional symbols and `--tickers` comma-separated symbols together. Multi-symbol work is concurrent and can return partial results when only some symbols or subresources fail.

Flags:

- `--summary`: emits compact screening keys: `symbol`, `composite`, `eps`, `rs`, `ad`, `smr`, `blue_dot`, `ant_signal`, `base_type`, `base_stage`, `pivot`, `base_depth_percent`, `industry_group_rs`, `up_down_volume`, `atr_percent`, `avg_dollar_volume`, `funds_float_percent`.
- `--compact`: removes duplicate formatted string fields while keeping raw values.
- `--flat`: flattens nested analysis fields inside the JSON envelope.

Decision rule: start with `stock analyze --summary` for candidate ranking, then rerun interesting symbols without `--summary` for detail.
