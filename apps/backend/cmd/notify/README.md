# notify

The `notify` service (`docs/ARCHITECTURE.md` §7.3): every notification and every outbound email.
It serves `EmailMagicLink` alone (#68, PR #77); the Target Warning state, preferences, device
registration, the CSV export email, the schedulers, the RabbitMQ consumers and the outbox are
planned. Its image, bearer check, start-up, migrations and shared variables are every service's
(`apps/README.md`, "Every service").

`EmailMagicLink` is exempt from the bearer check by procedure name, never by service, because the
BFF calls it for a request that carries no token (§4.2 point 3); every other procedure answers
`unauthenticated` without one. It sends the link over SMTP, upgrading to STARTTLS when the host
offers it, once per address and idempotency key (A14): the row is claimed first and a refused SMTP
host rolls it back. An empty key is `invalid_argument`.

| Path | Holds |
| --- | --- |
| `main.go` | The binary: its name, its schema step and its handler, handed to `serve.Main`. |
| `../../internal/notify/` | The store, the SMTP client, the Connect handler and `migrations/*.sql` (embedded). |

## Environment

Its own two, besides the shared `NOTIFY_*` set:

| Variable | Serve | Migrate | Meaning |
| --- | --- | --- | --- |
| `NOTIFY_SMTP_ADDR` | required | — | The mail host as `host:port`: `mailpit:1025` in Compose. Mailpit runs only on a laptop and in CI, behind the `mail` profile the deploy never passes, so the machine has no mail host until #87 item 8's provider, which also needs SMTP AUTH `notify` does not send yet. |
| `NOTIFY_MAIL_FROM` | required | — | The envelope sender and the `From` header. |

## Run

The normal way is the whole stack (`infra/compose/README.md`, "Run it"), which runs `notify-migrate`
to completion and then starts `notify`. That command carries `--profile mail`, without which
`notify` has no mail host (`infra/compose/README.md`, "Mailpit").

By hand, against that stack's Postgres, Keycloak and Mailpit, on a port the stack does not publish:

```bash
cd apps/backend
export NOTIFY_DATABASE_URL="postgres://notify:$NOTIFY_DB_PASSWORD@localhost:5432/notify"
go run ./cmd/notify migrate
NOTIFY_OIDC_ISSUER=http://localhost:8080/realms/mmr NOTIFY_OIDC_JWKS_URL=http://localhost:8080/realms/mmr/protocol/openid-connect/certs NOTIFY_SMTP_ADDR=localhost:1025 NOTIFY_MAIL_FROM=no-reply@mymoneyrecap.test NOTIFY_LISTEN=:8085 go run ./cmd/notify
```

The call and the message it produces are in `infra/compose/README.md`, "Mailpit".

## The table

The one table is `magic_link_send`, unique on `(email, idempotency_key)` (§8). A14's `(member_id,
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
empty-key test needs nothing: the tests against fake mail hosts need the `notify` database, and the
rest also send through Mailpit. The variables are in `infra/compose/README.md`, "Running the Go tests
against this stack".
