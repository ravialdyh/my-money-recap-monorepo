# members

The `members` service (`docs/ARCHITECTURE.md` §7.1): who the Member is, keyed by the Keycloak
subject. Slice 1 (#15) answers `GetOrCreateBySubject` and `GetProfile` over Connect; every RPC
verifies the bearer JWT against the realm JWKS (issuer, audience `bff`, expiry, signature) and
returns `unauthenticated` otherwise, or `unavailable` when the keys it needs cannot be fetched or
arrive as a 200 that is not JSON (§4.2). Either answer carries a generic message; the verifier's
reason, which can quote the realm's answer, goes to the log.

| Path | Holds |
| --- | --- |
| `main.go` | The binary: this service's name, its schema step and its handler, handed to `serve.Main`. `members` serves on `MEMBERS_LISTEN` (default `:8080`); `members migrate` applies the schema and exits. |
| `../../Dockerfile` | The one Dockerfile for every binary, built with the module root as context: `docker build --build-arg CMD=members .` |
| `../../internal/members/` | Store, Connect handler and `migrations/*.sql` (embedded). |
| `../../internal/serve/` | `serve.Main`: the start-up, the `migrate` subcommand and the shutdown `notify` also uses. |
| `../../internal/migrate/` | The migration runner `notify` also uses. |
| `../../internal/auth/` | The JWT interceptor. |

## Environment

| Variable | Serve | Migrate | Meaning |
| --- | --- | --- | --- |
| `MEMBERS_DATABASE_URL_FILE` | one of the two | one of the two | The file holding the whole connection string, `/run/secrets/members_database_url` in Compose. Its trimmed contents win over `MEMBERS_DATABASE_URL` and it is never also reported missing (A13). |
| `MEMBERS_DATABASE_URL` | one of the two | one of the two | `postgres://members:<password>@postgres:5432/members`; the `members` role reaches only this database (A9). The plain form is what a local `go run` uses. |
| `MEMBERS_OIDC_ISSUER` | required | — | As `BFF_OIDC_ISSUER` (`../bff/README.md`). |
| `MEMBERS_OIDC_JWKS_URL` | required | — | As `BFF_OIDC_JWKS_URL`. |
| `MEMBERS_OIDC_AUDIENCE` | `bff` | — | The `aud` a token must carry. |
| `MEMBERS_LISTEN` | `:8080` | — | Listen address. |

Every other variable takes the same `<NAME>_FILE` form, read inside `serve.Env.Get`; the database
URL is the one that moved first, because `members` reads one whole connection string rather than a
password of its own (A13).

## Run

The normal way is the whole stack (`infra/compose/README.md`, "Run it"), which runs `members-migrate`
to completion and then starts `members`.

By hand, against that stack's Postgres and Keycloak, on a port the stack does not publish:

```bash
cd apps/backend
export MEMBERS_DATABASE_URL="postgres://members:$MEMBERS_DB_PASSWORD@localhost:5432/members"
go run ./cmd/members migrate
MEMBERS_OIDC_ISSUER=http://localhost:8080/realms/mmr MEMBERS_OIDC_JWKS_URL=http://localhost:8080/realms/mmr/protocol/openid-connect/certs MEMBERS_LISTEN=:8086 go run ./cmd/members
```

Call it with a token for the realm's test user (`infra/compose/README.md`, "Try it"):

```bash
curl -s http://localhost:8081/money_recap.v1.MembersService/GetOrCreateBySubject -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"subject":"<the token sub>","email":"test@example.com"}'
```

The subject and email in the request must be the token's own `sub` and `email` claim; anything else
is `permission_denied`, and the claim is what gets stored; a token without the claim, which a refresh
issues after an admin clears the address, keeps the stored one. `GetProfile` answers `not_found` for a
Member the token does not name. A store failure is `unavailable` with a generic message; the
driver's text, which names the database role, the database and its address, goes to the log. A call
whose own deadline passes, or that its caller cancels, is not a store failure: it answers
`deadline_exceeded` or `canceled` and is not logged, since the caller logs its own.

## Migrations

`internal/members/migrations/NNNN_name.sql`, applied in file-name order, each in one transaction,
recorded in `schema_migrations`. Add a new numbered file; never edit an applied one.

## Tests

`go test ./internal/members/` runs them; each test's name and doc comment say what it proves. The
store-failure and caller-deadline tests need nothing; the rest need the stack's Postgres, and the
Connect-boundary test a real Keycloak token too. The variables are in `infra/compose/README.md`,
"Running the Go tests against this stack".
