# infra

Deployment and infrastructure-as-code.

- `infra/compose/` — the phase-1 `compose.yaml`: Keycloak (Phase Two image, pinned), Postgres, RabbitMQ, Redis, the `members` and `notify` migration steps, the OpenTofu step that builds the `mmr` realm, `members`, `notify`, the BFF, Nginx with its `acme.sh` sidecar, and Mailpit behind a dev-only profile. See [`compose/README.md`](compose/README.md) for how to start it.
- `infra/tofu/realm/` — the `mmr` realm in OpenTofu (official `keycloak/keycloak` provider, A13): the realm, the `ios` and `bff` clients, the `login-token` scope, the custom `ios direct grant` flow and the test Member. The Compose stack applies it on every `up`; the lock file covers `linux_amd64`, `linux_arm64` and `darwin_arm64`, which are CI, the one-shot container and a Mac running pre-commit.
- `infra/secrets/` — every secret, encrypted with SOPS and an age key (A13). Decrypted at deploy into one directory as one file per value, mounted with Compose `secrets:`; never plain text, never environment variables. See [`secrets/README.md`](secrets/README.md). This slice commits the shape: the key set, the decrypt step and its proof.
- `infra/nginx/` — the edge (§5, A7): one configuration file with the four routes and the per-IP limits, the static `/magic` fallback page and the apple-app-site-association file. Compose mounts both read-only into `nginx:stable`, beside an `acme.sh` sidecar that is written and inert until there is a domain. See [`nginx/README.md`](nginx/README.md).
- `infra/wizard/` — the one-time script for the nine steps only a human can do (`docs/research/2026-09-14-setup-automation.md` §4), with a non-interactive check mode that reports which are done. See [`wizard/README.md`](wizard/README.md).

Planned from slice 3 onward, per `docs/ARCHITECTURE.md` §5, §7.5 and §10:

- `infra/tofu/` — a second root for the machine: DNS records and the apple-app-site-association file, the VM with cloud-init and firewall, and API tokens for providers that expose them. Separate state from the realm root, written and never applied.
- `infra/compose/` grows to the full §10 set: `ledger`, `currency`, `otel-lgtm`, `pg-backup`.
- `infra/pg-backup/` — the nightly `pg_dump` image.
