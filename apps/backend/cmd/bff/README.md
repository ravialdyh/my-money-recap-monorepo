# bff

The BFF (`docs/ARCHITECTURE.md` §6): the only thing the phone talks to, owning no database. Slice 1
(#17) answers `GetHome` with a HomeView of zeros. Every RPC verifies the bearer JWT against the
realm JWKS (issuer, audience `bff`, expiry, signature; keys cached in memory and refetched on an
unknown key id) and returns `unauthenticated` before any service is called (§4.2, A6).

`GetHome` first calls `members.GetOrCreateBySubject` with the token's subject and `email` claim
(PRD §5 step 5, first-sign-in detection), forwarding the caller's token unchanged so `members`
verifies it again. Then the §6.1 fan-out runs in parallel with a one-second timeout each:
`members.GetProfile` (required), `ledger.GetMonthSummary` and `notify.GetWarningState` (optional).
`ledger` and `notify` are Go interfaces with in-process stubs (`internal/bff`) returning zeros and
"no warning"; slices 5 and 7 swap in Connect clients. The join fills `currency_code` with `GBP` and
leaves `display_name` empty for a Member who has not onboarded, the two skeleton-only states spec #10
records; `days_remaining` and `first_recap_on` come from the server clock in UTC until `SetTimezone`
lands (A3).

| Path | Holds |
| --- | --- |
| `main.go` | The binary; serves on `BFF_LISTEN` (default `:8080`). No subcommands. |
| `Dockerfile` | Built with the module root `apps/backend` as context: `docker build -f cmd/bff/Dockerfile .` |
| `../../internal/bff/` | The handler, the `Ledger` and `Notify` interfaces and their stubs, the unit tests and the end-to-end test. |
| `../../internal/auth/` | The JWT interceptor and the token-forwarding client interceptor. |

## Environment

| Variable | Default | Meaning |
| --- | --- | --- |
| `BFF_OIDC_ISSUER` | required | The `iss` Keycloak stamps into tokens: `http://localhost:8080/realms/mmr`. |
| `BFF_OIDC_JWKS_URL` | required | Where to fetch the signing keys from this process: `http://keycloak:8080/realms/mmr/protocol/openid-connect/certs`. The two are separate settings because the phone and a container reach Keycloak by different hostnames. |
| `BFF_OIDC_AUDIENCE` | `bff` | The `aud` a token must carry. |
| `BFF_MEMBERS_URL` | required | The `members` service: `http://members:8080` in Compose. |
| `BFF_LISTEN` | `:8080` | Listen address. |

## Run

The normal way is the whole stack: `docker compose -f infra/compose/compose.yaml up -d --build --wait`
starts `bff` once `members` is healthy and publishes it on `BFF_HOST_PORT` (8082; Keycloak owns
8080 on the host, `infra/compose/README.md`).

By hand, against that stack's Keycloak and `members`:

```bash
cd apps/backend
BFF_OIDC_ISSUER=http://localhost:8080/realms/mmr BFF_OIDC_JWKS_URL=http://localhost:8080/realms/mmr/protocol/openid-connect/certs BFF_MEMBERS_URL=http://localhost:8081 BFF_LISTEN=:8083 go run ./cmd/bff
```

Call it with a token for the realm's test user (`infra/compose/README.md`, "Try it"):

```bash
curl -s http://localhost:8082/money_recap.v1.BffService/GetHome -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{}'
```

## Tests

`go test ./internal/bff/` runs the unit tests: the zero HomeView, the caller's token forwarded to
`members`, a `notify` timeout still answering, a `members` timeout failing, `days_remaining` on the
last day of a month. The end-to-end test `TestEndToEnd` needs the stack and the `BFF_TEST_*`
variables listed at the top of `internal/bff/e2e_test.go` (`apps/README.md`); CI runs it in the
`backend` job.
