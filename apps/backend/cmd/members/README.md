# members

The service that knows who each Member is: their sign-in subject, email, Display Name, Avatar, Currency, time
zone and sign-up date. It also runs account deletion.

## Deleting an account

A Member can cancel within 24 hours. After that, every service removes its data and Keycloak removes the login.

```mermaid
flowchart TD
    request["Member asks to delete"]
    wait["24 hours to cancel"]
    tombstone["Account hidden from reads"]
    keycloak["Keycloak login deleted"]
    services["Ledger and Notify purge"]
    gone["Member row deleted"]
    request --> wait
    wait --> tombstone
    tombstone --> keycloak
    tombstone -- "members.deleted" --> services
    keycloak --> gone
    services -- "purged events" --> gone
```

## How it works

- **Calls it answers**: `GetOrCreateBySubject`, `GetProfile`, `SetTimezone`, `SaveProfile`, `RequestDeletion`
  and `CancelDeletion`.
- **Checks**: a Display Name is 1 to 30 letters, spaces, hyphens or apostrophes. The Avatar is one of six.
  The Currency is a real ISO 4217 code. The time zone is a real IANA zone.
- **Sign-up date**: set once, on the first `SetTimezone`, as the local date in that zone.
- **Tables**: `member`, `deletion_progress`, `idempotency`, `outbox` and `processed_events`.
- **Events it sends**: `members.created`, `members.updated`, `members.deletion.requested`,
  `members.deletion.cancelled` and `members.deleted`.
- **Events it reads**: `ledger.currency.converted`, and the two purged events from `ledger` and `notify`.
- **Deletion job**: runs every 10 minutes. It deletes the Keycloak login with `members`' own service account.

Its image, token check, start-up and shared settings are the same as every service's
([apps/README.md](../../../README.md)).

## Settings

| Variable | Default | Meaning |
| --- | --- | --- |
| `MEMBERS_AMQP_URL` | required | RabbitMQ, for the events it reads and sends |
| `MEMBERS_REALM_URL` | required | The realm that issues `members`' own token: `http://keycloak:8080/realms/mmr` in Compose |
| `MEMBERS_REALM_ADMIN_URL` | required | Where it deletes logins: `http://keycloak:8080/admin/realms/mmr` in Compose |
| `MEMBERS_CLIENT_SECRET` | required | The `members` client's secret |

## Run it

The whole stack starts it (`infra/compose/README.md`, "Run it"), after `members-migrate` has run.

To run it by hand against that stack, on a port the stack does not use:

```bash
cd apps/backend
export MEMBERS_DATABASE_URL="postgres://members:$MEMBERS_DB_PASSWORD@localhost:5432/members"
go run ./cmd/members migrate
export MEMBERS_AMQP_URL="amqp://mmr:$RABBITMQ_PASSWORD@localhost:5672/"
MEMBERS_OIDC_ISSUER=http://localhost:8080/realms/mmr MEMBERS_OIDC_JWKS_URL=http://localhost:8080/realms/mmr/protocol/openid-connect/certs MEMBERS_REALM_URL=http://localhost:8080/realms/mmr MEMBERS_REALM_ADMIN_URL=http://localhost:8080/admin/realms/mmr MEMBERS_CLIENT_SECRET="$MEMBERS_CLIENT_SECRET" MEMBERS_LISTEN=:8086 go run ./cmd/members
```

## Test it

From `apps/backend`:

```bash
go test ./internal/members/
```

Most tests need the stack's Postgres; the deletion tests also need Keycloak (`infra/compose/README.md`,
"Running the Go tests against this stack").
