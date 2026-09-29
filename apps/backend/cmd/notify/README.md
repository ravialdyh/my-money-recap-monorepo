# notify

The `notify` service (`docs/ARCHITECTURE.md` §7.3): every notification and every outbound email.
It serves `EmailMagicLink` (#68, PR #77) and, since #132, the Target Warning state:
`GetWarningState` and `EvaluateWarning`. Preferences, device registration, pushes, the CSV export
email, the schedulers, the other RabbitMQ consumers and the outbox are planned. Its image, bearer check, start-up, migrations and shared variables are every service's
(`apps/README.md`, "Every service").
Its `Ledger` client sends `notify`'s own service-account token, fetched with the client-credentials grant and
cached until it expires (#177, ARCHITECTURE §4.2), which `ledger` accepts for `GetSpentToday` and
`ListMonthsWithData` alone; no schedule calls it yet.

`EmailMagicLink` is exempt from the bearer check by procedure name, never by service, because the
BFF calls it for a request that carries no token (§4.2 point 3); every other procedure answers
`unauthenticated` without one. It sends the link over SMTP, upgrading to STARTTLS when the host
offers it, once per address and idempotency key (A14): the row is claimed first and a refused SMTP
host rolls it back. An empty key is `invalid_argument`.

| Path | Holds |
| --- | --- |
| `main.go` | The binary: its name, its schema step, its handler and its consumer, handed to `serve.Main`. |
| `../../internal/notify/` | `notify.go` (the magic-link email and its SMTP client), `warning.go` (the Target Warning, its event backup and the ownership check) and `migrations/*.sql` (embedded). |

## Environment

Its own four, besides the shared `NOTIFY_*` set:

| Variable | Serve | Migrate | Meaning |
| --- | --- | --- | --- |
| `NOTIFY_SMTP_ADDR` | required | none | The mail host as `host:port`: `mailpit:1025` in Compose. Mailpit runs only on a laptop and in CI, behind the `mail` profile the deploy never passes, so the machine has no mail host until #87 item 8's provider, which also needs SMTP AUTH `notify` does not send yet. |
| `NOTIFY_MAIL_FROM` | required | none | The envelope sender and the `From` header. |
| `NOTIFY_MEMBERS_URL` | required | none | `members`, for the profile read that proves a Member's owner: `http://members:8080` in Compose. |
| `NOTIFY_AMQP_URL` | required | none | The broker its consumer reads `ledger`'s spend events from: the `amqp_url` secret in Compose. |
| `NOTIFY_LEDGER_URL` | required | none | `ledger`, for the schedules' two reads: `http://ledger:8080` in Compose. |
| `NOTIFY_REALM_URL` | required | none | The realm, whose token endpoint issues `notify`'s own token: `http://keycloak:8080/realms/mmr` in Compose. |
| `NOTIFY_CLIENT_SECRET` | required | none | The `notify` client's secret (`NOTIFY_CLIENT_SECRET_FILE` works too). |

## Run

The normal way is the whole stack (`infra/compose/README.md`, "Run it"), which runs `notify-migrate`
to completion and then starts `notify`. That command carries `--profile mail`, without which
`notify` has no mail host (`infra/compose/README.md`, "Mailpit").

By hand, against that stack's Postgres, Keycloak, RabbitMQ and Mailpit, on a port the stack does not publish:

```bash
cd apps/backend
export NOTIFY_DATABASE_URL="postgres://notify:$NOTIFY_DB_PASSWORD@localhost:5432/notify"
go run ./cmd/notify migrate
NOTIFY_OIDC_ISSUER=http://localhost:8080/realms/mmr NOTIFY_OIDC_JWKS_URL=http://localhost:8080/realms/mmr/protocol/openid-connect/certs NOTIFY_SMTP_ADDR=localhost:1025 NOTIFY_MAIL_FROM=no-reply@mymoneyrecap.test NOTIFY_MEMBERS_URL=http://localhost:8081 NOTIFY_AMQP_URL="amqp://mmr:$RABBITMQ_PASSWORD@localhost:5672/" NOTIFY_LEDGER_URL=http://localhost:8084 NOTIFY_REALM_URL=http://localhost:8080/realms/mmr NOTIFY_CLIENT_SECRET="$NOTIFY_CLIENT_SECRET" NOTIFY_LISTEN=:8085 go run ./cmd/notify
```

The call and the message it produces are in `infra/compose/README.md`, "Mailpit".

## The Target Warning

`month_flag` holds one row per Member and month, primary key `(member_id, year_month)`, stamped once
with the time and ledger's warning copy as it read then (PRD §8.8). `EvaluateWarning` stamps it only
when `crossed_line` is true and the month is unstamped, and answers `fired`, `fired_now` and the kept
copy; the primary key is its A14 guarantee, so a repeat answers the same copy and only the stamping
call sees `fired_now`. It needs a key, and ledger's copy whenever `crossed_line` is true. It sends
nothing yet: pushes are slice 7. `GetWarningState` answers the same two facts for a month.

The event backup (#135, ARCHITECTURE §6.2 step 6) stamps the same way without the BFF's call: the
consumer reads `ledger.spend.logged` and `ledger.spend.edited` from the quorum queue `notify.ledger`, and
one with `crossed_line` stamps its month with its copy if unstamped, in the transaction that records its
message id in `processed_events`, so a second delivery changes nothing. A crossing event without a
Member, a month or a copy fails, and after five tries waits in `notify.ledger.dead` for a person.

Both check ARCHITECTURE §4.2 point 3 as ledger does: `member_facts` keeps the subject whose token
`members.GetProfile` last answered for the Member, and any other token is a miss that `members`
answers under that token, so a Member it does not own is `not_found` and `members` down is
`unavailable`. Feeding the replica from `members` events is slice 7.

## The magic-link table

The magic-link table is `magic_link_send`, unique on `(email, idempotency_key)` (§8). A14's `(member_id,
idempotency_key)` has no Member on this path, and the key alone was worse than no dedupe at all: a
key already claimed for one address answered success for a different address and sent nothing, so
`0002` moved the primary key onto the pair.

Every send first deletes rows older than the `retention` constant in `internal/notify/notify.go`,
in the same transaction. Nothing else would ever delete them: the row holds a raw address typed at
sign-in with no `member_id`, so ARCHITECTURE §7.3's `members.deleted` purge cannot reach it, and the
link it dedupes dies after fifteen minutes (PRD D32). Sweeping on the write path keeps the retention
bound off notify's planned scheduler.

## Tests

`go test ./internal/notify/` runs them; each test's name and doc comment say what it proves. Only the
empty-key test needs nothing: the Target Warning tests and the tests against fake mail hosts need the
`notify` database, and the rest also send through Mailpit. The variables are in `infra/compose/README.md`, "Running the Go tests
against this stack".
