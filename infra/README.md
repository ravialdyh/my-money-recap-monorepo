# infra

Deployment and infrastructure-as-code.

- `infra/compose/` — the phase-1 `compose.yaml`: Keycloak (Phase Two image, pinned) with the `mmr` realm bootstrapped from JSON, Postgres, RabbitMQ, Redis, the `members` migration step, `members` and the BFF. See [`compose/README.md`](compose/README.md) for how to start it.

Planned from slice 3 onward, per `docs/ARCHITECTURE.md` §5, §7.5 and §10:

- `infra/tofu/` — OpenTofu: DNS records and the apple-app-site-association file, the VM with cloud-init and firewall, the Keycloak realm (official `keycloak/keycloak` provider), and API tokens for providers that expose them.
- `infra/compose/` grows to the full §10 set: Nginx, the `acme.sh` sidecar, `ledger`, `notify`, `currency`, `otel-lgtm`, `pg-backup`.
- `infra/nginx/` — the Nginx configuration and the static `/magic` fallback page.
- `infra/pg-backup/` — the nightly `pg_dump` image.
- `infra/secrets/` — every secret, encrypted with SOPS and an age key (A13). Decrypted at deploy into `/run/secrets` and mounted as files; never plain text, never environment variables.
- `infra/wizard/` — the one-time script for the nine steps only a human can do (`docs/research/2026-09-14-setup-automation.md` §4); each step ends by pasting the value into `sops edit infra/secrets/prod.yaml`.
