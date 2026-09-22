# infra/compose

The phase-1 stack on a laptop (`docs/ARCHITECTURE.md` §10, A9). It runs Keycloak (Phase Two
image, pinned), Postgres, RabbitMQ, Redis, the `members` and `notify` migration steps, the OpenTofu
step that builds the `mmr` realm, the `members` service, `notify`, the BFF, Nginx and its `acme.sh`
sidecar. Mailpit catches the mail `notify` sends and is dev-only, behind the `mail` profile, which
every command here takes because `notify` has no other mail host on a laptop (see "Mailpit").
`otel-lgtm` and `pg-backup` come in later slices.

## Run it

Docker Desktop must be running.

```bash
cp infra/compose/.env.example infra/compose/.env
```

Fill every empty value in `.env` (any string will do locally; they are development credentials and
`.gitignore` keeps the file out of git). Then:

```bash
docker compose -f infra/compose/compose.yaml --profile mail up -d --build --wait
```

`--build` builds the `members`, `notify` and `bff` images from `apps/backend`; `--wait` returns once
every container reports healthy and the two migration steps have exited 0. Check with:

```bash
docker compose -f infra/compose/compose.yaml --profile mail ps
```

**`--profile mail` is part of the command, not an extra.** `notify` is configured to send through
`mailpit:1025` and has no second mail host, but its health check is the one every Go container
shares and pings only the database. An `up` without the profile therefore returns a `notify` that
reports healthy and answers every send with `unavailable`, so every `RequestMagicLink` the BFF
sends it fails with that code (#69). Leaving the profile out is a way to see what an `up` looks like without
a mail host, never a way to run the stack (see "Mailpit").

Stop with `--profile mail down`; add `-v` to drop the volumes. A `down` takes the profile the `up`
took, so a plain `down` leaves Mailpit running.

**One last `down -v`.** Postgres runs `postgres/10-databases.sh` only on an empty data directory,
so the `notify` database and role it now creates never appear on an existing volume and
`notify-migrate` aborts the `up` with `database "notify" does not exist`. A stack created before the
realm moved to OpenTofu also still holds the JSON-imported realm in Keycloak's database, and the
first `tofu apply` fails creating a realm that is already there. Run
`docker compose -f infra/compose/compose.yaml --profile mail down -v` once, then `up` as above. From
that stack onward a realm change is an `up` (see "The `mmr` realm").

**An existing `.env` needs the lines `.env.example` has gained.** `MEMBERS_DATABASE_URL` and
`NOTIFY_DATABASE_URL` are what Compose turns into the two connection-string secrets, and
`NOTIFY_DB_PASSWORD` is what the Postgres init script gives the `notify` role. Without the first two
Compose refuses the `up` with `environment variable "MEMBERS_DATABASE_URL" required by secret
"members_database_url" is not set` (see "Secrets"); without the third it refuses it with
`NOTIFY_DB_PASSWORD: set in .env`. `NOTIFY_HOST_PORT`, `MAILPIT_HOST_PORT`,
`MAILPIT_SMTP_HOST_PORT`, `REDIS_HOST_PORT` and `NGINX_HOST_PORT` have defaults and are needed only
by a worktree running its own stack (see "Parallel worktrees"). `KC_BOOTSTRAP_ADMIN_USERNAME` is no
longer read: the admin username is the literal `admin` in `compose.yaml`, so the line an older `.env`
carries is ignored, and a stack first booted with any other name must be rebuilt with `down -v`.

## What comes up

The host ports are the `*_HOST_PORT` variables; their defaults are in `.env.example`.

| Container | Image | Host port | Notes |
| --- | --- | --- | --- |
| postgres | `postgres:17` | 127.0.0.1:`POSTGRES_HOST_PORT` | `postgres/10-databases.sh` runs once on a fresh volume: databases `keycloak`, `members` and `notify` — three of the five A9 plans — one role each, `CONNECT` revoked from `PUBLIC` so no role can reach another's database (A9). |
| keycloak | the `KEYCLOAK_TAG` default in `compose.yaml` (Keycloak 26.6.6 with magic-link 0.75) | `KEYCLOAK_HOST_PORT` | One full Phase Two tag; never `latest` or a shorter tag (§7.5). Its file vault reads the two identity-provider secrets from `/run/secrets` (see "Secrets"). |
| rabbitmq | `rabbitmq:4-management` | — | No consumer until slice 5. |
| redis | `redis:8` | 127.0.0.1:`REDIS_HOST_PORT` | The BFF's magic-link limit, §4.4's fifth job and the first one built: five requests per address per hour, counted with `INCR` and `EXPIRE`, failing closed (#69). Published to loopback so a `bff` run on the host has a Redis too. The other four §4.4 jobs arrive with slices 4 and 5. |
| members-migrate | built from `apps/backend` (`Dockerfile`, `CMD=members`) | — | One-shot: `members migrate` applies the embedded SQL migrations to the `members` database and exits. Reads its connection string from `/run/secrets/members_database_url`, as `members` does. |
| realm-apply | `ghcr.io/opentofu/opentofu`, pinned | — | One-shot: copies `infra/tofu/realm` into a named volume, runs `tofu apply` against Keycloak's admin API and exits. Starts after Keycloak is healthy. |
| notify-migrate | built from `apps/backend` (`Dockerfile`, `CMD=notify`) | — | One-shot: `notify migrate` applies the embedded SQL migrations to the `notify` database and exits. Reads its connection string from `/run/secrets/notify_database_url`, as `notify` does. |
| mailpit | `axllent/mailpit`, pinned | 127.0.0.1:`MAILPIT_HOST_PORT` and `MAILPIT_SMTP_HOST_PORT` | Dev-only, behind the `mail` profile, so only an `up` carrying `--profile mail` starts it — which is why every command in "Run it" carries it (see "Mailpit"). Takes SMTP on 1025 and shows what arrived on 8025; both are published to loopback, the second so a `notify` run on the host has a mail host too. |
| members | same image | 127.0.0.1:`MEMBERS_HOST_PORT` | `MembersService` over Connect, `/healthz` pings the database. Its connection string arrives as a file, never as an environment variable (see "Secrets"). Starts after the migration step and the realm step completed; verifies every bearer token against the realm JWKS (audience `bff`). |
| notify | built from `apps/backend` (`Dockerfile`, `CMD=notify`) | 127.0.0.1:`NOTIFY_HOST_PORT` | `NotifyService` over Connect, `/healthz` pings the database. It serves `EmailMagicLink` alone (#68), which is the one procedure outside the bearer check (§4.2 point 3); every other procedure answers `unauthenticated` without a token and arrives in slice 7. Sends through `NOTIFY_SMTP_ADDR`, `mailpit:1025` here, so it is healthy but cannot send unless the `up` carried `--profile mail`. |
| bff | built from `apps/backend` (`Dockerfile`, `CMD=bff`) | `BFF_HOST_PORT` | `BffService` over Connect: what the phone calls. Listens on 8080 inside the network like every Go container (§10). Starts after `members` and Redis are healthy and the realm step completed; verifies the token itself, then forwards it to `members` unchanged (§4.2). `ledger` is an in-process stub until slice 5, and so is `notify.GetWarningState` until slice 7; the `notify` container is real for `EmailMagicLink` since #68, and the BFF calls it from `RequestMagicLink` (#69). |
| nginx | `nginx:stable` | `NGINX_HOST_PORT` | The edge (§5, A7): the four routes, the per-IP limits, the `/magic` fallback page and the apple-app-site-association file. One mounted configuration file and one mounted directory of static files, both read-only; see [`../nginx/README.md`](../nginx/README.md). Plain HTTP on a laptop, because ACME needs a domain. |
| acme | `neilpang/acme.sh`, pinned | — | The certificate sidecar (§5, §10). Runs the image's `daemon` command and shares the `acme-certs` volume with Nginx; with no domain registered it renews nothing, which is all it does on a laptop. |

Every long-running container has a health check and `depends_on: condition: service_healthy` orders
the start; the two migration steps and the realm step are ordered with
`condition: service_completed_successfully`. Keycloak's check is a bash `/dev/tcp` request rather
than `curl`, because the Keycloak image ships neither `curl` nor `wget`: it opens 9000 for
`/health/ready` and then 8080 for `/realms/master`. It no longer asks for `/realms/mmr`, because
Keycloak now starts with no realm and `realm-apply` builds it afterwards. The second hop stays,
because `/health/ready` answers UP on the management port while a first boot is still bootstrapping
on 8080, and `realm-apply` then gets `503 Service Unavailable` from the token endpoint.
Keycloak, the BFF and Nginx
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
| The machine | `infra/secrets/prod.yaml` | The deploy decrypts it on the runner with `infra/secrets/decrypt.sh`, and `machine-env.sh` writes the values into the machine's `.env`, so the same `environment:` source serves both. See [`../README.md`](../README.md), "Deploying". |

`.env.example` assembles each URL from `MEMBERS_DB_PASSWORD` and `NOTIFY_DB_PASSWORD` in the same
file, so a password still has one local source; the Postgres init script keeps reading the bare
passwords, because `CREATE ROLE` needs them by themselves.

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
`MAILPIT_HOST_PORT`, `MAILPIT_SMTP_HOST_PORT`, `NGINX_HOST_PORT`); the defaults in `.env.example` are the stock
stack the iOS app signs in against. `KC_HOSTNAME`, and with it the token issuer and the `members` issuer check, follow
`KEYCLOAK_HOST_PORT`, so a stack on port 18080 issues tokens for `http://localhost:18080/realms/mmr`.
Volumes and the `members`, `notify` and `bff` images are named after the project, so stacks do not share data either.

## A phone on the LAN

A physical iPhone cannot reach the Mac as `localhost`, and the token issuer must match the address the
phone uses. `MMR_HOST` (default `localhost`) is the one place that address lives: `KC_HOSTNAME`, the
`members` issuer check and the BFF issuer check all derive from it. To sign in from a device:

1. Set `MMR_HOST=<the Mac's LAN IP>` in `.env` and recreate the stack (`--profile mail up -d --wait`); Keycloak, `members`
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
  `external`. `REALM_SSL_REQUIRED` overrides it, and only the machine's `.env` sets it, to `external`. Sign-in with an email address stays on, because the Login Token endpoint finds the
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
- Test Member `test` / `test@example.com`, created only when `TEST_USER_PASSWORD` is non-empty. The
  machine's `.env` carries no such line, so no account with a known password exists there. On a
  laptop an empty value is refused by the realm root itself, which requires the password whenever
  `ssl_required` is `none`, so `realm-apply` exits 1 with `TEST_USER_PASSWORD is empty` and the
  `up` fails. The test Member's address in the OpenTofu state is `keycloak_user.test[0]`, and a
  `moved` block carries a stack built before that change across without re-creating the test Member.
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
and re-applies both on every `up`, so a change to either takes effect at once. The realm's OpenTofu
state is a file in the `realm-state` volume, never in the working tree.
`KC_BOOTSTRAP_ADMIN_PASSWORD` keeps the old rule, and so does the literal `admin` username beside it:
they seed the admin of an empty Keycloak database, so changing either reaches Keycloak only after
`down -v`, and until then `realm-apply` and the `kcadm` login under "Admin access" are refused. Nothing secret is
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

## The edge

Nginx is the one container in front of the BFF and Keycloak (§5, A7). On a laptop it serves plain
HTTP on `NGINX_HOST_PORT`, and [`../nginx/README.md`](../nginx/README.md) holds its four routes, its
two static files and its limits. Keycloak and the BFF keep their own published ports, and the app's
configuration, every test variable and `a10-smoke.sh` still use them, so nothing a caller follows
moved when Nginx arrived. The one value that did is `BFF_MAGIC_LINK_BASE_URL`: the link in the email
is now `http://$MMR_HOST:$NGINX_HOST_PORT/magic?lt=...`, so tapping it on a laptop reaches the
fallback page instead of a BFF that serves no such path.

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

Mailpit is the SMTP host `notify` sends to on a laptop and in CI, so the client is exercised on every
run rather than first meeting a real provider on a VM. It is dev-only and sits behind the `mail`
profile (#68), so it is not one of the thirteen §10 containers on any machine, a default `up` does
not start it and a plain `docker compose ps` does not list it:

```bash
docker compose -f infra/compose/compose.yaml --profile mail up -d --wait
```

The profile is what a command asks for, never what a container waits on: `notify` starts and reports
healthy with or without it, because its check is the shared one that pings only the database, and
making it dial SMTP instead would fail `--wait` on the machine every time a real provider blinks.
So an `up` that forgets `--profile mail` gives a `notify` that answers every send `unavailable` with
`dial tcp: lookup mailpit ... no such host` in its log. That is the whole of the trap, and the
commands in "Run it" carry the profile so that following them cannot reach it. CI passes the same
profile to its `up` and its `down` (`.github/workflows/ci.yml`, the `backend` job).

Read what arrived at `http://localhost:8025` in a browser, or over its HTTP interface, which is what
the tests use:

```bash
curl -s "http://localhost:8025/api/v1/search?query=to:someone@example.test" | jq '.messages[0].ID'
curl -s http://localhost:8025/api/v1/message/<ID> | jq -r .Text
```

Send one without a token, the way the BFF does (#69; `EmailMagicLink` is exempt from the bearer
check by procedure name, §4.2 point 3):

```bash
curl -s http://localhost:8083/money_recap.v1.NotifyService/EmailMagicLink -H 'Content-Type: application/json' \
  -d '{"idempotency_key":"'"$(uuidgen)"'","email":"someone@example.test","link":"https://mymoneyrecap.test/magic?lt=demo"}'
```

The same key twice sends one message and answers `{}` both times; an empty key is `invalid_argument`.

## Running the Go tests against this stack

The `internal/members`, `internal/notify` and `internal/migrate` tests and `internal/bff`'s
`TestEndToEnd`, `TestMagicLinkEndToEnd` and `TestEdge` skip unless their variables are set (`internal/members`' store-failure and
caller-deadline tests, `internal/notify`'s empty-key test and `internal/bff`'s other tests need
nothing); with the stack up, Mailpit started and the stock `.env` they are the values below. A
non-default `.env` substitutes its own `*_HOST_PORT`s. Either way the `members` database password is
the value `.env` held when the volumes were created; the test Member's password and the `bff` client secret are
whatever `.env` holds now, because `realm-apply` re-applies both on every `up`.

| Variable | Stock value | What it is |
| --- | --- | --- |
| `MEMBERS_TEST_DATABASE_URL` | `postgres://members:$MEMBERS_DB_PASSWORD@localhost:5432/members` | The `members` database on `POSTGRES_HOST_PORT`. Alone it enables the migration, concurrency and empty-email-claim tests, and the shared migration runner's own tests, which take this or `NOTIFY_TEST_DATABASE_URL`, whichever is set, and work in a throwaway schema they drop. |
| `MEMBERS_TEST_URL` | `http://localhost:8081` | `members` on `MEMBERS_HOST_PORT`, for the Connect-boundary tests. |
| `MEMBERS_TEST_KEYCLOAK` | `http://localhost:8080` | Keycloak on `KEYCLOAK_HOST_PORT`, where the suite mints the test user's token. |
| `MEMBERS_TEST_USER_PASSWORD` | `$TEST_USER_PASSWORD` | The `test` user's password, for the password grant on the `ios` client. |
| `MEMBERS_TEST_BFF_CLIENT_SECRET` | `$BFF_CLIENT_SECRET` | **Optional.** Set it and the suite adds the wrong-audience case: a `bff` service-account token, which `members` must reject as unauthenticated. Unset, that assertion silently does not run. |
| `BFF_TEST_URL` | `http://localhost:8082` | The BFF on `BFF_HOST_PORT`, the HTTP boundary `TestEndToEnd` drives. |
| `BFF_TEST_KEYCLOAK` | `http://localhost:8080` | As `MEMBERS_TEST_KEYCLOAK`. |
| `BFF_TEST_USER_PASSWORD` | `$TEST_USER_PASSWORD` | As `MEMBERS_TEST_USER_PASSWORD`. |
| `BFF_TEST_DATABASE_URL` | `postgres://members:$MEMBERS_DB_PASSWORD@localhost:5432/members` | The same `members` database; `TestEndToEnd` clears its own row before it runs, and `TestMagicLinkEndToEnd` counts and then drops the row its fresh address makes. |
| `BFF_TEST_MAILPIT` | `http://localhost:8025` | Mailpit's HTTP interface on `MAILPIT_HOST_PORT`, where `TestMagicLinkEndToEnd` reads the message it caused and pulls the login token out of the link. With `BFF_TEST_URL`, `BFF_TEST_KEYCLOAK` and `BFF_TEST_DATABASE_URL` it is that test's whole set; it needs no `BFF_TEST_USER_PASSWORD`, because it spends a fresh address and never signs in as `test`. |
| `BFF_TEST_EDGE` | `http://localhost:8090` | Nginx on `NGINX_HOST_PORT`, the edge `TestEdge` drives. With `BFF_TEST_URL`, `BFF_TEST_KEYCLOAK` and `BFF_TEST_USER_PASSWORD` it is that test's whole set: it compares one `GetHome` through the edge with the same call straight to the BFF, fetches the `/magic` page and the association file, and proves the stricter limit zone matches the `RequestMagicLink` path and nothing else (#72). |
| `NOTIFY_TEST_URL` | `http://localhost:8083` | `notify` on `NOTIFY_HOST_PORT`, the Connect boundary the magic-link test drives with no bearer token at all. |
| `NOTIFY_TEST_MAILPIT` | `http://localhost:8025` | Mailpit's HTTP interface on `MAILPIT_HOST_PORT`, where that test reads the message it caused. |
| `NOTIFY_TEST_DATABASE_URL` | `postgres://notify:$NOTIFY_DB_PASSWORD@localhost:5432/notify` | The `notify` database, where the tests read back the `magic_link_send` row and clean it up. Alone it enables the in-process store tests, `cmd/notify`'s migrate case and, when `MEMBERS_TEST_DATABASE_URL` is unset, the shared migration runner's tests. |
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
