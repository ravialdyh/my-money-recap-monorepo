# notify

The service that sends every push and every email: the sign-in link, the CSV export, the Target Warning, Recap
ready and the daily reminder.

## Pushes

| Push | On at first | When | Text |
| --- | --- | --- | --- |
| Recap ready | yes | 09:00 local on the 1st, if last month has data | "Your {Month} recap is ready. See how you did." |
| Daily reminder | no | The Member's chosen time, if no Spend is logged today | "Log today's spends before you forget." |
| Target Warning | yes | The first time a month's Spent passes the line | "You've passed your saving line for {Month}." |

Each push goes out at most once. None goes out while its switch is off, while a deletion is pending, or to a
phone that logged out.

## Recap ready, step by step

Every minute, `notify` picks the Members whose time has come and pushes to each one at most once.

```mermaid
flowchart TD
    tick["Every minute"]
    due["Members due now"]
    ask["Ask Ledger: last month?"]
    claim["Claim the sent log"]
    push["Apple push"]
    next["Set next fire time"]
    tick --> due
    due --> ask
    ask -- "has data" --> claim
    claim --> push
    push --> next
    ask -- "no data" --> next
```

## How it works

- **Calls it answers**: `EmailMagicLink`, `EmailExport`, `GetWarningState`, `EvaluateWarning`,
  `RegisterDevice`, `ForgetDevice`, `GetPrefs` and `SavePrefs`.
- **Target Warning**: stamped once per Member and month, either by the BFF's call or by the spend event,
  whichever comes first. The stamp shows Home's banner and sends the push.
- **Schedules**: each Member's next Recap time and next reminder time are kept in their own time zone. It reads
  `ledger` with its own service-account token.
- **Tables**: `month_flag`, `sent_log`, `device`, `notification_preference`, `member_facts` and
  `magic_link_send`.
- **Events it reads**: `ledger.spend.logged`, `ledger.spend.edited`, and every `members.*` event.
- **Events it sends**: `notify.member.purged`.
- **No push key** (every laptop): a push is logged and not sent.

Its image, token check, start-up, shared tables and shared settings are the same as every service's
([apps/README.md](../../../README.md)).

## Run it

The whole stack starts it with `--profile mail` ([compose README, "Run it"](../../../../infra/compose/README.md#run-it)), after
`notify-migrate` has run.

To run it by hand against that stack, on a port the stack does not use:

```bash
cd apps/backend
export NOTIFY_DATABASE_URL="postgres://notify:$NOTIFY_DB_PASSWORD@localhost:5432/notify"
go run ./cmd/notify migrate
NOTIFY_OIDC_ISSUER=http://localhost:8080/realms/mmr NOTIFY_OIDC_JWKS_URL=http://localhost:8080/realms/mmr/protocol/openid-connect/certs NOTIFY_SMTP_ADDR=localhost:1025 NOTIFY_MAIL_FROM=no-reply@mymoneyrecap.test NOTIFY_MEMBERS_URL=http://localhost:8081 NOTIFY_AMQP_URL="amqp://mmr:$RABBITMQ_PASSWORD@localhost:5672/" NOTIFY_LEDGER_URL=http://localhost:8084 NOTIFY_REALM_URL=http://localhost:8080/realms/mmr NOTIFY_CLIENT_SECRET="$NOTIFY_CLIENT_SECRET" NOTIFY_LISTEN=:8087 go run ./cmd/notify
```

The sign-in email it sends shows up in Mailpit ([TESTING.md, "Mailpit"](../../../../infra/compose/TESTING.md#mailpit)).

## Test it

From `apps/backend`:

```bash
go test ./internal/notify/
```

Most tests need the stack's `notify` database, and some also Mailpit, `members` and RabbitMQ
([TESTING.md, "Running the Go tests against this stack"](../../../../infra/compose/TESTING.md#running-the-go-tests-against-this-stack)).

## Settings

| Variable | Default | Meaning |
| --- | --- | --- |
| `NOTIFY_SMTP_ADDR` | required | The mail host: `mailpit:1025` in Compose |
| `NOTIFY_MAIL_FROM` | required | The sender address |
| `NOTIFY_MEMBERS_URL` | required | `members`: `http://members:8080` in Compose |
| `NOTIFY_LEDGER_URL` | required | `ledger`: `http://ledger:8080` in Compose |
| `NOTIFY_REALM_URL` | required | The realm that issues `notify`'s own token |
| `NOTIFY_CLIENT_SECRET` | required | The `notify` client's secret |
| `NOTIFY_APNS_HOST` | `https://api.push.apple.com` | Apple's push service |
| `NOTIFY_APNS_TOPIC` | `com.mymoneyrecap.app` | The app's bundle id |
| `NOTIFY_APPLE_TEAM_ID` | empty | The Apple Team ID |
| `NOTIFY_APPLE_PUSH_KEY_ID` | empty | The push key's Key ID |
| `NOTIFY_APPLE_PUSH_KEY` | empty | The `.p8` push key, in base64 on one line |
