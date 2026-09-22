# bff

The BFF (`docs/ARCHITECTURE.md` §6): the only thing the phone talks to, owning no database. It
serves `GetHome` and `RequestMagicLink`; the rest of `BffService` is planned. Every RPC but
`RequestMagicLink` passes the bearer check every service shares (`apps/README.md`, "Every
service") before any service is called, and the token is then forwarded unchanged (§4.2).

`RequestMagicLink` (#69) is exempt from that check by procedure name, because it is sent before a
Member has a token, so the handler reads the address from the request and never from the context's
claims. It gets a login token from the realm on A10's terms, as the `bff` service account through
the Phase Two extension's `login-token` endpoint, builds `BFF_MAGIC_LINK_BASE_URL?lt=…` from it, and
hands the address and the link to `notify.EmailMagicLink` (§7.5), exempt on the same rule (#68). The
`Idempotency-Key` header is copied when the caller sends one and is a fresh version 4 UUID when it
does not, because `notify` rejects an empty key (A14).

Before any of that, the request is counted against its address in Redis, `ml:<address>:<hour>` with
the address lower-cased: §4.4's magic-link rate limit, five per address per hour (PRD §11), which
fails closed with `resource_exhausted` (§4.4, A6). It is the BFF's one Redis client.

`GetHome` first calls `members.GetOrCreateBySubject` with the token's subject and `email` claim (PRD
§5 step 5, first-sign-in detection), forwarding the caller's token unchanged so `members` verifies
it again, with a one-second timeout; the Profile it answers is the one the HomeView shows. Then the
§6.1 fan-out runs in parallel with a one-second timeout each: `ledger.GetMonthSummary` (required) and
`notify.GetWarningState` (optional). A failed `members` call reaches the phone with its own Connect
code, a failed `ledger` call as `unavailable`, each with a message naming only the call; the upstream
text, which can hold a host or a database name, goes to the log. `ledger` and `notify` are the
in-process stubs `internal/bff/bff.go` describes, returning zeros and "no warning". The
join fills `currency_code` with `GBP` and leaves `display_name` empty for a Member who has not
onboarded, the two skeleton-only states spec #10 records; `days_remaining` and `first_recap_on` come
from the server clock in UTC until `SetTimezone` lands (A3).

| Path | Holds |
| --- | --- |
| `main.go` | The binary; serves on `BFF_LISTEN` (default `:8080`). No subcommands, and its own start-up rather than `serve.Main` (see below). |
| `../../internal/bff/` | The handlers, the `Ledger` and `Notify` interfaces and their stubs, the magic-link path in `magiclink.go`, the unit tests and the three end-to-end tests (`TestEndToEnd`, `TestMagicLinkEndToEnd`, `TestEdge`). |
| `../../internal/auth/` | The JWT interceptor and the token-forwarding client interceptor. |

The BFF keeps its own `main.go` rather than `serve.Main` (`apps/README.md`, "Every service") because
it has none of the three things that body is: no database URL, no pool to open, no schema to apply,
and a `/healthz` with nothing to ping (§6). Putting it
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
| `BFF_CLIENT_SECRET` | required | The `bff` client's secret, which buys the service-account token the `login-token` endpoint needs. An OpenTofu variable rather than a generated value (§7.5), and the one secret the BFF reads; `BFF_CLIENT_SECRET_FILE` works like every other `_FILE`. |
| `BFF_MAGIC_LINK_BASE_URL` | required | The address the link is built on, the `lt` token its query: `http://<domain>:80/magic` on the machine today and `https://<domain>/magic` once TLS arrives (#87), `http://localhost:8090/magic` on a laptop, where Nginx serves the fallback page on `NGINX_HOST_PORT` (#72). |
| `BFF_REDIS_ADDR` | required | Redis for the magic-link limit: `redis:6379` in Compose, `localhost:REDIS_HOST_PORT` from the host. No default, because a BFF that cannot find Redis refuses every magic-link request (§4.4). |
| `BFF_LISTEN` | `:8080` | Listen address. |

## Run

The normal way is the whole stack (`infra/compose/README.md`, "Run it"), which starts `bff` once
`members` and Redis are healthy and the realm step has completed, and publishes it on
`BFF_HOST_PORT`.

By hand, against that stack's Keycloak and `members`, on a port the stack does not publish:

```bash
cd apps/backend
BFF_OIDC_ISSUER=http://localhost:8080/realms/mmr BFF_OIDC_JWKS_URL=http://localhost:8080/realms/mmr/protocol/openid-connect/certs BFF_MEMBERS_URL=http://localhost:8081 BFF_NOTIFY_URL=http://localhost:8083 BFF_REALM_URL=http://localhost:8080/realms/mmr BFF_CLIENT_SECRET="$BFF_CLIENT_SECRET" BFF_MAGIC_LINK_BASE_URL=http://localhost:8090/magic BFF_REDIS_ADDR=localhost:6379 BFF_LISTEN=:8084 go run ./cmd/bff
```

A token for the realm's test Member and the `GetHome` call are in `infra/compose/README.md`, "Try it".

## Tests

`go test ./internal/bff/` runs the unit tests, which need nothing; each test's name and doc comment
say what it proves. `TestEndToEnd`, `TestMagicLinkEndToEnd` and `TestEdge` drive the running stack at
the BFF's HTTP boundary and through the edge; their variables are in `infra/compose/README.md`,
"Running the Go tests against this stack".
