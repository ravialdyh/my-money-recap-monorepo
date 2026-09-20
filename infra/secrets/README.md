# infra/secrets

Every secret, encrypted with SOPS and an age key (`docs/ARCHITECTURE.md` A13, §7.5). One age private
key is the only secret a human ever holds; everything else is ciphertext in git and arrives inside a
container as a file, so nothing sensitive shows up in `docker inspect` or a process list.

## What is here, and what is not

| Path | Committed | What it is |
| --- | --- | --- |
| `example.yaml` | yes | The key set with placeholder values, the convention `.env.example` already follows. It is what the wizard checks presence against, so the two files cannot drift. |
| `decrypt.sh` | yes | The deploy's decrypt step: one file per value, mode 0400. |
| `decrypt-smoke.sh` | yes | Its proof, on a throwaway age key it generates itself. |
| `prod.yaml` | **no** | The real values. Git-ignored until the nine human steps have been done; committing it is the first item of the slice's follow-up issue. |
| `.sops.yaml` | **no** | The creation rule naming the age recipient. Git-ignored for the same reason: a rule with an invalid recipient makes `sops edit` fail outright rather than degrade, so it cannot be committed before a key exists. |

No age key has been generated yet, so this slice commits the shape and nothing else. The wizard's
first step is `age-keygen`, which writes the new public key into `.sops.yaml` and creates `prod.yaml`
from `example.yaml`.

## The key set

Seven of the nine human steps (`docs/research/2026-09-14-setup-automation.md` §4) put a value here;
Apple Developer Program enrolment and `age-keygen` produce none. The last two keys are what the stack
itself needs.

| Key | Comes from | Consumed by |
| --- | --- | --- |
| `app_store_connect_api_key` | App Store Connect, Users and Access > Integrations | The App Store Connect automation, slice 8 |
| `apple_services_id` | Apple Developer, Identifiers | The Keycloak Apple identity provider |
| `apple_sign_in_key` | Apple Developer, Keys — the Sign in with Apple `.p8` | The same provider, which signs the client secret on every token exchange |
| `apple_push_key` | Apple Developer, Keys — the APNs `.p8` | `notify`, slice 7 |
| `google_client_id` | Google Auth Platform > Clients | The Keycloak Google identity provider |
| `google_client_secret` | the same client | The same provider |
| `open_exchange_rates_app_id` | Open Exchange Rates | `currency`, slice 6 |
| `smtp_token` | the SMTP provider's first account-level token | `notify`, slice 7; Mailpit stands in locally |
| `registrar_token` | the registrar, scoped to one zone | OpenTofu's DNS records |
| `cloud_api_token` | the cloud account | OpenTofu's VM, firewall and cloud-init |
| `deploy_ssh_private_key` | `ssh-keygen`, inside the cloud-account step | The deploy workflow's transport to the machine; OpenTofu hands the public half to cloud-init |
| `members_db_password` | chosen once | The Postgres init script, which needs the password by itself to create the role |
| `members_database_url` | assembled from it | `members` and `members migrate`, as `MEMBERS_DATABASE_URL_FILE` |

The last two hold the same password twice on purpose: `members` reads one whole connection string
rather than a password of its own, and `postgres/10-databases.sh` reads the bare password. Keep them
in step when either changes.

## Decrypting

```bash
SOPS_AGE_KEY_FILE=~/.config/sops/age/keys.txt infra/secrets/decrypt.sh infra/secrets/prod.yaml /run/secrets
```

Each top-level key becomes a file of that name holding the value with no trailing newline, mode 0400
in a directory the umask leaves at 0700. Compose mounts those files with `secrets:`; Keycloak's file
vault reads the same directory. The deploy workflow runs this as its **first** step, before anything
is copied, because the transport credential is inside the file.

`infra/secrets/decrypted/` is the one path inside the repository that is safe to point it at, and
`.gitignore` covers it. The plaintext is what needs guarding and no scanner guards it: these files
are bare values whose names carry no keyword context, so nothing would flag one that got committed.

## Proving it

```bash
infra/secrets/decrypt-smoke.sh
```

It generates a throwaway age key, encrypts `example.yaml` with it, runs the same `decrypt.sh` the
deploy runs, and asserts one file per key with mode 0400 and the value intact. Nothing secret is
committed and no GitHub secret is needed. CI runs it in the `backend` job, which installs `sops`
3.13.3 and `age` 1.3.2 pinned to a checksum.

The ciphertext is safe to commit: the pinned gitleaks and the `detect-private-key` hook were both run
against a real SOPS-and-age file and neither flags it, so no allowlist exists. If Semgrep or Trivy's
secret scanner ever flags one, the fix is a path allowlist carrying the reason, never a scanner
turned off.
