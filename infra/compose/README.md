# infra/compose

The phase-1 stack on a laptop (`docs/ARCHITECTURE.md` §10, A9). It runs Keycloak (Phase Two
image, pinned), Postgres, RabbitMQ, Redis, the `members` and `notify` migration steps, the OpenTofu
step that builds the `mmr` realm, the `members` service, `notify` and the BFF. Mailpit catches the
mail `notify` sends and is dev-only, behind a profile (see "Mailpit"). Nginx, `otel-lgtm` and
`pg-backup` come in later slices.

## Run it

Docker Desktop must be running.

```bash
cp infra/compose/.env.example infra/compose/.env
```

Fill every empty value in `.env` (any string will do locally; they are development credentials and
`.gitignore` keeps the file out of git). Then:

```bash
docker compose -f infra/compose/compose.yaml up -d --build --wait
```

`--build` builds the `members` and `bff` images from `apps/backend`; `--wait` returns once every container
reports healthy and the migration step has exited 0. Check with:

```bash
docker compose -f infra/compose/compose.yaml ps
```

Stop with `down`; add `-v` to drop the volumes. A `down` takes the profile the `up` took, so
`--profile mail down` is what also removes Mailpit.

**One last `down -v`.** Postgres runs `postgres/10-databases.sh` only on an empty data directory,
so the `notify` database and role it now creates never appear on an existing volume and
`notify-migrate` aborts the `up` with `database "notify" does not exist`. A stack created before the
realm moved to OpenTofu also still holds the JSON-imported realm in Keycloak's database, and the
first `tofu apply` fails creating a realm that is already there. Run
`docker compose -f infra/compose/compose.yaml down -v` once, then `up` as above. From that stack
onward a realm change is an `up` (see "The `mmr` realm").

**An existing `.env` needs the lines `.env.example` has gained.** `MEMBERS_DATABASE_URL` and
`NOTIFY_DATABASE_URL` are what Compose turns into the two connection-string secrets, and
`NOTIFY_DB_PASSWORD` is what the Postgres init script gives the `notify` role. Without the first two
Compose refuses the `up` with `environment variable "MEMBERS_DATABASE_URL" required by secret
"members_database_url" is not set` (see "Secrets"); without the third it refuses it with
`NOTIFY_DB_PASSWORD: set in .env`. `NOTIFY_HOST_PORT`, `MAILPIT_HOST_PORT` and `MAILPIT_SMTP_HOST_PORT` have
defaults and are needed only by a worktree running its own stack (see "Parallel worktrees").

## What comes up

The host ports are the `*_HOST_PORT` variables; their defaults are in `.env.example`.

| Container | Image | Host port | Notes |
| --- | --- | --- | --- |
| postgres | `postgres:17` | 127.0.0.1:`POSTGRES_HOST_PORT` | `postgres/10-databases.sh` runs once on a fresh volume: databases `keycloak`, `members` and `notify` — three of the five A9 plans — one role each, `CONNECT` revoked from `PUBLIC` so no role can reach another's database (A9). |
| keycloak | the `KEYCLOAK_TAG` default in `compose.yaml` (Keycloak 26.6.6 with magic-link 0.75) | `KEYCLOAK_HOST_PORT` | One full Phase Two tag; never `latest` or a shorter tag (§7.5). |
| rabbitmq | `rabbitmq:4-management` | — | No consumer until slice 5. |
| redis | `redis:8` | — | No client until slice 5. |
| members-migrate | built from `apps/backend` (`Dockerfile`, `CMD=members`) | — | One-shot: `members migrate` applies the embedded SQL migrations to the `members` database and exits. Reads its connection string from `/run/secrets/members_database_url`, as `members` does. |
| realm-apply | `ghcr.io/opentofu/opentofu`, pinned | — | One-shot: copies `infra/tofu/realm` into a named volume, runs `tofu apply` against Keycloak's admin API and exits. Starts after Keycloak is healthy. |
| notify-migrate | built from `apps/backend` (`Dockerfile`, `CMD=notify`) | — | One-shot: `notify migrate` applies the embedded SQL migrations to the `notify` database and exits. Reads its connection string from `/run/secrets/notify_database_url`, as `notify` does. |
| mailpit | `axllent/mailpit`, pinned | 127.0.0.1:`MAILPIT_HOST_PORT` and `MAILPIT_SMTP_HOST_PORT` | Dev-only, behind the `mail` profile, so a default `up` never starts it (see "Mailpit"). Takes SMTP on 1025 and shows what arrived on 8025; both are published to loopback, the second so a `notify` run on the host has a mail host too. |
| members | same image | 127.0.0.1:`MEMBERS_HOST_PORT` | `MembersService` over Connect, `/healthz` pings the database. Its connection string arrives as a file, never as an environment variable (see "Secrets"). Starts after the migration step and the realm step completed; verifies every bearer token against the realm JWKS (audience `bff`). |
| notify | built from `apps/backend` (`Dockerfile`, `CMD=notify`) | 127.0.0.1:`NOTIFY_HOST_PORT` | `NotifyService` over Connect, `/healthz` pings the database. Slice 3 serves `EmailMagicLink` alone, which is the one procedure outside the bearer check (§4.2 point 3); every other procedure answers `unauthenticated` without a token and arrives in slice 7. Sends through `NOTIFY_SMTP_ADDR`, `mailpit:1025` here. |
| bff | built from `apps/backend` (`Dockerfile`, `CMD=bff`) | `BFF_HOST_PORT` | `BffService` over Connect: what the phone calls. Listens on 8080 inside the network like every Go container (§10). Starts after `members` is healthy; verifies the token itself, then forwards it to `members` unchanged (§4.2). `ledger` is an in-process stub until slice 5, and so is `notify.GetWarningState` until slice 7; the `notify` container is real from slice 3 for `EmailMagicLink`, which the BFF calls from `RequestMagicLink` (#69). |

Every long-running container has a health check and `depends_on: condition: service_healthy` orders
the start; the two migration steps and the realm step are ordered with
`condition: service_completed_successfully`. Keycloak's check is a bash `/dev/tcp` request rather
than `curl`, because the Keycloak image ships neither `curl` nor `wget`: it opens 9000 for
`/health/ready` and then 8080 for `/realms/master`. It no longer asks for `/realms/mmr`, because
Keycloak now starts with no realm and `realm-apply` builds it afterwards. The second hop stays,
because `/health/ready` answers UP on the management port while a first boot is still bootstrapping
on 8080, and `realm-apply` then gets `503 Service Unavailable` from the token endpoint.
Keycloak and the BFF
publish on every interface because the phone reaches them; the Postgres, `members` and `notify`
ports bind to loopback and exist for the Connect-boundary tests (`apps/README.md`), as does
Mailpit's, which is where the magic-link test reads the message it caused.

## Secrets

`members` and `members-migrate` read their connection string from
`/run/secrets/members_database_url`, and `notify` and `notify-migrate` theirs from
`/run/secrets/notify_database_url`, not from the environment, so neither is in `docker inspect`
or the process list (`docs/ARCHITECTURE.md` A13). The shared reader takes any variable in the
`<NAME>_FILE` form, trims the file and falls back to the plain variable, so a file-backed variable is
never also reported missing.

The Compose secret's source differs by machine and the containers cannot tell:

| Where | Source | How |
| --- | --- | --- |
| A laptop and CI | `MEMBERS_DATABASE_URL` and `NOTIFY_DATABASE_URL` in `.env` | `secrets: members_database_url: {environment: ...}`. Docker materialises the file at mode 0444 with the exact value and no trailing newline, and does not export the variable into the container. |
| The VM | `infra/secrets/prod.yaml` | `infra/secrets/decrypt.sh` writes one file per value at mode 0400 before Compose starts. See [`../secrets/README.md`](../secrets/README.md). |

`.env.example` assembles each URL from `MEMBERS_DB_PASSWORD` and `NOTIFY_DB_PASSWORD` in the same
file, so a password still has one local source; the Postgres init script keeps reading the bare
passwords, because `CREATE ROLE` needs them by themselves.

## Parallel worktrees

Compose identifies a stack by project name, so two worktrees that both run this file as `mmr` recreate
each other's containers, and two Keycloaks cannot both bind 8080. Each worktree's `.env` therefore
sets its own `COMPOSE_PROJECT_NAME` and host ports (`KEYCLOAK_HOST_PORT`, `MEMBERS_HOST_PORT`,
`BFF_HOST_PORT`, `NOTIFY_HOST_PORT`, `POSTGRES_HOST_PORT`, `MAILPIT_HOST_PORT`,
`MAILPIT_SMTP_HOST_PORT`); the defaults in `.env.example` are the stock stack the iOS app signs in
against. `KC_HOSTNAME`, and with it the token issuer and the `members` issuer check, follow
`KEYCLOAK_HOST_PORT`, so a stack on port 18080 issues tokens for `http://localhost:18080/realms/mmr`.
Volumes and the `members` and `bff` images are named after the project, so stacks do not share data either.

## A phone on the LAN

A physical iPhone cannot reach the Mac as `localhost`, and the token issuer must match the address the
phone uses. `MMR_HOST` (default `localhost`) is the one place that address lives: `KC_HOSTNAME`, the
`members` issuer check and the BFF issuer check all derive from it. To sign in from a device:

1. Set `MMR_HOST=<the Mac's LAN IP>` in `.env` and recreate the stack (`up -d --wait`); Keycloak, `members`
   and `bff` restart with the new issuer `http://<ip>:${KEYCLOAK_HOST_PORT}/realms/mmr`. The realm is unaffected: it holds no address.
2. In Xcode, edit the scheme's Run action and add the environment variables
   `MMR_ISSUER=http://<ip>:${KEYCLOAK_HOST_PORT}/realms/mmr` and `MMR_BFF_URL=http://<ip>:${BFF_HOST_PORT}`; select
   a signing team, pick the phone, Run (`apps/ios/README.md`, "Run on a physical iPhone").
3. Set it back to `localhost` and recreate again before the simulator or the UI test next run.

## The `mmr` realm

`infra/tofu/realm` describes the realm and the `realm-apply` container applies it on every `up`
(A13). Keycloak starts with no realm at all. Editing a value under `infra/tofu/realm` and running
`up` is enough; there is no import to re-run and no `down -v` to pay.

- Realm `mmr`. Hostname is fixed to `http://${MMR_HOST}:${KEYCLOAK_HOST_PORT}`, so the issuer
  `http://localhost:8080/realms/mmr` (with the default port) is the same whether the caller is the
  simulator or a container reaching `keycloak:8080`. `ssl_required` is a variable rather than a
  default, `none` here, because every laptop caller speaks plain HTTP and the provider's default is
  `external`. Sign-in with an email address stays on, because the Login Token endpoint finds the
  Member by email and otherwise creates a second one.
- `ios`: public client, PKCE S256 required, direct access grants enabled for the test seams, redirect
  URI `mymoneyrecap://auth/callback` (custom scheme until universal links). An
  audience mapper adds `bff` to its access tokens. Its direct grant is the custom `ios direct grant`
  flow: two conditional branches on the optional `login-token` client scope. A token request with
  `scope=login-token` runs the Login Token authenticator (A10); one without runs the username and
  password validators. The two cannot be plain alternatives, because the Login Token authenticator
  fails the whole flow when the `login_token` field is absent (`a10-smoke.sh` proved this on the
  pinned image). Keycloak's built-in scopes (`profile`, `email`, `basic`, ...) come with the realm,
  so access tokens carry `sub`; the realm attribute the JSON import needed for that is gone.
- `bff`: confidential client with a service account holding `realm-management` `manage-users`, which
  the Login Token endpoint needs (A10). Its secret is a variable rather than a generated value,
  because `a10-smoke.sh` and the `members` wrong-audience test authenticate as this client.
- Test Member `test` / `test@example.com`.

`realm-apply` reads `BFF_CLIENT_SECRET` and `TEST_USER_PASSWORD` from `.env` as `TF_VAR_` variables
and re-applies both on every `up`, so a change to either takes effect at once. The realm's OpenTofu
state is a file in the `realm-state` volume, never in the working tree.
`KC_BOOTSTRAP_ADMIN_USERNAME` and `KC_BOOTSTRAP_ADMIN_PASSWORD` keep the old rule: they seed the
admin of an empty Keycloak database, so changing either reaches Keycloak only after `down -v`, and
until then `realm-apply` and the `kcadm` login under "Admin access" are refused. Nothing secret is
in a committed file.

## Try it

With the stock `.env`; a non-default `.env` substitutes its `KEYCLOAK_HOST_PORT` and `BFF_HOST_PORT`.
Discovery document from the host:

```bash
curl -s http://localhost:8080/realms/mmr/.well-known/openid-configuration | jq .issuer
```

A password grant for the test user on the `ios` client:

```bash
curl -s -X POST http://localhost:8080/realms/mmr/protocol/openid-connect/token -d grant_type=password -d client_id=ios -d username=test -d "password=$TEST_USER_PASSWORD" | jq -r .access_token
```

`GetHome` through the BFF with that token (the shape the app sees):

```bash
curl -s http://localhost:8082/money_recap.v1.BffService/GetHome -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{}'
```

## Mailpit

Mailpit is the SMTP host `notify` sends to on a laptop and in CI, so the client is exercised on every
run rather than first meeting a real provider on a VM. It is dev-only and sits behind the `mail`
profile, so a default `up` does not start it and `docker compose ps` does not list it:

```bash
docker compose -f infra/compose/compose.yaml --profile mail up -d --wait
```

Read what arrived at `http://localhost:8025` in a browser, or over its HTTP interface, which is what
the tests use:

```bash
curl -s "http://localhost:8025/api/v1/search?query=to:someone@example.test" | jq '.messages[0].ID'
curl -s http://localhost:8025/api/v1/message/<ID> | jq -r .Text
```

Send one without a token, the way the BFF will (`EmailMagicLink` is exempt from the bearer check by
procedure name, §4.2 point 3):

```bash
curl -s http://localhost:8083/money_recap.v1.NotifyService/EmailMagicLink -H 'Content-Type: application/json' \
  -d '{"idempotency_key":"'"$(uuidgen)"'","email":"someone@example.test","link":"https://mymoneyrecap.test/magic?lt=demo"}'
```

The same key twice sends one message and answers `{}` both times; an empty key is `invalid_argument`.

## Running the Go tests against this stack

The `internal/members` and `internal/notify` tests and `internal/bff`'s `TestEndToEnd` skip unless
their variables are set (`internal/members`' store-failure and caller-deadline tests,
`internal/notify`'s empty-key test and `internal/bff`'s other tests need nothing); with the stack up,
Mailpit started and the stock `.env` they are the values below. A non-default `.env`
substitutes its own `*_HOST_PORT`s. Either way the `members` database password is the value `.env`
held when the volumes were created; the test Member's password and the `bff` client secret are
whatever `.env` holds now, because `realm-apply` re-applies both on every `up`.

| Variable | Stock value | What it is |
| --- | --- | --- |
| `MEMBERS_TEST_DATABASE_URL` | `postgres://members:$MEMBERS_DB_PASSWORD@localhost:5432/members` | The `members` database on `POSTGRES_HOST_PORT`. Alone it enables the migration, concurrency and empty-email-claim tests. |
| `MEMBERS_TEST_URL` | `http://localhost:8081` | `members` on `MEMBERS_HOST_PORT`, for the Connect-boundary tests. |
| `MEMBERS_TEST_KEYCLOAK` | `http://localhost:8080` | Keycloak on `KEYCLOAK_HOST_PORT`, where the suite mints the test user's token. |
| `MEMBERS_TEST_USER_PASSWORD` | `$TEST_USER_PASSWORD` | The `test` user's password, for the password grant on the `ios` client. |
| `MEMBERS_TEST_BFF_CLIENT_SECRET` | `$BFF_CLIENT_SECRET` | **Optional.** Set it and the suite adds the wrong-audience case: a `bff` service-account token, which `members` must reject as unauthenticated. Unset, that assertion silently does not run. |
| `BFF_TEST_URL` | `http://localhost:8082` | The BFF on `BFF_HOST_PORT`, the HTTP boundary `TestEndToEnd` drives. |
| `BFF_TEST_KEYCLOAK` | `http://localhost:8080` | As `MEMBERS_TEST_KEYCLOAK`. |
| `BFF_TEST_USER_PASSWORD` | `$TEST_USER_PASSWORD` | As `MEMBERS_TEST_USER_PASSWORD`. |
| `BFF_TEST_DATABASE_URL` | `postgres://members:$MEMBERS_DB_PASSWORD@localhost:5432/members` | The same `members` database; the end-to-end test clears its own row before it runs. |
| `NOTIFY_TEST_URL` | `http://localhost:8083` | `notify` on `NOTIFY_HOST_PORT`, the Connect boundary the magic-link test drives with no bearer token at all. |
| `NOTIFY_TEST_MAILPIT` | `http://localhost:8025` | Mailpit's HTTP interface on `MAILPIT_HOST_PORT`, where that test reads the message it caused. |
| `NOTIFY_TEST_DATABASE_URL` | `postgres://notify:$NOTIFY_DB_PASSWORD@localhost:5432/notify` | The `notify` database, where the tests read back the `magic_link_send` row and clean it up. Alone it enables the in-process store tests, the migration-runner test and `cmd/notify`'s migrate case. |
| `NOTIFY_TEST_SMTP` | `localhost:1025` | Mailpit's SMTP port on `MAILPIT_SMTP_HOST_PORT`, for the in-process send. |

Run them with `-p 1`:

```bash
cd apps/backend && go test -p 1 ./...
```

**`-p 1` is load-bearing.** The realm has exactly one test user (`test`), and both `internal/members`
and `internal/bff` sign in as it and delete its `member` row. Run those two packages at once and they
race: one deletes the row the other just created, and the failures show up as row-count mismatches
and `not_found` errors far from the cause. CI passes `-p 1` for the same reason, and the step is named
after it.

## A10 smoke test

`a10-smoke.sh` proves A10 on the pinned image, and CI runs it on every pull request that touches the
backend, `infra/` or the contract (the `backend` job, after the Go tests, with a throwaway `.env`). Against the running stack, using the same `.env`, it:

1. gets a service-account token for `bff` and calls `POST /realms/mmr/login-token` for the test
   user's email with `expiration_seconds` 900, `reusable` false, `force_create` true and
   `set_email_verified` true;
2. redeems the `lt:` token on `ios` at the token endpoint with `grant_type=password`,
   `scope=login-token` and the `login_token` field, and asserts the access token's `sub` is the test
   user's id from the admin API;
3. redeems it again and asserts HTTP 401 `invalid_grant`;
4. checks the password grant on `ios` still works, and prints the image tag.

```bash
infra/compose/a10-smoke.sh
```

It reads `KEYCLOAK_HOST_PORT` from `.env` (see "Parallel worktrees").

## The realm re-applies as a no-op

`realm-noop.sh` runs `tofu plan` in the same state volume `realm-apply` writes and fails when the
plan is not empty, so a resource the provider would re-create on every `up` is caught. It fails the
same way when `infra/tofu/realm` has been edited since the last `up`, which is the check's other
use. CI runs it in the `backend` job after the A10 smoke test.

```bash
infra/compose/realm-noop.sh
```

## Admin access

The `mmr` realm has `sslRequired: none` so the simulator and a container can use plain HTTP. The
`master` realm keeps Keycloak's default (`external`), and Docker Desktop's host traffic does not
count as local, so the admin console at `http://localhost:8080` answers "HTTPS required" from the
host. A container on the Compose network does count as local, which is how `realm-apply` reaches
`master` over plain HTTP. Inspect the realm with `kcadm` inside the container instead:

```bash
docker compose -f infra/compose/compose.yaml exec keycloak bash -c '/opt/keycloak/bin/kcadm.sh config credentials --server http://127.0.0.1:8080 --realm master --user "$KC_BOOTSTRAP_ADMIN_USERNAME" --password "$KC_BOOTSTRAP_ADMIN_PASSWORD" && /opt/keycloak/bin/kcadm.sh get clients -r mmr --fields clientId,publicClient'
```

That login uses the container's current `KC_BOOTSTRAP_ADMIN_USERNAME` and
`KC_BOOTSTRAP_ADMIN_PASSWORD`, and fails once either differs from the value the volume was first
booted with (see "The `mmr` realm").

To use the admin console from a browser, relax `master` once per volume after that login:

```bash
docker compose -f infra/compose/compose.yaml exec keycloak /opt/keycloak/bin/kcadm.sh update realms/master -s sslRequired=none
```
