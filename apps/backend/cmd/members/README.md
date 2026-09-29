# members

The `members` service (`docs/ARCHITECTURE.md` §7.1): who the Member is, keyed by the Keycloak
subject. It serves `GetOrCreateBySubject` and `GetProfile` (#15), `SetTimezone` (#96), `SaveProfile`
(#98) and `CancelDeletion` (#105); the rest of `MembersService` is planned. Its image, bearer check, start-up, migrations and variables are every service's
(`apps/README.md`, "Every service"), plus four of its own.

| Variable | Default | Meaning |
| --- | --- | --- |
| `MEMBERS_AMQP_URL` | required | The broker its outbox relay publishes to: the `amqp_url` secret in Compose. |
| `MEMBERS_REALM_URL` | required | The realm, whose token endpoint issues `members`' own service-account token: `http://keycloak:8080/realms/mmr` in Compose. |
| `MEMBERS_REALM_ADMIN_URL` | required | The realm's admin address, where that account deletes Keycloak accounts: `http://keycloak:8080/admin/realms/mmr` in Compose. |
| `MEMBERS_CLIENT_SECRET` | required | The `members` client's secret (`MEMBERS_CLIENT_SECRET_FILE` works too). |

| Path | Holds |
| --- | --- |
| `main.go` | The binary: its name, its schema step, its handler and its outbox relay, handed to `serve.Main`. |
| `../../internal/members/` | Store, Connect handler, the events it stages (`announce.go`) and `migrations/*.sql` (embedded). |

## Run

The normal way is the whole stack (`infra/compose/README.md`, "Run it"), which runs `members-migrate`
to completion and then starts `members`.

By hand, against that stack's Postgres, Keycloak and RabbitMQ, on a port the stack does not publish:

```bash
cd apps/backend
export MEMBERS_DATABASE_URL="postgres://members:$MEMBERS_DB_PASSWORD@localhost:5432/members"
go run ./cmd/members migrate
export MEMBERS_AMQP_URL="amqp://mmr:$RABBITMQ_PASSWORD@localhost:5672/"
MEMBERS_OIDC_ISSUER=http://localhost:8080/realms/mmr MEMBERS_OIDC_JWKS_URL=http://localhost:8080/realms/mmr/protocol/openid-connect/certs MEMBERS_REALM_URL=http://localhost:8080/realms/mmr MEMBERS_REALM_ADMIN_URL=http://localhost:8080/admin/realms/mmr MEMBERS_CLIENT_SECRET="$MEMBERS_CLIENT_SECRET" MEMBERS_LISTEN=:8086 go run ./cmd/members
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
and it is keyed like `SetTimezone`. Every Profile it answers carries `currency_digits`, the Currency's decimals
from `internal/cldr`, the data `ledger` rounds with, which step 3's amount field allows (#103).

`CancelDeletion` is "Keep my account" on the restore sheet (PRD §6.18, D60), keyed like `SetTimezone`: it
clears a `deletion_requested_at` under 24 hours old and stages `members.deletion.cancelled` in the same
transaction. With no stamp it answers the Profile and changes nothing; a stamp 24 hours old or more is
`failed_precondition`, since the deletion is then due. Nothing sets the stamp yet: `RequestDeletion`
arrives with the Settings slice, so the tests write it straight into the row.

`Keycloak.DeleteAccount` deletes a Keycloak account by subject with `members`' own service account, which
holds `manage-users` for the deletion job alone (#177, ARCHITECTURE §7.1 step 4, §4.2); a 404 counts as done.
No job calls it yet.

## Events

A first sight stages `members.created`, and a write that changes any of the facts other services copy
(`MemberFacts`: the Display Name, zone, sign-up date, Currency and email) stages one `members.updated`
carrying all of them, in the outbox, inside the write's own transaction (ARCHITECTURE §4.3, A4). A write
that changes none of them stages nothing: an Avatar, a step, the same answer again, a replayed key.
`CancelDeletion` stages `members.deletion.cancelled` the same way, which `notify` consumes since #176. The
relay beside the handler polls the outbox every second, publishes with publisher confirms and marks a row
sent only once the broker has confirmed it; `FOR UPDATE SKIP LOCKED` keeps two relays off the same row
(`internal/events`). A relay that loses the broker logs it and dials again five seconds later.

## Tests

`go test ./internal/members/` runs them; each test's name and doc comment say what it proves. The
store-failure and caller-deadline tests need nothing; the rest need the stack's Postgres, and the
Connect-boundary test a real Keycloak token too. The account-deletion test needs Keycloak and
`members`' own secret, and deletes only a throwaway account it creates. The variables are in `infra/compose/README.md`,
"Running the Go tests against this stack".
