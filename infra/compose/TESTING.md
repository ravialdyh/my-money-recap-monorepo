# Testing against the Compose stack

How to try the backend by hand, and run its tests, against the stack that the [compose README, "Run
it"](README.md#run-it) starts.

## Try it by hand

A token for the test Member, kept in `$TOKEN`:

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/realms/mmr/protocol/openid-connect/token -d grant_type=password -d client_id=ios -d username=test -d "password=$TEST_USER_PASSWORD" | jq -r .access_token)
```

Home through the BFF with that token. Until the test Member has saved setup steps 1 and 2, it answers
`failed_precondition`.

```bash
curl -s http://localhost:8082/money_recap.v1.BffService/GetHome -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{}'
```

The same call through the edge answers the same:

```bash
curl -s http://localhost:8090/api/money_recap.v1.BffService/GetHome -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{}'
```

## Mailpit

Every email `notify` sends lands in Mailpit. Read it at `http://localhost:8025`, or over its API:

```bash
curl -s "http://localhost:8025/api/v1/search?query=to:someone@example.test" | jq '.messages[0].ID'
```

## Running the Go tests against this stack

The tests that need the stack skip unless their variables are set. With the stack up and the stock `.env`:

| Variable | Value |
| --- | --- |
| `MMR_TEST_KEYCLOAK` | `http://localhost:8080` |
| `MMR_TEST_USER_PASSWORD` | `$TEST_USER_PASSWORD` |
| `MMR_TEST_MAILPIT` | `http://localhost:8025` |
| `MMR_TEST_AMQP_URL` | `amqp://mmr:$RABBITMQ_PASSWORD@localhost:5672/` |
| `MMR_TEST_MEMBERS_DATABASE_URL` | `postgres://members:$MEMBERS_DB_PASSWORD@localhost:5432/members` |
| `NOTIFY_TEST_DATABASE_URL` | `postgres://notify:$NOTIFY_DB_PASSWORD@localhost:5432/notify` |
| `LEDGER_TEST_DATABASE_URL` | `postgres://ledger:$LEDGER_DB_PASSWORD@localhost:5432/ledger` |
| `MEMBERS_TEST_URL` | `http://localhost:8081` |
| `BFF_TEST_URL` | `http://localhost:8082` |
| `NOTIFY_TEST_URL` | `http://localhost:8083` |
| `LEDGER_TEST_URL` | `http://localhost:8084` |
| `BFF_TEST_EDGE` | `http://localhost:8090` |
| `BFF_TEST_REDIS` | `localhost:6379` |
| `NOTIFY_TEST_SMTP` | `localhost:1025` |
| `MEMBERS_TEST_BFF_CLIENT_SECRET` | `$BFF_CLIENT_SECRET` (optional: adds the wrong-audience test) |
| `NOTIFY_TEST_CLIENT_SECRET` | `$NOTIFY_CLIENT_SECRET` |
| `MEMBERS_TEST_CLIENT_SECRET` | `$MEMBERS_CLIENT_SECRET` |

Run them one package at a time, because several suites sign in as the one test Member:

```bash
(cd apps/backend && go test -p 1 ./...)
```

## Smoke tests

Both run against the running stack, and CI runs both in the `backend` job.

`login-token-smoke.sh` proves the email-link sign-in on the pinned Keycloak: a login token works once, the
refresh token is a 30-day offline one that works once, and log out ends it.

```bash
infra/compose/login-token-smoke.sh
```

`realm-noop.sh` fails if applying the realm again would change anything.

```bash
infra/compose/realm-noop.sh
```
