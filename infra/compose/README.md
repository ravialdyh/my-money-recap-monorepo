# infra/compose

The phase-1 stack (`docs/ARCHITECTURE.md` §10, A9): one `compose.yaml` that a laptop, CI and the
machine all run. "What comes up" lists its containers. Every command here carries `--profile mail`,
because Mailpit, behind that profile, is the only mail host `notify` has (see "Mailpit").

## Run it

Docker Desktop must be running.

```bash
cp infra/compose/.env.example infra/compose/.env
```

Fill every empty value in `.env`: any string will do locally, except that `MEMBERS_DB_PASSWORD` and
`NOTIFY_DB_PASSWORD` go into a connection URL unescaped, so use letters and digits for those. They are
development credentials, and `.gitignore` keeps the file out of git. Then:

```bash
docker compose -f infra/compose/compose.yaml --profile mail up -d --build --wait
```

`--build` builds the `members`, `notify` and `bff` images from `apps/backend`; `--wait` returns once
every container reports healthy and the two migration steps and `realm-apply` have exited 0. Check
with:

```bash
docker compose -f infra/compose/compose.yaml --profile mail ps
```

**`--profile mail` is part of the command, not an extra:** without it `notify` reports healthy and
cannot send, so every `RequestMagicLink` fails (see "Mailpit").

Stop with `--profile mail down`; add `-v` to drop the volumes. A `down` takes the profile the `up`
took, so a plain `down` leaves Mailpit running.

## What comes up

The host ports are the `*_HOST_PORT` variables, whose defaults are in `compose.yaml`.

| Container | Image | Host port | Notes |
| --- | --- | --- | --- |
| postgres | `postgres:17` | 127.0.0.1:`POSTGRES_HOST_PORT` | `postgres/10-databases.sh` runs once on a fresh volume: databases `keycloak`, `members` and `notify` — three of the five A9 plans — one role each, `CONNECT` revoked from `PUBLIC` so no role can reach another's database (A9). |
| keycloak | `quay.io/phasetwo/phasetwo-keycloak`, one full Phase Two tag pinned in `compose.yaml` | `KEYCLOAK_HOST_PORT` | Never `latest` or a shorter tag (§7.5). Its file vault reads the two identity-provider secrets from `/run/secrets` (see "Secrets"). |
| redis | `redis:8` | 127.0.0.1:`REDIS_HOST_PORT` | The BFF's magic-link rate limit, the one §4.4 job built (`apps/backend/cmd/bff/README.md`); the JWKS is cached in each process's memory (`internal/auth`). |
| rabbitmq | `rabbitmq:4-management` | — | The broker (§4.3), running so the outbox has somewhere to publish; nothing publishes to it, consumes from it or waits on it yet. It runs as its own `rabbitmq` user: any RabbitMQ command run as root before the server has written its Erlang cookie, such as the health check on a slow first start, wrote that file as root, and the server then died reading it (`eacces`, #93). Its AMQP and management ports stay inside the stack. |
| realm-apply | `ghcr.io/opentofu/opentofu`, pinned | — | One-shot: copies `infra/tofu/realm` into a named volume, runs `tofu apply` against Keycloak's admin API and exits. Starts after Keycloak is healthy (see "The `mmr` realm"). |
| members-migrate | the `members` image, `migrate` command | — | One-shot: applies the `members` migrations and exits, reading its connection string from `/run/secrets/members_database_url` as `members` does. |
| notify-migrate | the `notify` image, `migrate` command | — | The same for `notify`. |
| mailpit | `axllent/mailpit`, pinned | 127.0.0.1:`MAILPIT_HOST_PORT` and `MAILPIT_SMTP_HOST_PORT` | Behind the `mail` profile: the mail host `notify` sends to (see "Mailpit"). SMTP on 1025, what arrived on 8025. |
| members | built from `apps/backend` (`Dockerfile`, `CMD=members`) | 127.0.0.1:`MEMBERS_HOST_PORT` | `MembersService` over Connect (`apps/backend/cmd/members/README.md`). Starts after its migration step and the realm step completed. |
| notify | built from `apps/backend` (`Dockerfile`, `CMD=notify`) | 127.0.0.1:`NOTIFY_HOST_PORT` | `NotifyService` over Connect (`apps/backend/cmd/notify/README.md`). Starts after its migration step and the realm step completed. |
| bff | built from `apps/backend` (`Dockerfile`, `CMD=bff`) | `BFF_HOST_PORT` | `BffService` over Connect, what the phone calls (`apps/backend/cmd/bff/README.md`). Starts after `members` and Redis are healthy and the realm step completed. |
| nginx | `nginx:stable` | `NGINX_HOST_PORT` | The edge (§5, A7), plain HTTP until a certificate exists (see "The edge"). Nginx resolves `bff` and `keycloak` once, at start, so Compose restarts it whenever an `up` recreates either (`restart: true`); otherwise `/api/` and `/auth/` could answer 502 from an old address. |
| acme | `neilpang/acme.sh`, pinned | — | The certificate sidecar (§5, §10), sharing the `acme-certs` volume with Nginx; what it does with no domain is in [`../nginx/README.md`](../nginx/README.md#tls-and-the-acmesh-sidecar). |

Every long-running container but the `acme` sidecar has a health check, and
`depends_on: condition: service_healthy` orders the start; the two migration steps and the realm step
are ordered with `condition: service_completed_successfully`. Keycloak's check is a bash `/dev/tcp`
request rather than `curl`, because the Keycloak image ships neither `curl` nor `wget`: it opens 9000
for `/health/ready` and then 8080 for `/realms/master`, not `/realms/mmr`, which `realm-apply` builds
only once Keycloak is healthy. The second hop is there because `/health/ready` answers UP on the
management port while a first boot is still bootstrapping on 8080, and `realm-apply` would then get
`503 Service Unavailable` from the token endpoint.

Keycloak, the BFF and Nginx publish on every interface because the phone reaches them. The Postgres,
Redis, `members`, `notify` and Mailpit ports bind to loopback, for the Go tests ("Running the Go tests
against this stack") and for a service run on the host by hand. RabbitMQ publishes none.

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
| The machine | `infra/secrets/prod.yaml` | The deploy decrypts it on the runner with `infra/secrets/decrypt.sh`, and `machine-env.sh` writes the values into the machine's `.env`, so the same `environment:` source serves both. See [`../README.md`](../README.md), "Deploying". |

`.env.example` assembles each URL from `MEMBERS_DB_PASSWORD` and `NOTIFY_DB_PASSWORD` in the same
file, so a password still has one local source; the Postgres init script keeps reading the bare
passwords, because `CREATE ROLE` needs them by themselves.

**Chosen once.** Postgres runs `postgres/10-databases.sh` only on an empty data directory, and
`KC_BOOTSTRAP_ADMIN_PASSWORD`, with the literal `admin` username beside it in `compose.yaml`, seeds
the admin of an empty Keycloak database. A change to that script, to a database password or to the
admin password therefore reaches the stack only after `down -v`; until then the step or service using
the old value is refused, `realm-apply` and the `kcadm` login under "Admin access" included.
`BFF_CLIENT_SECRET` and `TEST_USER_PASSWORD` are not chosen once: `realm-apply` re-applies both on
every `up` (see "The `mmr` realm").

Keycloak takes the Google client secret and the Apple `.p8` through its file vault (`KC_VAULT=file`,
`KC_VAULT_DIR=/run/secrets`, the `KEY_ONLY` key resolver), so the realm holds only
`${vault.google_client_secret}` and `${vault.apple_sign_in_key}` (#71). Their laptop and CI source
is `vault/`, two committed files holding `replace-me`, which needs no `.env` line; Compose mounts
each under the name Keycloak looks for, with every underscore doubled. Why, and why the Apple
extension never asks the vault at all, is in
[`../secrets/README.md`](../secrets/README.md#keycloaks-file-vault).

## Parallel worktrees

Compose identifies a stack by project name, so two worktrees that both run this file as `mmr` recreate
each other's containers, and two Keycloaks cannot both bind 8080. Each worktree's `.env` therefore
sets its own `COMPOSE_PROJECT_NAME` and host ports (`KEYCLOAK_HOST_PORT`, `MEMBERS_HOST_PORT`,
`BFF_HOST_PORT`, `NOTIFY_HOST_PORT`, `POSTGRES_HOST_PORT`, `REDIS_HOST_PORT`,
`MAILPIT_HOST_PORT`, `MAILPIT_SMTP_HOST_PORT`, `NGINX_HOST_PORT`); the defaults in `compose.yaml` are the stock
stack the iOS app signs in against. `KC_HOSTNAME`, and with it the token issuer and the issuer checks in
`members`, `notify` and the BFF, follow `KEYCLOAK_HOST_PORT`, so a stack on port 18080 issues tokens for `http://localhost:18080/realms/mmr`.
Volumes and the `members`, `notify` and `bff` images are named after the project, so stacks do not share data either.

## A phone on the LAN

A physical iPhone cannot reach the Mac as `localhost`, and the token issuer must match the address the
phone uses. `MMR_HOST` (default `localhost`) is the one place that address lives: `KC_HOSTNAME`, the issuer
check in `members`, `notify` and the BFF, and the BFF's magic-link base address all derive from it. To sign in from a device:

1. Set `MMR_HOST=<the Mac's LAN IP>` in `.env` and recreate the stack (`--profile mail up -d --wait`); Keycloak, `members`,
   `notify` and `bff` restart with the new issuer `http://<ip>:${KEYCLOAK_HOST_PORT}/realms/mmr`. The realm is unaffected: it holds no address.
2. In Xcode, edit the scheme's Run action and add the environment variables
   `MMR_ISSUER=http://<ip>:${KEYCLOAK_HOST_PORT}/realms/mmr` and `MMR_BFF_URL=http://<ip>:${BFF_HOST_PORT}`; select
   a signing team, pick the phone, Run (`apps/ios/README.md`, "Run on a physical iPhone").
3. Set it back to `localhost` and recreate again before the simulator or the UI test next run.

## The `mmr` realm

`infra/tofu/realm` describes the realm and the `realm-apply` container applies it on every `up`
(A13). Keycloak starts with no realm at all, and editing a value under `infra/tofu/realm` and
running `up` is enough to change it.

- Realm `mmr`. Hostname is fixed to `http://${MMR_HOST}:${KEYCLOAK_HOST_PORT}`, so the issuer
  `http://localhost:8080/realms/mmr` (with the default port) is the same whether the caller is the
  simulator or a container reaching `keycloak:8080`. `ssl_required` is a variable rather than a
  default, `none` here, because every laptop caller speaks plain HTTP and the provider's default is
  `external`. `REALM_SSL_REQUIRED` overrides it, and only the machine's `.env` sets it, to `external`. Sign-in with an email address stays on, because the Login Token endpoint finds the
  Member by email and otherwise creates a second one.
- `ios`: public client, PKCE S256 required, direct access grants enabled for the test seams, redirect
  URI `mymoneyrecap://auth/callback` (custom scheme until universal links). An
  audience mapper adds `bff` to its access tokens. Every sign-in asks for `offline_access`, so its
  refresh token is an offline one the realm keeps 30 days from last use, with no overall cap (PRD D62).
  Each refresh token works once: the realm refuses a reused one and ends that sign-in. Its direct grant is the custom `ios direct grant`
  flow: two conditional branches on the optional `login-token` client scope. A token request with
  `scope=login-token` runs the Login Token authenticator (A10); one without runs the username and
  password validators. The two cannot be plain alternatives, because the Login Token authenticator
  fails the whole flow when the `login_token` field is absent (`a10-smoke.sh` proved this on the
  pinned image). Keycloak's built-in scopes (`profile`, `email`, `basic`, ...) come with the realm,
  so access tokens carry `sub`.
- `bff`: confidential client with a service account holding `realm-management` `manage-users`, which
  the Login Token endpoint needs (A10). Its secret is a variable rather than a generated value,
  because the BFF signs in as this client for that endpoint (`BFF_CLIENT_SECRET`), and so do
  `a10-smoke.sh` and the `members` wrong-audience test.
- Test Member `test` / `test@example.com`, created only when `TEST_USER_PASSWORD` is non-empty. The
  machine's `.env` carries no such line, so no account with a known password exists there. On a
  laptop an empty value is refused by the realm root itself, which requires the password whenever
  `ssl_required` is `none`, so `realm-apply` exits 1 with `TEST_USER_PASSWORD is empty` and the
  `up` fails.
- Identity providers `google` (Keycloak's built-in Google provider, "Store tokens" off) and
  `apple` (the image's Apple extension through the generic OIDC resource with
  `provider_id = "apple"`, the Team ID and Key ID in `extra_config`), both with "Trust email" on
  (PRD §9). The app sends `kc_idp_hint` with one of those two aliases, so Keycloak skips its own
  login page and redirects straight to the provider (#71). Their client ids and the Apple Team ID
  and Key ID are `GOOGLE_CLIENT_ID`, `APPLE_SERVICES_ID`, `APPLE_TEAM_ID` and
  `APPLE_SIGN_IN_KEY_ID`, which `.env` may set and nothing requires: unset, each is `replace-me`,
  so the providers exist, the buttons reach them, and the provider answers with its own error
  page. Their secrets are vault references (see "Secrets").

`realm-apply` reads `BFF_CLIENT_SECRET` and `TEST_USER_PASSWORD` from `.env` as `TF_VAR_` variables
and re-applies both on every `up`, so a change to either takes effect at once; the admin password it
signs in with is chosen once (see "Secrets"). The realm's OpenTofu state is a file in the
`realm-state` volume, never in the working tree. Nothing secret is in a committed file.

## Try it

With the stock `.env`; a non-default `.env` substitutes its `KEYCLOAK_HOST_PORT` and `BFF_HOST_PORT`.
Discovery document from the host:

```bash
curl -s http://localhost:8080/realms/mmr/.well-known/openid-configuration | jq .issuer
```

A password grant for the test Member on the `ios` client:

```bash
curl -s -X POST http://localhost:8080/realms/mmr/protocol/openid-connect/token -d grant_type=password -d client_id=ios -d username=test -d "password=$TEST_USER_PASSWORD" | jq -r .access_token
```

`GetHome` through the BFF with that token (the shape the app sees):

```bash
curl -s http://localhost:8082/money_recap.v1.BffService/GetHome -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{}'
```

## The edge

Nginx is the one container in front of the BFF and Keycloak (§5, A7). It serves plain HTTP on
`NGINX_HOST_PORT` until a certificate exists, and [`../nginx/README.md`](../nginx/README.md) holds its
four routes, its two static files, its limits and its TLS sidecar. Keycloak and the BFF also keep
their own published ports, which the app's configuration, `a10-smoke.sh` and every test variable but
`BFF_TEST_EDGE` use. The link in the email is `http://$MMR_HOST:$NGINX_HOST_PORT/magic?lt=...`
(`BFF_MAGIC_LINK_BASE_URL`), so tapping it on a laptop reaches the fallback page.

The same call through the edge and straight to the BFF, which answer the same thing:

```bash
curl -s http://localhost:8090/api/money_recap.v1.BffService/GetHome -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{}'
curl -s http://localhost:8082/money_recap.v1.BffService/GetHome -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{}'
```

The two static files take no token at all:

```bash
curl -si http://localhost:8090/.well-known/apple-app-site-association | grep -i '^HTTP/\|^Content-Type'
curl -s http://localhost:8090/magic | grep -o 'Open My Money Recap[^<]*'
```

## Mailpit

Mailpit is the only SMTP host `notify` has: `compose.yaml` sets `NOTIFY_SMTP_ADDR` to `mailpit:1025`
and the sender to `no-reply@mymoneyrecap.test` for every stack, the machine's included. On a laptop
and in CI that exercises the client on every run rather than first meeting a real provider on a VM.
Mailpit sits behind the `mail` profile (#68), so it is not one of §10's thirteen containers, a
default `up` does not start it and a plain `docker compose ps` does not list it. The deploy's `up`
carries no profile, so the machine's `notify` cannot send until #87 item 8 gives it a provider
([`../README.md`](../README.md), "Not in the repository yet").

```bash
docker compose -f infra/compose/compose.yaml --profile mail up -d --wait
```

The profile is what a command asks for, never what a container waits on: `notify` starts and reports
healthy with or without it, because its check is the shared one that pings only the database, and
making it dial SMTP instead would fail `--wait` on the machine every time a real provider blinks.
So an `up` that forgets `--profile mail` gives a `notify` that answers every send `unavailable` with
`dial tcp: lookup mailpit ... no such host` in its log, and every `RequestMagicLink` fails with that
code, because the BFF's `EmailMagicLink` call does (#69). The commands in "Run it" carry the profile
so that following them cannot reach this. CI passes the same profile to its `up`, its `ps` and its
`down` (`.github/workflows/ci.yml`, the `backend` job).

Read what arrived at `http://localhost:8025` in a browser, or over its HTTP interface, which is what
the tests use:

```bash
curl -s "http://localhost:8025/api/v1/search?query=to:someone@example.test" | jq '.messages[0].ID'
curl -s http://localhost:8025/api/v1/message/<ID> | jq -r .Text
```

Send one without a token, the way the BFF does (`apps/backend/cmd/notify/README.md` says why it may):

```bash
curl -s http://localhost:8083/money_recap.v1.NotifyService/EmailMagicLink -H 'Content-Type: application/json' \
  -d '{"idempotency_key":"'"$(uuidgen)"'","email":"someone@example.test","link":"https://mymoneyrecap.test/magic?lt=demo"}'
```

The same key twice sends one message and answers `{}` both times; an empty key is `invalid_argument`.

## Running the Go tests against this stack

The tests that reach the stack skip unless their variables are set; every other test needs nothing.
With the stack up, Mailpit started and the stock `.env` they are the values below. A
non-default `.env` substitutes its own `*_HOST_PORT`s. Either way the `members` database password is
the value `.env` held when the volumes were created (see "Secrets"); the test Member's password and the `bff` client secret are
whatever `.env` holds now, because `realm-apply` re-applies both on every `up`. The `MMR_TEST_*` names
are the facts more than one suite reads; the others belong to one suite.

| Variable | Stock value | What it is |
| --- | --- | --- |
| `MMR_TEST_MEMBERS_DATABASE_URL` | `postgres://members:$MEMBERS_DB_PASSWORD@localhost:5432/members` | The `members` database on `POSTGRES_HOST_PORT`. Alone it enables the migration, concurrency and empty-email-claim tests, and the shared migration runner's own tests, which take this or `NOTIFY_TEST_DATABASE_URL`, whichever is set, and work in a throwaway schema they drop. `TestEndToEnd` clears its own row before it runs, and `TestMagicLinkEndToEnd` counts and then drops the row its fresh address makes. |
| `MMR_TEST_KEYCLOAK` | `http://localhost:8080` | Keycloak on `KEYCLOAK_HOST_PORT`, where the suites mint the test Member's token and `TestMagicLinkEndToEnd` redeems its login token. |
| `MMR_TEST_USER_PASSWORD` | `$TEST_USER_PASSWORD` | The test Member's password, for the password grant on the `ios` client. |
| `MMR_TEST_MAILPIT` | `http://localhost:8025` | Mailpit's HTTP interface on `MAILPIT_HOST_PORT`, where the `notify` tests and `TestMagicLinkEndToEnd` read the message they caused; the last pulls the login token out of the link. |
| `MEMBERS_TEST_URL` | `http://localhost:8081` | `members` on `MEMBERS_HOST_PORT`, for the Connect-boundary test, with `MMR_TEST_KEYCLOAK` and `MMR_TEST_USER_PASSWORD`. |
| `MEMBERS_TEST_BFF_CLIENT_SECRET` | `$BFF_CLIENT_SECRET` | **Optional.** Set it and the suite adds the wrong-audience case: a `bff` service-account token, which `members` must reject as unauthenticated. Unset, that assertion silently does not run. |
| `BFF_TEST_URL` | `http://localhost:8082` | The BFF on `BFF_HOST_PORT`, the HTTP boundary the three end-to-end tests drive. `TestEndToEnd` adds `MMR_TEST_KEYCLOAK`, `MMR_TEST_USER_PASSWORD` and `MMR_TEST_MEMBERS_DATABASE_URL`; `TestMagicLinkEndToEnd` takes `MMR_TEST_MAILPIT` in place of the password, because it spends a fresh address and never signs in as `test`. |
| `BFF_TEST_EDGE` | `http://localhost:8090` | Nginx on `NGINX_HOST_PORT`, the edge `TestEdge` drives. With `BFF_TEST_URL`, `MMR_TEST_KEYCLOAK` and `MMR_TEST_USER_PASSWORD` it is that test's whole set: it compares one `GetHome` through the edge with the same call straight to the BFF, fetches the `/magic` page and the association file, and proves the stricter limit zone matches the `RequestMagicLink` path and nothing else (#72). |
| `NOTIFY_TEST_URL` | `http://localhost:8083` | `notify` on `NOTIFY_HOST_PORT`, the Connect boundary the magic-link test drives with no bearer token at all. |
| `NOTIFY_TEST_DATABASE_URL` | `postgres://notify:$NOTIFY_DB_PASSWORD@localhost:5432/notify` | The `notify` database, where the tests read back the `magic_link_send` row and clean it up. Alone it enables the in-process store tests and, when `MMR_TEST_MEMBERS_DATABASE_URL` is unset, the shared migration runner's tests. |
| `NOTIFY_TEST_SMTP` | `localhost:1025` | Mailpit's SMTP port on `MAILPIT_SMTP_HOST_PORT`, for the in-process send. |

Run them with `-p 1`:

```bash
cd apps/backend && go test -p 1 ./...
```

**`-p 1` is load-bearing.** The realm has exactly one test Member (`test`), and both `internal/members`
and `internal/bff` sign in as it and delete its `member` row. Run those two packages at once and they
race: one deletes the row the other just created, and the failures show up as row-count mismatches
and `not_found` errors far from the cause. CI passes `-p 1` for the same reason, and the step is named
after it.

## A10 smoke test

`a10-smoke.sh` proves A10 on the pinned image, and CI runs it on every pull request that touches the
backend, `infra/` or the contract (the `backend` job, after the Go tests, with a throwaway `.env`). Against the running stack, using the same `.env`, it:

1. gets a service-account token for `bff` and calls `POST /realms/mmr/login-token` for the test
   Member's email with `expiration_seconds` 900, `reusable` false, `force_create` true and
   `set_email_verified` true;
2. redeems the `lt:` token on `ios` at the token endpoint with `grant_type=password`,
   `scope=login-token offline_access` and the `login_token` field, as the app does, and asserts the
   access token's `sub` is the test Member's id from the admin API and the refresh token is an
   offline one (PRD D62's 30 days);
3. redeems it again and asserts HTTP 401 `invalid_grant`;
4. checks the password grant on `ios` still works, and that a refresh token already used once is
   refused (one use per token, PRD §9);
5. posts the offline refresh token to the logout endpoint and asserts a refresh then fails, as Log
   out will (PRD §9), and prints the image tag.

```bash
infra/compose/a10-smoke.sh
```

It reads `KEYCLOAK_HOST_PORT` from `.env` when set (see "Parallel worktrees").

## The realm re-applies as a no-op

`realm-noop.sh` runs `tofu plan` in the same state volume `realm-apply` writes and fails when the
plan is not empty, so a resource the provider would re-create on every `up` is caught. It fails the
same way when `infra/tofu/realm` has been edited since the last `up`, which is the check's other
use. CI runs it in the `backend` job after the A10 smoke test.

```bash
infra/compose/realm-noop.sh
```

## Admin access

The `mmr` realm has `ssl_required` `none` (`REALM_SSL_REQUIRED`) so the simulator and a container
can use plain HTTP. The
`master` realm keeps Keycloak's default (`external`), and Docker Desktop's host traffic does not
count as local, so the admin console at `http://localhost:8080` answers "HTTPS required" from the
host. A container on the Compose network does count as local, which is how `realm-apply` reaches
`master` over plain HTTP. Inspect the realm with `kcadm` inside the container instead:

```bash
docker compose -f infra/compose/compose.yaml exec keycloak bash -c '/opt/keycloak/bin/kcadm.sh config credentials --server http://127.0.0.1:8080 --realm master --user "$KC_BOOTSTRAP_ADMIN_USERNAME" --password "$KC_BOOTSTRAP_ADMIN_PASSWORD" && /opt/keycloak/bin/kcadm.sh get clients -r mmr --fields clientId,publicClient'
```

That login fails once the container's `KC_BOOTSTRAP_ADMIN_PASSWORD` differs from the one the volume
was first booted with (see "Secrets").

To use the admin console from a browser, relax `master` once per volume after that login:

```bash
docker compose -f infra/compose/compose.yaml exec keycloak /opt/keycloak/bin/kcadm.sh update realms/master -s sslRequired=none
```
