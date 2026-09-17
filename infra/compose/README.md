# infra/compose

The phase-1 stack on a laptop (`docs/ARCHITECTURE.md` §10, A9). Slice 1 runs Keycloak (Phase Two
image, pinned), Postgres, RabbitMQ, Redis, the `members` migration step, the `members` service and
the BFF. Nginx, `otel-lgtm` and `pg-backup` come in later slices.

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

Stop with `down`; add `-v` to drop the volumes, which is also how you get a fresh realm import
(see below).

## What comes up

The host ports are the `*_HOST_PORT` variables; their defaults are in `.env.example`.

| Container | Image | Host port | Notes |
| --- | --- | --- | --- |
| postgres | `postgres:17` | 127.0.0.1:`POSTGRES_HOST_PORT` | `postgres/10-databases.sh` runs once on a fresh volume: databases `keycloak` and `members`, one role each, `CONNECT` revoked from `PUBLIC` so neither role can reach the other's database (A9). |
| keycloak | the `KEYCLOAK_TAG` default in `compose.yaml` (Keycloak 26.6.6 with magic-link 0.75) | `KEYCLOAK_HOST_PORT` | One full Phase Two tag; never `latest` or a shorter tag (§7.5). |
| rabbitmq | `rabbitmq:4-management` | — | No consumer until slice 5. |
| redis | `redis:8` | — | No client until slice 5. |
| members-migrate | built from `apps/backend` (`Dockerfile`, `CMD=members`) | — | One-shot: `members migrate` applies the embedded SQL migrations to the `members` database and exits. |
| members | same image | 127.0.0.1:`MEMBERS_HOST_PORT` | `MembersService` over Connect, `/healthz` pings the database. Starts after the migration step completed and Keycloak is healthy; verifies every bearer token against the realm JWKS (audience `bff`). |
| bff | built from `apps/backend` (`Dockerfile`, `CMD=bff`) | `BFF_HOST_PORT` | `BffService` over Connect: what the phone calls. Listens on 8080 inside the network like every Go container (§10). Starts after `members` is healthy; verifies the token itself, then forwards it to `members` unchanged (§4.2). `ledger` and `notify` are in-process stubs until slices 5 and 7. |

Every long-running container has a health check and `depends_on: condition: service_healthy` orders
the start; the migration step is ordered with `condition: service_completed_successfully`. Keycloak
and the BFF publish on every interface because the phone reaches them; the Postgres and `members`
ports bind to loopback and exist for the Connect-boundary tests (`apps/README.md`).

## Parallel worktrees

Compose identifies a stack by project name, so two worktrees that both run this file as `mmr` recreate
each other's containers, and two Keycloaks cannot both bind 8080. Each worktree's `.env` therefore
sets its own `COMPOSE_PROJECT_NAME` and host ports (`KEYCLOAK_HOST_PORT`, `MEMBERS_HOST_PORT`,
`BFF_HOST_PORT`, `POSTGRES_HOST_PORT`); the defaults in `.env.example` are the stock stack the iOS app signs in
against. `KC_HOSTNAME`, and with it the token issuer and the `members` issuer check, follow
`KEYCLOAK_HOST_PORT`, so a stack on port 18080 issues tokens for `http://localhost:18080/realms/mmr`.
Volumes and the `members` and `bff` images are named after the project, so stacks do not share data either.

## A phone on the LAN

A physical iPhone cannot reach the Mac as `localhost`, and the token issuer must match the address the
phone uses. `MMR_HOST` (default `localhost`) is the one place that address lives: `KC_HOSTNAME`, the
`members` issuer check and the BFF issuer check all derive from it. To sign in from a device:

1. Set `MMR_HOST=<the Mac's LAN IP>` in `.env` and recreate the stack (`up -d --wait`); Keycloak, `members`
   and `bff` restart with the new issuer `http://<ip>:${KEYCLOAK_HOST_PORT}/realms/mmr`. The realm import is unaffected.
2. In Xcode, edit the scheme's Run action and add the environment variables
   `MMR_ISSUER=http://<ip>:${KEYCLOAK_HOST_PORT}/realms/mmr` and `MMR_BFF_URL=http://<ip>:${BFF_HOST_PORT}`; select
   a signing team, pick the phone, Run (`apps/ios/README.md`, "Run on a physical iPhone").
3. Set it back to `localhost` and recreate again before the simulator or the UI test next run.

## The `mmr` realm

`keycloak/mmr-realm.json` is imported by `--import-realm` on first boot. Keycloak skips the import
once the realm exists in its database, so a change to the JSON needs `down -v` to take effect. Slice 3
replaces this bootstrap with OpenTofu (A13).

- Realm `mmr`. Hostname is fixed to `http://${MMR_HOST}:${KEYCLOAK_HOST_PORT}`, so the issuer
  `http://localhost:8080/realms/mmr` (with the default port) is the same whether the caller is the
  simulator or a container reaching `keycloak:8080`.
- `ios`: public client, PKCE S256 required, direct access grants enabled for the test seams, redirect
  URI `mymoneyrecap://auth/callback` (custom scheme until slice 3 serves universal links). An
  audience mapper adds `bff` to its access tokens. Its direct grant is the custom `ios direct grant`
  flow: two conditional branches on the optional `login-token` client scope. A token request with
  `scope=login-token` runs the Login Token authenticator (A10); one without runs the username and
  password validators. The two cannot be plain alternatives, because the Login Token authenticator
  fails the whole flow when the `login_token` field is absent (`a10-smoke.sh` proved this on the
  pinned image). The realm attribute `CreateDefaultClientScopes` keeps Keycloak's built-in scopes
  (`profile`, `email`, `basic`, ...) when the JSON declares its own `clientScopes`; without it the
  imported realm has none and access tokens carry no `sub`.
- `bff`: confidential client with a service account holding `realm-management` `manage-users`, which
  the Login Token endpoint needs (A10). Secret from `BFF_CLIENT_SECRET`.
- Test user `test` / `test@example.com`, password from `TEST_USER_PASSWORD`.

The two `${VARIABLE}` placeholders in the JSON are filled from the container environment, which
Compose reads from `.env`. Nothing secret is in a committed file.

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

## A10 smoke test

`a10-smoke.sh` proves A10 on the pinned image, and CI runs it on every pull request (`a10-smoke` job,
throwaway `.env`, only the `keycloak` service and its `postgres` dependency started). Against the running stack, using the same `.env`, it:

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

## Admin access

The `mmr` realm has `sslRequired: none` so the simulator and a container can use plain HTTP. The
`master` realm keeps Keycloak's default (`external`), and Docker Desktop's host traffic does not
count as local, so the admin console at `http://localhost:8080` answers "HTTPS required" from the
host. Inspect the realm with `kcadm` inside the container instead:

```bash
docker compose -f infra/compose/compose.yaml exec keycloak bash -c '/opt/keycloak/bin/kcadm.sh config credentials --server http://127.0.0.1:8080 --realm master --user "$KC_BOOTSTRAP_ADMIN_USERNAME" --password "$KC_BOOTSTRAP_ADMIN_PASSWORD" && /opt/keycloak/bin/kcadm.sh get clients -r mmr --fields clientId,publicClient'
```

To use the admin console from a browser, relax `master` once per volume after that login:

```bash
docker compose -f infra/compose/compose.yaml exec keycloak /opt/keycloak/bin/kcadm.sh update realms/master -s sslRequired=none
```
