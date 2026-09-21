# bff

The BFF (`docs/ARCHITECTURE.md` §6): the only thing the phone talks to, owning no database. Slice 1
(#17) answers `GetHome` with a HomeView of zeros. Every RPC verifies the bearer JWT against the
realm JWKS (issuer, audience `bff`, expiry, signature; keys cached in memory and refetched on an
unknown key id) and returns `unauthenticated` before any service is called (§4.2, A6), or
`unavailable` when the keys it needs cannot be fetched or arrive as a 200 that is not JSON, so the
phone keeps its session (PRD D61). Either answer carries a generic message; the verifier's reason,
which can quote the realm's answer, goes to the log.

`GetHome` first calls `members.GetOrCreateBySubject` with the token's subject and `email` claim (PRD
§5 step 5, first-sign-in detection), forwarding the caller's token unchanged so `members` verifies
it again, with the same one-second timeout. Then the §6.1 fan-out runs in parallel with a one-second
timeout each: `members.GetProfile` (required), `ledger.GetMonthSummary` and `notify.GetWarningState`
(optional). A failed `members` call reaches the phone with its own Connect code, a failed `ledger`
call as `unavailable`, each with a message naming only the call; the upstream text, which can hold a
host or a database name, goes to the log. `ledger` and `notify` are Go interfaces with in-process
stubs (`internal/bff`) returning zeros and "no warning"; slices 5 and 7 swap in Connect clients. The
join fills `currency_code` with `GBP` and leaves `display_name` empty for a Member who has not
onboarded, the two skeleton-only states spec #10 records; `days_remaining` and `first_recap_on` come
from the server clock in UTC until `SetTimezone` lands (A3).

| Path | Holds |
| --- | --- |
| `main.go` | The binary; serves on `BFF_LISTEN` (default `:8080`). No subcommands, and its own start-up rather than `serve.Main` (see below). |
| `../../Dockerfile` | The one Dockerfile for every binary, built with the module root as context: `docker build --build-arg CMD=bff .` |
| `../../internal/bff/` | The handler, the `Ledger` and `Notify` interfaces and their stubs, the unit tests and the end-to-end test. |
| `../../internal/auth/` | The JWT interceptor and the token-forwarding client interceptor. |

`members` and `notify` hand their name, their schema step and their handler to `serve.Main`, which
is their whole start-up: the connection pool, the `migrate` subcommand, the bearer check and the
shutdown, in one place, so `ledger` and `currency` add a service rather than a fourth copy of it.
The BFF keeps its own `main.go` because it has none of the three things that body is: no database
URL, no pool to open, no schema to apply, and a `/healthz` with nothing to ping (§6). Putting it
through the same call would mean a nil pool and a nil migration step threaded through every
service's start-up, which is a silent breakage away from a stack that does not come up.

## Environment

| Variable | Default | Meaning |
| --- | --- | --- |
| `BFF_OIDC_ISSUER` | required | The `iss` Keycloak stamps into tokens; Compose derives it from `MMR_HOST` and `KEYCLOAK_HOST_PORT` (`http://localhost:8080/realms/mmr` with the stock `.env`). |
| `BFF_OIDC_JWKS_URL` | required | Where to fetch the signing keys from this process: `http://keycloak:8080/realms/mmr/protocol/openid-connect/certs` inside the Compose network, `http://localhost:8080/...` from the host. The two are separate settings because the phone and a container reach Keycloak by different hostnames. |
| `BFF_OIDC_AUDIENCE` | `bff` | The `aud` a token must carry. |
| `BFF_MEMBERS_URL` | required | The `members` service: `http://members:8080` in Compose. |
| `BFF_LISTEN` | `:8080` | Listen address. |

## Run

The normal way is the whole stack (`infra/compose/README.md`, "Run it"), which starts `bff` once
`members` is healthy and publishes it on `BFF_HOST_PORT`.

By hand, against that stack's Keycloak and `members`, on a port the stack does not publish:

```bash
cd apps/backend
BFF_OIDC_ISSUER=http://localhost:8080/realms/mmr BFF_OIDC_JWKS_URL=http://localhost:8080/realms/mmr/protocol/openid-connect/certs BFF_MEMBERS_URL=http://localhost:8081 BFF_LISTEN=:8084 go run ./cmd/bff
```

A token for the realm's test user and the `GetHome` call are in `infra/compose/README.md`, "Try it".

## Tests

`go test ./internal/bff/` runs the unit tests: the zero HomeView, the caller's token forwarded to
`members`, a `notify` timeout still answering, a `members` timeout failing (on `GetProfile` and on
`GetOrCreateBySubject`), a failed upstream call keeping its code and dropping its text,
`days_remaining` on the last day of a month. The end-to-end test
`TestEndToEnd` needs the stack and the `BFF_TEST_*` variables listed at the top of
`internal/bff/e2e_test.go` (`apps/README.md`); CI runs it in the `backend` job.
