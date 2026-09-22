# infra/secrets

Every secret, encrypted with SOPS and an age key (`docs/ARCHITECTURE.md` A13, §7.5). One age private
key is the only secret a human ever holds; everything else is ciphertext in git. The deploy decrypts
it on the runner and sends the machine a `.env` of the domain, the eight stack values and the four
identity-provider ids (`../compose/machine-env.sh`), and Keycloak's two identity-provider secrets as
vault files. Compose writes the two connection strings from that `.env` into files `members` and
`notify` read, and hands the other values to their containers as environment variables
(`../README.md`, "Deploying").

## What is here, and what is not

| Path | Committed | What it is |
| --- | --- | --- |
| `example.yaml` | yes | The key set with placeholder values, the convention `.env.example` already follows. It is what the wizard checks presence against, so the two files cannot drift. |
| `decrypt.sh` | yes | The deploy's decrypt step: one file per value, mode 0400. |
| `decrypt-smoke.sh` | yes | Its proof, on a throwaway age key it generates itself. |
| `keyset.sh` | yes | The one reader of `example.yaml`'s shape: `keys` lists the key set, `value` prints the placeholder one key carries. The smoke and the wizard both go through it, so a change to the file's shape cannot leave the two disagreeing about the key set. |
| `prod.yaml` | **no** | The real values. Git-ignored until the nine human steps have been done; committing it is the first item of #87. |
| `.sops.yaml`, at the **repository root** | **no** | The creation rule naming the age recipient. It sits at the root because sops resolves the rule by walking up from the working directory, not from the file it is given, so a copy in this directory is invisible to `sops edit infra/secrets/prod.yaml` run from the root. Git-ignored for the same reason as `prod.yaml`: a rule with an invalid recipient makes `sops edit` fail outright rather than degrade, so it cannot be committed before a key exists. |

No age key has been generated yet, so the repository holds the shape and nothing else (#66, PR
#76). The wizard's
first step is `age-keygen`, which writes the new public key into `.sops.yaml` and creates `prod.yaml`
from `example.yaml`.

## The key set

Seven of the nine human steps (`docs/research/2026-09-14-setup-automation.md` §4) put a value here;
Apple Developer Program enrolment and `age-keygen` produce none. The last eight keys are what the
stack itself needs, and the wizard's check mode weighs them alongside the steps.

| Key | Comes from | Consumed by |
| --- | --- | --- |
| `app_store_connect_api_key` | App Store Connect, Users and Access > Integrations | The TestFlight half of the tag deploy (A15), which #87 still owes |
| `apple_services_id` | Apple Developer, Identifiers | The Keycloak Apple identity provider's client id, an OpenTofu variable (#71) |
| `apple_team_id` | Apple Developer, Membership details | The same provider's `teamId`, the issuer of the client-secret JWT it signs |
| `apple_sign_in_key_id` | Apple Developer, Keys — the Key ID of the Sign in with Apple `.p8` | The same provider's `keyId`, the JWT's `kid` |
| `apple_sign_in_key` | Apple Developer, Keys — the Sign in with Apple `.p8` | The same provider, which signs the client secret on every token exchange; referenced as `${vault.apple_sign_in_key}`, which the pinned extension does not resolve (see "Keycloak's file vault") |
| `apple_push_key` | Apple Developer, Keys — the APNs `.p8` | `notify`, slice 7 |
| `google_client_id` | Google Auth Platform > Clients | The Keycloak Google identity provider's client id, an OpenTofu variable (#71) |
| `google_client_secret` | the same client | The same provider, through Keycloak's file vault as `${vault.google_client_secret}` |
| `open_exchange_rates_app_id` | Open Exchange Rates | `currency`, slice 7 |
| `smtp_token` | the SMTP provider's first account-level token | `notify`'s SMTP login on the machine, owed by #87 (`notify` sends no AUTH yet); Mailpit needs none locally |
| `domain` | the domain bought in the same step | The deploy workflow's SSH host, and the machine's `MMR_HOST` |
| `registrar_token` | the registrar, scoped to one zone | OpenTofu's DNS records |
| `cloud_api_token` | the cloud account | OpenTofu's VM, firewall and cloud-init |
| `deploy_ssh_private_key` | `ssh-keygen`, inside the cloud-account step | The deploy workflow's transport to the machine; OpenTofu hands the public half to cloud-init |
| `members_db_password` | chosen once | The Postgres init script, which needs the password by itself to create the role |
| `members_database_url` | assembled from it | `members` and `members migrate`, as `MEMBERS_DATABASE_URL_FILE` |
| `notify_db_password` | chosen once | The same init script, which creates the `notify` role the same way |
| `notify_database_url` | assembled from it | `notify` and `notify migrate`, as `NOTIFY_DATABASE_URL_FILE` |
| `postgres_password` | chosen once | The Postgres superuser, `POSTGRES_PASSWORD` |
| `keycloak_db_password` | chosen once | The init script's `keycloak` role and Keycloak itself, `KEYCLOAK_DB_PASSWORD` |
| `kc_bootstrap_admin_password` | chosen once | Keycloak's first admin, whose username is the literal `admin`, and `realm-apply`'s login |
| `bff_client_secret` | chosen once | The `bff` client in the realm and the BFF, `BFF_CLIENT_SECRET` |

Each stack value becomes the Compose variable of the same name in capitals on the machine
(`infra/compose/machine-env.sh`). There is no `test_user_password`: the realm creates its test Member
only when given a password, and the machine is never given one.

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
itself stays on the runner. What reaches the machine is its `.env`, written from the stack values and
the identity providers' ids, and the two identity-provider secrets as files for Keycloak's file vault
(below, and `../README.md`, "Deploying"). The machine never holds the age key.

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
3.13.3 and `age` 1.3.2 pinned to a checksum.

The ciphertext is safe to commit: the pinned gitleaks and the `detect-private-key` hook were both run
against a real SOPS-and-age file and neither flags it, so no allowlist exists. If Semgrep or Trivy's
secret scanner ever flags one, the fix is a path allowlist carrying the reason, never a scanner
turned off.
