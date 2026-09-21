# bff

The BFF (`docs/ARCHITECTURE.md` §6): the only thing the phone talks to, owning no database. Slice 1
(#17) answers `GetHome` with a HomeView of zeros. Every RPC but `RequestMagicLink` verifies the
bearer JWT against the realm JWKS (issuer, audience `bff`, expiry, signature; keys cached in memory
and refetched on an unknown key id) and returns `unauthenticated` before any service is called
(§4.2, A6), or `unavailable` when the keys it needs cannot be fetched or arrive as a 200 that is not
JSON, so the phone keeps its session (PRD D61). Either answer carries a generic message; the
verifier's reason, which can quote the realm's answer, goes to the log.

`RequestMagicLink` (#69) is the one procedure outside that check, exempted by procedure name in
`internal/auth` because it is sent before a Member has a token — which is the whole point of it — so
the handler reads the address from the request and never from the context's claims. It asks the
realm for a service-account token as the `bff` client, calls the Phase Two extension's `login-token`
endpoint on A10's terms (`expiration_seconds` 900, `reusable: false`, `force_create: true`,
`set_email_verified: true`), builds `BFF_MAGIC_LINK_BASE_URL?lt=…` from the `lt:<uuid>` it gets back,
and hands the address and the link to `notify.EmailMagicLink` (§7.5). That call carries no token
either, and reaches `notify` only because `EmailMagicLink` is exempt on the same rule (#68). The
`Idempotency-Key` header is copied when the caller sends one and is a fresh version 4 UUID per
accepted request when it does not: nothing sets that header until slice 4, and `notify` rejects an
empty key, so a missing one would dedupe every later send into silence (A14).

Before any of that, the request is counted against its address in Redis — §4.4's fifth job and the
BFF's first Redis client. PRD §11's five per address per hour is a security requirement, and the
key is the address rather than the token subject because this procedure carries no token. It
**fails closed**: a Redis that cannot be reached answers `resource_exhausted` rather than letting
the send through, unlike §4.4's fail-open per-Member limit, because A6 sets the rule per job and
the edge cannot stand behind this one — `limit_req` keys on an IP address and cannot see an email
address inside a request body. The counter is `INCR` and `EXPIRE` on `ml:<address>:<hour>`, one
round trip, with the address lower-cased so a capital letter cannot buy five more.

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
| `../../internal/bff/` | The handlers, the `Ledger` and `Notify` interfaces and their stubs, the magic-link path in `magiclink.go`, the unit tests and the end-to-end test. |
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
| `BFF_NOTIFY_URL` | required | The `notify` service: `http://notify:8080` in Compose. Only `RequestMagicLink` calls it; `GetHome`'s `notify` is still the in-process stub. |
| `BFF_REALM_URL` | required | The realm from this process, parent of the token and `login-token` endpoints: `http://keycloak:8080/realms/mmr` in Compose. Separate from `BFF_OIDC_ISSUER` for the same reason `BFF_OIDC_JWKS_URL` is. |
| `BFF_CLIENT_SECRET` | required | The `bff` client's secret, which buys the service-account token the `login-token` endpoint needs. An OpenTofu variable rather than a generated value (§7.5), and the one secret the BFF reads from the environment; `BFF_CLIENT_SECRET_FILE` works like every other. |
| `BFF_MAGIC_LINK_BASE_URL` | required | The address the link is built on, the `lt` token its query: `https://<domain>/magic` on the VM, `http://localhost:8082/magic` on a laptop, where nothing serves it until Nginx arrives (#72). |
| `BFF_REDIS_ADDR` | required | Redis for the magic-link limit: `redis:6379` in Compose, `localhost:REDIS_HOST_PORT` from the host. No default, because a BFF that cannot find Redis refuses every magic-link request (§4.4). |
| `BFF_LISTEN` | `:8080` | Listen address. |

## Run

The normal way is the whole stack (`infra/compose/README.md`, "Run it"), which starts `bff` once
`members` is healthy and publishes it on `BFF_HOST_PORT`.

By hand, against that stack's Keycloak and `members`, on a port the stack does not publish:

```bash
cd apps/backend
BFF_OIDC_ISSUER=http://localhost:8080/realms/mmr BFF_OIDC_JWKS_URL=http://localhost:8080/realms/mmr/protocol/openid-connect/certs BFF_MEMBERS_URL=http://localhost:8081 BFF_NOTIFY_URL=http://localhost:8083 BFF_REALM_URL=http://localhost:8080/realms/mmr BFF_CLIENT_SECRET="$BFF_CLIENT_SECRET" BFF_MAGIC_LINK_BASE_URL=http://localhost:8082/magic BFF_REDIS_ADDR=localhost:6379 BFF_LISTEN=:8084 go run ./cmd/bff
```

A token for the realm's test user and the `GetHome` call are in `infra/compose/README.md`, "Try it".

## Tests

`go test ./internal/bff/` runs the unit tests: the zero HomeView, the caller's token forwarded to
`members`, a `notify` timeout still answering, a `members` timeout failing (on `GetProfile` and on
`GetOrCreateBySubject`), a failed upstream call keeping its code and dropping its text,
`days_remaining` on the last day of a month; and for `RequestMagicLink`, the login-token request
carrying A10's four settings and the service account's own token, the address coming from the
request while the context holds a different Member's claims, a fresh key per request when the header
is absent, each way the realm or `notify` can refuse, the sixth request for an address refused
before the realm is called even when it changes case, and the closed door — a Redis that answers
nothing returns `resource_exhausted` rather than sending, which the end-to-end sixth request
would pass either way. The end-to-end test `TestEndToEnd` needs the stack and the `BFF_TEST_*`
variables listed at the top of `internal/bff/e2e_test.go` (`apps/README.md`); CI runs it in the
`backend` job.
