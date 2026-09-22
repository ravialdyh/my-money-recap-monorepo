# infra/secrets

Every secret, encrypted with SOPS and an age key (`docs/ARCHITECTURE.md` A13, §7.5). One age private
key is the only secret a human ever holds; everything else is ciphertext in git. The deploy decrypts
it on the runner; what it sends the machine, and how Compose hands each value on, is in
[`../README.md`](../README.md), "Deploying".

## What is here, and what is not

| Path | Committed | What it is |
| --- | --- | --- |
| `example.yaml` | yes | The key set with placeholder values, the convention `.env.example` already follows. It is what the wizard checks presence against, so the two files cannot drift. |
| `decrypt.sh` | yes | The deploy's decrypt step: one file per value, mode 0400. |
| `decrypt-smoke.sh` | yes | Its proof, on a throwaway age key it generates itself. |
| `keyset.sh` | yes | The one reader of `example.yaml`'s shape: `keys` lists the key set, `value` prints the placeholder one key carries, and `stack` names the eight values the stack itself needs. The smoke, the wizard and `machine-env.sh` all go through it, so a change to the file's shape cannot leave them disagreeing about the key set. |
| `prod.yaml` | **no** | The real values. Git-ignored until it holds real values; committing it then, with `.sops.yaml`, is the first item of #87. |
| `.sops.yaml`, at the **repository root** | **no** | The creation rule naming the age recipient. It sits at the root because sops resolves the rule by walking up from the working directory, not from the file it is given, so a copy in this directory is invisible to `sops edit infra/secrets/prod.yaml` run from the root. Git-ignored until `prod.yaml` is committed, and it could not be committed sooner: a rule with an invalid recipient makes `sops edit` fail outright rather than degrade, so it waits for a real key. |

No age key has been generated yet, so the repository holds the shape and nothing else (#66, PR
#76). The wizard's
first step is `age-keygen`, which writes the new public key into `.sops.yaml` and creates `prod.yaml`
from `example.yaml`.

## The key set

Seven of the nine human steps (`docs/research/2026-09-14-setup-automation.md` §4) put a value here;
Apple Developer Program enrolment and `age-keygen` produce none. Which step fills which key, and where
its value comes from, is `STEPS` in [`../wizard/setup.sh`](../wizard/setup.sh). The last eight keys,
`keyset.sh stack`, come from no step: six passwords chosen once and the two connection strings
assembled from two of them. They are what the stack itself needs, and the wizard's check mode weighs
them alongside the steps.

| Key | Consumed by |
| --- | --- |
| `app_store_connect_api_key` | The TestFlight half of the tag deploy (A15), which #87 still owes |
| `apple_services_id` | The Keycloak Apple identity provider's client id, an OpenTofu variable (#71) |
| `apple_team_id` | The same provider's `teamId`, the issuer of the client-secret JWT it signs |
| `apple_sign_in_key_id` | The same provider's `keyId`, the JWT's `kid` |
| `apple_sign_in_key` | The same provider, which signs the client secret on every token exchange; referenced as `${vault.apple_sign_in_key}`, which the pinned extension does not resolve (see "Keycloak's file vault") |
| `apple_push_key` | `notify`'s push sender, planned |
| `google_client_id` | The Keycloak Google identity provider's client id, an OpenTofu variable (#71) |
| `google_client_secret` | The same provider, through Keycloak's file vault as `${vault.google_client_secret}` |
| `open_exchange_rates_app_id` | The planned `currency` service's rate fetch |
| `smtp_token` | `notify`'s SMTP login on the machine, owed by #87 (`notify` sends no AUTH yet); Mailpit needs none locally |
| `domain` | The deploy workflow's SSH host, and the machine's `MMR_HOST` |
| `registrar_token` | The DNS records of the machine root (#74), not written yet |
| `cloud_api_token` | The VM, firewall and cloud-init of the machine root (#74), not written yet |
| `deploy_ssh_private_key` | The deploy workflow's transport to the machine; the machine root (#74), not written yet, is to hand the public half to cloud-init |
| `members_db_password` | The Postgres init script, which needs the password by itself to create the role |
| `members_database_url` | `members` and `members migrate`, as `MEMBERS_DATABASE_URL_FILE` |
| `notify_db_password` | The same init script, which creates the `notify` role the same way |
| `notify_database_url` | `notify` and `notify migrate`, as `NOTIFY_DATABASE_URL_FILE` |
| `postgres_password` | The Postgres superuser, `POSTGRES_PASSWORD` |
| `keycloak_db_password` | The init script's `keycloak` role and Keycloak itself, `KEYCLOAK_DB_PASSWORD` |
| `kc_bootstrap_admin_password` | Keycloak's first admin, whose username is the literal `admin`, and `realm-apply`'s login |
| `bff_client_secret` | The `bff` client in the realm and the BFF, `BFF_CLIENT_SECRET` |

Each pair holds the same password twice on purpose: `members` and `notify` each read one whole
connection string rather than a password of their own, and `postgres/10-databases.sh` reads the bare
password. Keep a pair in step when either half changes.

## Decrypting

```bash
SOPS_AGE_KEY_FILE=~/.config/sops/age/keys.txt infra/secrets/decrypt.sh infra/secrets/prod.yaml infra/secrets/decrypted
```

Each top-level key becomes a file of that name holding the value with no trailing newline, mode 0400
in a directory the umask leaves at 0700. The deploy workflow runs this on the runner as the step right
after the checkout, before anything is copied, because the transport credential is inside the file. It
reads the age key from the `SOPS_AGE_KEY` secret of the GitHub `production` environment, which holds
the whole key file; sops carries age inside it, so no `age` binary is needed there. The directory
itself stays on the runner, and the machine never holds the age key.

`infra/secrets/decrypted/` is the one path inside the repository that is safe to point it at, and
`.gitignore` covers it. The plaintext is what needs guarding and no scanner guards it: these files
are bare values whose names carry no keyword context, so nothing would flag one that got committed.

## Keycloak's file vault

The realm never holds the two identity-provider secrets, only a reference to each:
`${vault.google_client_secret}` and `${vault.apple_sign_in_key}` (#71). Keycloak resolves a reference
when it needs the value, from a file in its vault directory, `/run/secrets` inside the container,
which Compose fills with `secrets:`. Two details of Keycloak 26.6 decide the file names, both proved
on the pinned image:

- **The key resolver is `KEY_ONLY`.** The default, `REALM_UNDERSCORE_KEY`, looks for
  `mmr_google__client__secret`, which nothing writes.
- **Every resolver doubles an underscore**, so even `KEY_ONLY` looks for `google__client__secret`. A
  file under the bare key is found only through a deprecated fallback that logs a warning on every
  lookup. Compose therefore mounts each file under the doubled name (`target:` in
  `infra/compose/compose.yaml`), and `decrypt.sh` keeps naming files after the bare key.

On a laptop there is no decrypted directory, so the two sources are `infra/compose/vault/`, committed
files holding the same `replace-me` placeholders as `example.yaml`. On the machine the deploy writes
the two decrypted values into that same directory instead (`../README.md`, "Deploying").

**Google's secret resolves; Apple's does not.** Keycloak's own Google provider asks the vault for its
client secret on every code exchange. The Apple extension pinned in the image (1.17.0) does not: it
reads the `.p8` from the raw client-secret field and base64-decodes the literal
`${vault.apple_sign_in_key}`, so it signs nothing and sends Apple no client secret, even with a real
key in the vault file. Until that changes, Sign in with Apple cannot finish a code exchange with real
values (#71).

## Proving it

```bash
infra/secrets/decrypt-smoke.sh
```

It generates a throwaway age key, encrypts `example.yaml` with it, runs the same `decrypt.sh` the
deploy runs, and asserts one file per key with mode 0400 and the value intact. It then runs
`infra/compose/machine-env.sh` over those files, as the deploy does next, and asserts the machine's
`.env` carries the stack values intact, `ssl_required` `external`, no test Member password and no
deploy key, and that a value holding a quote stops it rather than reaching Compose. Nothing secret is
committed and no GitHub secret is needed. CI runs it in the `backend` job, which installs `sops`
and `age` pinned to a checksum in `.github/workflows/ci.yml`.

The ciphertext is safe to commit: the pinned gitleaks and the `detect-private-key` hook were both run
against a real SOPS-and-age file and neither flags it, so no allowlist exists. If Semgrep or Trivy's
secret scanner ever flags one, the fix is a path allowlist carrying the reason, never a scanner
turned off.
