# notify

The `notify` service (`docs/ARCHITECTURE.md` §7.3): every notification and every outbound email.
Slice 3 (#68) serves `EmailMagicLink` and nothing else; the Target Warning state, preferences,
device registration, the CSV export email, the schedulers, the RabbitMQ consumers and the outbox
arrive in slice 7.

`EmailMagicLink` is a pre-authentication procedure: the BFF reaches it on a request that carries no
token, so it is exempt from the bearer check by procedure name and never by service (§4.2 point 3,
`internal/auth`). Every other procedure of the same service answers `unauthenticated`
without a valid token, as `members` does. It sends the link over plain SMTP, and one idempotency key
sends one message: the row is claimed first and a refused SMTP host rolls it back, so a retried send
answers success without a second email and a failed one leaves the key free (A14). An empty key is
`invalid_argument`, because a missing key would dedupe every later send into silence.

| Path | Holds |
| --- | --- |
| `main.go` | The binary: this service's name, its schema step and its handler, handed to `serve.Main`. `notify` serves on `NOTIFY_LISTEN` (default `:8080`); `notify migrate` applies the schema and exits. |
| `../../Dockerfile` | The one Dockerfile for every binary, built with the module root as context: `docker build --build-arg CMD=notify .` |
| `../../internal/notify/` | The store, the SMTP client, the Connect handler and `migrations/*.sql` (embedded). |
| `../../internal/serve/` | `serve.Main`: the start-up, the `migrate` subcommand and the shutdown `members` also uses. |
| `../../internal/migrate/` | The migration runner `members` also uses. |
| `../../internal/auth/` | The JWT interceptor and its pre-authentication exemption. |

## Environment

| Variable | Serve | Migrate | Meaning |
| --- | --- | --- | --- |
| `NOTIFY_DATABASE_URL_FILE` | one of the two | one of the two | The file holding the whole connection string, `/run/secrets/notify_database_url` in Compose. Its trimmed contents win over `NOTIFY_DATABASE_URL` and it is never also reported missing (A13). |
| `NOTIFY_DATABASE_URL` | one of the two | one of the two | `postgres://notify:<password>@postgres:5432/notify`; the `notify` role reaches only this database (A9). The plain form is what a local `go run` uses. |
| `NOTIFY_SMTP_ADDR` | required | — | The mail host as `host:port`: `mailpit:1025` in Compose, a provider on the VM. |
| `NOTIFY_MAIL_FROM` | required | — | The envelope sender and the `From` header. |
| `NOTIFY_OIDC_ISSUER` | required | — | As `BFF_OIDC_ISSUER` (`../bff/README.md`). |
| `NOTIFY_OIDC_JWKS_URL` | required | — | As `BFF_OIDC_JWKS_URL`. |
| `NOTIFY_OIDC_AUDIENCE` | `bff` | — | The `aud` a token must carry. |
| `NOTIFY_LISTEN` | `:8080` | — | Listen address. |

## Run

The normal way is the whole stack (`infra/compose/README.md`, "Run it"), which runs `notify-migrate`
to completion and then starts `notify`. That command carries `--profile mail`, and it has to:
`NOTIFY_SMTP_ADDR` is `mailpit:1025` and Mailpit is behind that profile, so an `up` without it
leaves `notify` reporting healthy — `/healthz` pings only the database — while every send answers
`unavailable` with `dial tcp: lookup mailpit ... no such host` in the log
(`infra/compose/README.md`, "Mailpit").

By hand, against that stack's Postgres, Keycloak and Mailpit, on a port the stack does not publish:

```bash
cd apps/backend
export NOTIFY_DATABASE_URL="postgres://notify:$NOTIFY_DB_PASSWORD@localhost:5432/notify"
go run ./cmd/notify migrate
NOTIFY_OIDC_ISSUER=http://localhost:8080/realms/mmr NOTIFY_OIDC_JWKS_URL=http://localhost:8080/realms/mmr/protocol/openid-connect/certs NOTIFY_SMTP_ADDR=localhost:1025 NOTIFY_MAIL_FROM=no-reply@mymoneyrecap.test NOTIFY_LISTEN=:8085 go run ./cmd/notify
```

The call and the message it produces are in `infra/compose/README.md`, "Mailpit".

## Migrations

`internal/notify/migrations/NNNN_name.sql`, applied in file-name order, each in one transaction,
recorded in `schema_migrations`. Add a new numbered file; never edit an applied one. The one table
is `magic_link_send`, unique on `(email, idempotency_key)` (§8). A14's `(member_id,
idempotency_key)` has no Member on this path, and the key alone was worse than no dedupe at all: a
key already claimed for one address answered success for a different address and sent nothing, so
`0002` moved the primary key onto the pair.

Every send first deletes rows older than the `retention` constant in `internal/notify/notify.go`,
in the same transaction. Nothing else would ever delete them: the row holds a raw address typed at
sign-in with no `member_id`, so ARCHITECTURE §7.3's `members.deleted` purge cannot reach it, and the
link it dedupes dies after fifteen minutes (PRD D32). Sweeping on the write path keeps the retention
bound off slice 7's scheduler.

## Tests

`go test ./internal/notify/` runs the empty-key test and the Mailpit-polling helper's own test,
which need nothing. The rest need the stack, Mailpit and the `NOTIFY_TEST_*` variables listed in
`infra/compose/README.md`. The boundary test calls `EmailMagicLink` with no bearer token, reads the
message through Mailpit's HTTP interface, and asserts that the same key twice sends one message and
that `GetPrefs` is still `unauthenticated`. The rest drive the SMTP client against fake mail hosts
on a free port: one that accepts and then says nothing, which must not outlive the caller's deadline
or block the pool's `Close`; one refusing the recipient with a permanent 5xx, which is
`invalid_argument` while every other SMTP failure stays `unavailable`; and one that would accept an
address holding CR or LF, which `net/smtp` refuses before any recipient. CI runs them all in the
`backend` job.
