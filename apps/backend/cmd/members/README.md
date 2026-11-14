# members

The `members` service (`docs/ARCHITECTURE.md` §7.1): who the Member is, keyed by the Keycloak
subject. It serves `GetOrCreateBySubject` and `GetProfile`, `SetTimezone`, `SaveProfile`
`RequestDeletion` and `CancelDeletion`, every RPC of `MembersService`; it consumes `ledger.currency.converted` and the two purged events, and runs the deletion job. Its image, bearer check, start-up, migrations and variables are every service's
(`apps/README.md`, "Every service"), plus four of its own.

| Variable | Default | Meaning |
| --- | --- | --- |
| `MEMBERS_AMQP_URL` | required | The broker its outbox relay publishes to and its consumer reads from: the `amqp_url` secret in Compose. |
| `MEMBERS_REALM_URL` | required | The realm, whose token endpoint issues `members`' own service-account token: `http://keycloak:8080/realms/mmr` in Compose. |
| `MEMBERS_REALM_ADMIN_URL` | required | The realm's admin address, where that account deletes Keycloak accounts: `http://keycloak:8080/admin/realms/mmr` in Compose. |
| `MEMBERS_CLIENT_SECRET` | required | The `members` client's secret (`MEMBERS_CLIENT_SECRET_FILE` works too). |

| Path | Holds |
| --- | --- |
| `main.go` | The binary: its name, its schema step, its handler, its consumer and its outbox relay, handed to `serve.Main`. |
| `../../internal/members/` | Store, Connect handler, the events it stages (`announce.go`), the one it consumes (`consume.go`) and `migrations/*.sql` (embedded). |

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
from `internal/cldr`, the data `ledger` rounds with, which step 3's amount field allows.

`CancelDeletion` is "Keep my account" on the restore sheet (PRD §6.18, D60), keyed like `SetTimezone`: it
clears a `deletion_requested_at` under 24 hours old and stages `members.deletion.cancelled` in the same
transaction. With no stamp it answers the Profile and changes nothing; a stamp 24 hours old or more is
`failed_precondition`, since the deletion is then due. `RequestDeletion` is D11's Delete my account, keyed the
same way: it stamps `deletion_requested_at` and stages `members.deletion.requested` in one transaction, and a
Member already stamped keeps the first stamp, with nothing staged.

`Keycloak.DeleteAccount` deletes a Keycloak account by subject with `members`' own service account, which
holds `manage-users` for the deletion job alone (ARCHITECTURE §7.1 step 4, §4.2); a 404 counts as done.

## The deletion job

Every 10 minutes (`Sweep`, ARCHITECTURE §7.1 steps 3 to 6), `DeleteDue` tombstones each Member whose
`deletion_requested_at` is 24 hours old or more: one transaction per Member sets `deleted_at`, writes the
`deletion_progress` row and stages `members.deleted`. A tombstoned row is absent to every read, and
`GetOrCreateBySubject` answers its subject `permission_denied`. The same run deletes each tombstone's Keycloak
account still held, with `Keycloak.DeleteAccount`, and stamps `keycloak_deleted_at`; a failed call waits for the
next run. A second consumer reads `members.purges`, bound to `ledger.member.purged` and `notify.member.purged`,
and stamps the progress row. Whichever stamp completes the three hard-deletes the member row, its keys, its
outbox rows and the progress row. A deletion still incomplete an hour after its tombstone writes a log line on
every run, until observability lands (A11). `currency` takes no part.

## Events

A first sight stages `members.created`, and a write that changes any of the facts other services copy
(`MemberFacts`: the Display Name, zone, sign-up date, Currency and email) stages one `members.updated`
carrying all of them, in the outbox, inside the write's own transaction (ARCHITECTURE §4.3, A4). A write
that changes none of them stages nothing: an Avatar, a step, the same answer again, a replayed key.
`RequestDeletion` and `CancelDeletion` stage `members.deletion.requested` and `members.deletion.cancelled` the same
way, which `notify` consumes, and the deletion job stages `members.deleted`. The
relay beside the handler polls the outbox every second, publishes with publisher confirms and marks a row
sent only once the broker has confirmed it; `FOR UPDATE SKIP LOCKED` keeps two relays off the same row
(`internal/events`). A relay that loses the broker logs it and dials again five seconds later.

Its first consumer reads the quorum queue `members.ledger`, bound to `ledger.currency.converted`
(ARCHITECTURE §7.2 currency flow). Each message id is recorded in `processed_events` in the transaction that
applies it, so a second delivery changes nothing; the event sets `currency_code` and stages one
`members.updated`, unless a conversion newer than it was already applied (`currency_at`), so a late delivery
never takes the code back. A malformed event is rejected, never nacked, and after five tries waits in
`members.ledger.dead` for a person. Until the event arrives the Profile's Currency lags `ledger`'s by seconds;
every view the BFF answers takes its Currency from `ledger`, so no screen shows it.

## Tests

`go test ./internal/members/` runs them; each test's name and doc comment say what it proves. The
store-failure and caller-deadline tests need nothing; the rest need the stack's Postgres, and the
Connect-boundary test a real Keycloak token too. The account-deletion and deletion-job tests need Keycloak and
`members`' own secret, and delete only a throwaway account each creates; the job's tests run in a scratch
schema (`internal/testdb`), so the stack's relay and job never meet their rows. The variables are in `infra/compose/README.md`,
"Running the Go tests against this stack".
