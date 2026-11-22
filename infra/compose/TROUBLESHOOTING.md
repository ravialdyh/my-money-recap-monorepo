# Troubleshooting the Compose stack

Fixes for what goes wrong most often with `infra/compose/compose.yaml`. Each one is a symptom, then the fix.

## A password change did nothing, or a step cannot sign in

Some values are used only when their volume is first created. Changing them later does nothing until the volume is
deleted:

- the Postgres passwords (`POSTGRES_PASSWORD`, `KEYCLOAK_DB_PASSWORD` and the three `*_DB_PASSWORD`);
- `RABBITMQ_PASSWORD`;
- `KC_BOOTSTRAP_ADMIN_PASSWORD`, the Keycloak admin.

On a laptop, delete the volumes and start again. This deletes the stack's data:

```bash
docker compose -f infra/compose/compose.yaml --profile mail down -v
```

`TEST_USER_PASSWORD` and the three client secrets are not affected ([README, "The `mmr` realm"](README.md#the-mmr-realm)).

## Every sign-in link fails

`notify` was started without `--profile mail`, so it has no mail host. Its log says
`lookup mailpit ... no such host`. Run the `up` command from the README, which carries the profile.

## A consumer fails with `PRECONDITION_FAILED`

A RabbitMQ queue keeps the settings it was first made with. After a change to the queue settings in
`apps/backend/internal/events`, delete the volumes with `down -v` as above.

## The `ledger` database is missing

The volume was made before `ledger` existed. The database script runs only on an empty volume, so `down -v`
fixes it.

## Go tests fail with `not_found` or wrong row counts

Two test packages ran at once and both signed in as the one test Member. Run them with `-p 1`:

```bash
(cd apps/backend && go test -p 1 ./...)
```

## The Keycloak admin console says "HTTPS required"

The `master` realm wants HTTPS from the host. Look at the realm from inside the container instead:

```bash
docker compose -f infra/compose/compose.yaml exec keycloak bash -c '/opt/keycloak/bin/kcadm.sh config credentials --server http://127.0.0.1:8080 --realm master --user "$KC_BOOTSTRAP_ADMIN_USERNAME" --password "$KC_BOOTSTRAP_ADMIN_PASSWORD" && /opt/keycloak/bin/kcadm.sh get clients -r mmr --fields clientId,publicClient'
```

To use the console in a browser, after that login, turn the check off once per volume:

```bash
docker compose -f infra/compose/compose.yaml exec keycloak /opt/keycloak/bin/kcadm.sh update realms/master -s sslRequired=none
```

## Docker runs out of memory

Keycloak's memory is `KEYCLOAK_JAVA_HEAP` in `.env`; `.env.example` sets `-Xms64m -Xmx512m`. Check that only one
stack is running:

```bash
docker compose ls
```
