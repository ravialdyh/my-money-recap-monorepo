# currency

The `currency` service (`docs/ARCHITECTURE.md` §7.4): the day's exchange rate from Open Exchange Rates,
for the Currency sheet's preview (PRD §8.9 step 1, D35). It serves `Quote` and nothing else. Its image and bearer
check are every service's (`apps/README.md`, "Every service"); it owns no database, no broker, no outbox and no
consumer (owner, 2026-09-29), so, like the BFF, it has its own small start-up in `main.go` rather than
`serve.Main`, and holds nothing about any Member, which keeps it out of account deletion (§7.1).

`Quote(from, to)` reads the provider's latest table against the US dollar, the only base the free plan serves, and
answers the cross rate, units of `to` per one `from`, as a decimal string rounded half away from zero to 8
significant digits, with source `PROVIDER` and the table's own date as `as_of`. A code the table lacks is
`failed_precondition` carrying `UnsupportedPair`, which the BFF turns into D7's typed rate. A failed fetch, a table
with no timestamp or no rates, and a stack with no App ID are `unavailable`, D6's failed state; with no App ID it
fetches nothing. A code no country uses is `invalid_argument`. It never touches an amount: rounding and rewriting
are `ledger`'s (A2).

The day's table is cached whole in Redis under `fx:{date}:USD` for 36 hours, the date being the server's in UTC.
A miss, a table Redis holds but cannot be read, or Redis down fetches again, one of the free plan's 1,000 a month.
The App ID travels in the `Authorization` header, never in an address a log could print.

Idle on a laptop stack it measured 9.2 MiB (`docker stats --no-stream`, 2026-09-30), under its siblings' 13 to
17 MiB, the owner's hosting ceiling (spec #169). Its cost is $0.

| Path | Holds |
| --- | --- |
| `main.go` | The binary: its variables, the Redis client, the bearer check and `serve.Run`. |
| `../../internal/currency/` | `Quote`, the fetch, the cache and the rounding. |

## Environment

| Variable | Default | Meaning |
| --- | --- | --- |
| `CURRENCY_OIDC_ISSUER` | required | The `iss` Keycloak stamps into tokens, as every service's. |
| `CURRENCY_OIDC_JWKS_URL` | required | Where this process fetches the signing keys. |
| `CURRENCY_OIDC_AUDIENCE` | `bff` | The `aud` a Member's token must carry. |
| `CURRENCY_REDIS_ADDR` | required | Redis for the day's table: `redis:6379` in Compose. |
| `CURRENCY_OPEN_EXCHANGE_RATES_APP_ID` | empty | The App ID, from `.env`'s `OPEN_EXCHANGE_RATES_APP_ID`, empty on every laptop stack; the machine's comes from the secrets file (`infra/secrets/README.md`). `CURRENCY_OPEN_EXCHANGE_RATES_APP_ID_FILE` works too. |
| `CURRENCY_LISTEN` | `:8080` | Listen address. |

## Run

The normal way is the whole stack (`infra/compose/README.md`, "Run it"), which starts `currency` once Redis is
healthy and publishes it on `CURRENCY_HOST_PORT` (8085 by default) on loopback. The BFF reaches it as
`BFF_CURRENCY_URL` and never waits for it to start.

## Tests

`go test ./internal/currency/` proves `Quote` against a fake rates site in the test and the stack's Redis
(`BFF_TEST_REDIS`): the cross rate and its date, a second quote not fetching, the 36-hour key, the unsupported
pair, a failed fetch, a broken table, no App ID, and Redis down fetching every time. Their clock sits in 2001, so
their keys never meet the service's own. `TestCurrencyChangeEndToEnd` (`internal/bff`) proves the stack's empty
App ID answers `unavailable` through the API while a typed rate still previews.
