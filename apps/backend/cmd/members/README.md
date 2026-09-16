# members

The `members` service (`docs/ARCHITECTURE.md` §7.1): who the Member is, keyed by the Keycloak
subject. Slice 1 (#15) answers `GetOrCreateBySubject` and `GetProfile` over Connect; every RPC
verifies the bearer JWT against the realm JWKS (issuer, audience `bff`, expiry, signature) and
returns `unauthenticated` otherwise (§4.2).

| Path | Holds |
| --- | --- |
| `main.go` | The binary. `members` serves on `MEMBERS_LISTEN` (default `:8080`); `members migrate` applies the schema and exits. |
| `Dockerfile` | Built with the module root `apps/backend` as context: `docker build -f cmd/members/Dockerfile .` |
| `../../internal/members/` | Store, Connect handler and `migrations/*.sql` (embedded). |
| `../../internal/auth/` | The JWT interceptor. |

## Environment

| Variable | Serve | Migrate | Meaning |
| --- | --- | --- | --- |
| `MEMBERS_DATABASE_URL` | required | required | `postgres://members:<password>@postgres:5432/members`; the `members` role reaches only this database (A9). |
| `MEMBERS_OIDC_ISSUER` | required | — | The `iss` Keycloak stamps into tokens: `http://localhost:8080/realms/mmr`. |
| `MEMBERS_OIDC_JWKS_URL` | required | — | Where to fetch the signing keys from this process: `http://keycloak:8080/realms/mmr/protocol/openid-connect/certs`. |
| `MEMBERS_OIDC_AUDIENCE` | `bff` | — | The `aud` a token must carry. |
| `MEMBERS_LISTEN` | `:8080` | — | Listen address. |

Slice 3 moves the password to a `*_FILE` secret (A13).

## Run

The normal way is the whole stack: `docker compose -f infra/compose/compose.yaml up -d --build --wait`
runs `members-migrate` to completion and then starts `members` (`infra/compose/README.md`).

By hand, against that stack's Postgres and Keycloak:

```bash
cd apps/backend
export MEMBERS_DATABASE_URL="postgres://members:$MEMBERS_DB_PASSWORD@localhost:5432/members"
go run ./cmd/members migrate
MEMBERS_OIDC_ISSUER=http://localhost:8080/realms/mmr MEMBERS_OIDC_JWKS_URL=http://localhost:8080/realms/mmr/protocol/openid-connect/certs MEMBERS_LISTEN=:8082 go run ./cmd/members
```

Call it with a token for the realm's test user (`infra/compose/README.md`, "Try it"):

```bash
curl -s http://localhost:8081/money_recap.v1.MembersService/GetOrCreateBySubject -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"subject":"<the token sub>","email":"test@example.com"}'
```

The subject in the request must be the token's own `sub`; anything else is `permission_denied`, and
`GetProfile` answers `not_found` for a Member the token does not name.

## Migrations

`internal/members/migrations/NNNN_name.sql`, applied in file-name order, each in one transaction,
recorded in `schema_migrations`. Add a new numbered file; never edit an applied one.

## Tests

`go test ./...` runs the unit tests (JWT verification against a local JWKS). The boundary tests need
the stack and the `MEMBERS_TEST_*` variables listed at the top of `internal/members/members_test.go`:
the concurrent-upsert test against the real Postgres, and the Connect calls through a real Keycloak
token. CI runs them in the `backend` job.
