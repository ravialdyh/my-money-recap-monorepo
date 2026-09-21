# infra

Deployment and infrastructure-as-code.

- `infra/compose/` — the phase-1 `compose.yaml`: Keycloak (Phase Two image, pinned), Postgres, RabbitMQ, Redis, the `members` and `notify` migration steps, the OpenTofu step that builds the `mmr` realm, `members`, `notify`, the BFF, Nginx with its `acme.sh` sidecar, and Mailpit behind a dev-only profile. See [`compose/README.md`](compose/README.md) for how to start it.
- `infra/tofu/realm/` — the `mmr` realm in OpenTofu (official `keycloak/keycloak` provider, A13): the realm, the `ios` and `bff` clients, the `login-token` scope, the custom `ios direct grant` flow, the test Member, and the Google and Apple identity providers with their secrets as Keycloak vault references. The Compose stack applies it on every `up`; the lock file covers `linux_amd64`, `linux_arm64` and `darwin_arm64`, which are CI, the one-shot container and a Mac running pre-commit.
- `infra/secrets/` — every secret, encrypted with SOPS and an age key (A13). Decrypted on the deploy runner into one directory as one file per value; the machine receives its `.env` and Keycloak's two vault files from it, never the age key (see "Deploying"). See [`secrets/README.md`](secrets/README.md). This slice commits the shape: the key set, the decrypt step and its proof.
- `infra/nginx/` — the edge (§5, A7): one configuration file with the four routes and the per-IP limits, the static `/magic` fallback page and the apple-app-site-association file. Compose mounts both read-only into `nginx:stable`, beside an `acme.sh` sidecar that is written and inert until there is a domain. See [`nginx/README.md`](nginx/README.md).
- `.github/workflows/deploy.yml` — the deploy (A15), written and never run; see "Deploying" below.
- `infra/wizard/` — the one-time script for the nine steps only a human can do (`docs/research/2026-09-14-setup-automation.md` §4), with a non-interactive check mode that reports which are done. See [`wizard/README.md`](wizard/README.md).

Planned from slice 3 onward, per `docs/ARCHITECTURE.md` §5, §7.5 and §10:

- `infra/tofu/` — a second root for the machine: DNS records and the apple-app-site-association file, the VM with cloud-init and firewall, and API tokens for providers that expose them. Separate state from the realm root, written and never applied.
- `infra/compose/` grows to the full §10 set: `ledger`, `currency`, `otel-lgtm`, `pg-backup`.
- `infra/pg-backup/` — the nightly `pg_dump` image.

## Deploying

`.github/workflows/deploy.yml` is the only way anything deploys (A15). It is written, `actionlint` is
its only gate, and it has never run: the first tag push is the follow-up issue's, not #73's.

**Trigger.** A tag matching `v[0-9]+.[0-9]+.[0-9]+`: `v1.2.3` and `v10.20.30`, never `v1.2`,
`v1.2.3-rc.1` or a branch name. The `push` trigger names `tags` and nothing else, so GitHub runs it
for no branch push, `main` included, and there is no `workflow_dispatch`. One job, in the GitHub
`production` environment. A second tag waits for the first rather than cancelling it halfway.

**The one hand-held secret.** `SOPS_AGE_KEY`, the whole age key file, is the environment's only
secret. Everything else comes out of `infra/secrets/prod.yaml`. The registry login uses the job's own
`GITHUB_TOKEN`, which GitHub issues per run and nobody holds.

**The steps, in order.**

1. Check out the repository. The encrypted file is in it, so nothing can be decrypted earlier, and a
   checkout copies nothing to the machine.
2. Decrypt `prod.yaml` on the runner with `infra/secrets/decrypt.sh`. The pinned, checksummed `sops`
   is installed inside this same step, so no step sits between the checkout and the decrypt, and
   every decrypted line is masked in the log.
3. Write the SSH client configuration from three decided values: the host is the `domain` value, the
   user is the literal `ubuntu` (the cloud-init default of Ubuntu's images), and the host-key policy
   is `accept-new`. The key is `deploy_ssh_private_key`, written with a final newline whether or not
   the value ends in one; OpenSSH refuses a key without it as `invalid format`. **`accept-new` checks
   nothing here**: a hosted runner starts with an empty `known_hosts` every run, so every deploy trusts
   whichever host answers at the domain and then sends it the machine's `.env`. Pinning the machine's
   host key through cloud-init, and checking it strictly here, belongs to the machine root (#74).
4. Build `members`, `notify` and `bff` from the one `apps/backend/Dockerfile`, one `CMD` build
   argument each, and push them to GHCR as `ghcr.io/<owner>/<repository>/<service>:<tag>`. GHCR
   needs no account beyond the repository's own.
5. Copy to `~/mmr` on the machine, replacing what the last deploy left: `infra/compose/compose.yaml`
   and everything it bind-mounts relative to itself, in the same layout — `infra/compose/postgres`,
   `infra/tofu/realm`, `infra/nginx/default.conf`, `infra/nginx/html` — then the machine's `.env`,
   mode 0600, written by `infra/compose/machine-env.sh` from the decrypted values, and Keycloak's two
   vault files, `google_client_secret` and `apple_sign_in_key`, into `infra/compose/vault/` where the
   laptop keeps its placeholders. That directory is mode 0700 and the two files 0644: Keycloak runs as
   uid 2000 in the Phase Two image and `ubuntu` is uid 1000, so a file only its owner can read would be
   unreadable to Keycloak, while the directory keeps every other account on the machine out.
6. Log the machine in to GHCR with the job's token, pull the three images, tag them `mmr-members`,
   `mmr-notify` and `mmr-bff` (the names Compose gives the images it builds on a laptop), log out,
   and `docker compose up -d --wait`. Compose finds its images present and builds nothing; the
   `build:` sections stay for the laptop. `realm-apply` builds the realm and the migration steps run
   in the order `service_completed_successfully` already gives them. `nginx` is restarted last,
   because a running container keeps serving the files it was started with after the copy replaces
   them.

**The machine's `.env`** is the laptop's with four differences: `MMR_HOST` is the domain,
`REALM_SSL_REQUIRED=external` sets the realm's `ssl_required` (spec #61: never the laptop's `none`),
`NGINX_HOST_PORT=80`, and there is no `TEST_USER_PASSWORD`, so the realm creates no test Member.
The eight stack values in `infra/secrets/example.yaml`, and the four identity-provider ids the realm
takes as variables (`GOOGLE_CLIENT_ID`, `APPLE_SERVICES_ID`, `APPLE_TEAM_ID`,
`APPLE_SIGN_IN_KEY_ID`), arrive under their names in capitals, so none is left at `replace-me`. Every
value is single-quoted, so Compose interpolates nothing inside it, and a value holding a quote or a
line break stops the deploy instead.

**What A13 keeps and does not.** Every secret is encrypted in one committed file, the age key is the
only one a human holds, and `bff_client_secret` rotates by edit, commit, tag, because `realm-apply`
writes it into the realm on every `up`. `members` and `notify` still read their connection strings
from files Compose writes under `/run/secrets`, and Keycloak reads the two identity-provider secrets
from its file vault. But the other values reach Compose through a `.env` on the machine's disk rather
than through files decrypted under `/run/secrets`, and Postgres, Keycloak and the BFF receive theirs as
environment variables, exactly as on a laptop. Chosen by the owner on 2026-09-22 over spec #61's
files-only design: Keycloak has no way to read its own database or admin password from a file, anyone
who can run `docker inspect` on the machine can already read any file there, and the machine never
holds the age key.

**Chosen once, not rotated by a tag.** The other seven stack values are fixed when their volume is
first initialised: `postgres/10-databases.sh` sets the three database passwords on an empty Postgres
(the two connection strings carry two of them), and `kc_bootstrap_admin_password` seeds the admin of
an empty Keycloak database. A tag that changes one leaves the running role or admin on the old
value, so the migration steps, Keycloak or `realm-apply` fail to sign in, and a changed
`postgres_password` changes nothing. Change one by hand on the machine first, then in `prod.yaml`;
`down -v`, the laptop's answer, deletes the data there.

**Not in the repository yet.** `infra/secrets/prod.yaml` is git-ignored and does not exist, and the
workflow reads it anyway: committing it is the follow-up issue's first item. Nor do the `production`
environment and its `SOPS_AGE_KEY` exist yet, or the machine, which needs Docker, its Compose plugin
and the deploy key's public half from cloud-init.
