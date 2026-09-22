# members

The `members` service (`docs/ARCHITECTURE.md` §7.1): who the Member is, keyed by the Keycloak
subject. It serves `GetOrCreateBySubject` and `GetProfile` (#15), `SetTimezone` (#96) and `SaveProfile`
(#98); the rest of `MembersService` is planned. Its image, bearer check, start-up, migrations and variables are every service's
(`apps/README.md`, "Every service"); it reads no variable of its own.

| Path | Holds |
| --- | --- |
| `main.go` | The binary: its name, its schema step and its handler, handed to `serve.Main`. |
| `../../internal/members/` | Store, Connect handler and `migrations/*.sql` (embedded). |

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

Call it with a token for the realm's test Member (`infra/compose/README.md`, "Try it"):

```bash
curl -s http://localhost:8086/money_recap.v1.MembersService/GetOrCreateBySubject -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"subject":"<the token sub>","email":"test@example.com"}'
```

The subject and email in the request must be the token's own `sub` and `email` claim; anything else
is `permission_denied`, and the claim is what gets stored; a token without the claim, which a refresh
issues after an admin clears the address, keeps the stored one. `GetProfile` answers `not_found` for a
Member the token does not name. A store failure is `unavailable` with a generic message; the
driver's text, which names the database role, the database and its address, goes to the log. A call
whose own deadline passes, or that its caller cancels, is not a store failure: it answers
`deadline_exceeded` or `canceled` and is not logged, since the caller logs its own.

`SetTimezone` checks the zone against the zone data Go links into the binary (`time/tzdata`, because
the image has no zone files), stores it, and on the first call stamps `signed_up_on`: `created_at`'s
local date in that zone, which no later call changes. It needs an `idempotency_key`, claimed in the
`idempotency` table inside the write's own transaction under its `(member_id, idempotency_key)` key:
the same key again replays the first answer and writes nothing, the same key with a different request
is `invalid_argument`, and the next write sweeps keys older than 24 hours (A14).

`SaveProfile` writes only the answers a call carries, checked with `internal/answers` as the BFF checks
them: a Display Name stored in NFC and trimmed, 1 to 30 code points of letters in any script, spaces,
hyphens and apostrophes; one of the six Avatar keys `Profile.avatar_key` lists; an ISO 4217 Currency. It
keeps `onboarding_step` at the highest step ever saved, so saving an earlier step again never lowers it,
and it is keyed like `SetTimezone`.

## Tests

`go test ./internal/members/` runs them; each test's name and doc comment say what it proves. The
store-failure and caller-deadline tests need nothing; the rest need the stack's Postgres, and the
Connect-boundary test a real Keycloak token too. The variables are in `infra/compose/README.md`,
"Running the Go tests against this stack".
