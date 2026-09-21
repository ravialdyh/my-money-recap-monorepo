# infra/wizard

The nine setup steps only a human can do (`docs/research/2026-09-14-setup-automation.md` §4), in
order, in one script. Everything downstream of them is `tofu apply` and `docker compose up`, so this
is where the setup stops living in the developer's head and starts living in the repository.

```bash
infra/wizard/setup.sh            # walk the nine steps
infra/wizard/setup.sh --check    # report which are done; non-zero while any checkable one is not
```

Check mode is non-interactive and touches nothing, so the developer can stop halfway and come back.

## The nine steps

| # | Step | Produces | Checked by |
| --- | --- | --- | --- |
| 1 | One age key pair | `.sops.yaml` and `infra/secrets/prod.yaml` | attempting a decrypt |
| 2 | Apple Developer Program enrolment | nothing | nothing: reported by hand |
| 3 | App Store Connect API key | `app_store_connect_api_key` | key presence |
| 4 | Apple Services ID and the two `.p8` keys | `apple_services_id`, `apple_team_id`, `apple_sign_in_key_id`, `apple_sign_in_key`, `apple_push_key` | key presence |
| 5 | Google OAuth client | `google_client_id`, `google_client_secret` | key presence |
| 6 | Open Exchange Rates App ID | `open_exchange_rates_app_id` | key presence |
| 7 | SMTP provider account and its first token | `smtp_token` | key presence |
| 8 | Domain, and a registrar token scoped to its zone | `domain`, `registrar_token` | key presence |
| 9 | Cloud account, its API token and a deploy key | `cloud_api_token`, `deploy_ssh_private_key` | key presence |

Step 1 is first because every other step pastes into the file it creates, and it is `age-keygen`:
free, local, and not a purchase. Step 9 also runs `ssh-keygen`, so the deploy workflow's transport
credential comes out of a step that already exists and the nine stay nine. No step blocks on a
purchase: each one can be declined and picked up later.

**Step 1 works with a key that is already there.** `$SOPS_AGE_KEY_FILE`, or
`~/.config/sops/age/keys.txt`, is the path sops itself uses, so a developer may arrive holding a key
from another project. The wizard never overwrites that key, and never overwrites an existing
`prod.yaml` either: it derives the recipient with `age-keygen -y` and writes whichever of the two
files step 1 owes — the creation rule, the secrets file — is still missing. Step 9 runs `ssh-keygen`
only once `prod.yaml` exists, so no private key is ever left on disk with nowhere to paste it.

## How a step is judged done

**Key presence is judged against `infra/secrets/example.yaml`**, which is the committed key set, so
the two files cannot drift. A key it declares and `prod.yaml` lacks is named and fails the run; so is
a key it declares that no step and no stack value produces, and a key a step produces that it does
not declare. A key still holding the placeholder value example.yaml carries for it has not been
filled in, because `prod.yaml` is born as an encryption of example.yaml and every key is present from
that moment.

**The age step attempts a decrypt.** Reading the recipient out of the SOPS metadata would prove
nothing: the metadata is there by construction the instant the file is encrypted, whoever holds the
private half. Only a decrypt distinguishes the machine that can read the file from one that cannot,
and the same attempt is what makes the seven value steps readable, so an absent or undecryptable
`prod.yaml` reports all eight checkable steps missing.

**Apple Developer Program enrolment produces no value**, so there is nothing on this machine to look
at. It is reported by hand and never counts toward the exit status (slice 3 spec #61, testing
decision 9).

**The values the stack assembles for itself are checked too.** `members_db_password`,
`members_database_url`, `notify_db_password`, `notify_database_url`, `postgres_password`,
`keycloak_db_password`, `kc_bootstrap_admin_password` and `bff_client_secret` come from no step — a
password chosen once, and for the two databases a connection string built from it — but the deploy
hands whatever they hold to the machine, so a placeholder left in one reaches it. They are
`STACK_KEYS` in `setup.sh`. They get their own
row, below the nine, and they count toward the exit status.

## Where the creation rule goes

`.sops.yaml` is written at the repository root, not beside `prod.yaml`, because `sops` resolves
creation rules by walking up from the working directory rather than from the file named on the
command line. `sops edit infra/secrets/prod.yaml`, the command `docs/ARCHITECTURE.md` §10 and
`infra/secrets/README.md` both quote, is run from the root, so that is where the rule has to be. It
stays git-ignored this slice along with `prod.yaml` (A13, `infra/secrets/README.md`).

## Proving it

```bash
infra/wizard/check-smoke.sh
```

It generates a throwaway age key in a temporary directory, points the wizard at it with
`WIZARD_SECRETS_DIR` and `WIZARD_RULES_FILE`, and asserts the nine cases that matter: an absent
secrets file reports all eight checkable steps missing and Apple by hand; the decrypt attempt fails
without the private half; it passes with it, and every value filled exits zero; a stack value left at
its placeholder is named and fails; a key `example.yaml` declares and `prod.yaml` lacks is caught; so
is a key no step produces; a walk that finds an age key already on disk still writes the creation
rule and the secrets file, and check mode then reads them; a walk with no secrets file leaves no SSH
private key behind; and neither script parses `example.yaml` anywhere but `infra/secrets/keyset.sh`.

Nothing secret is committed, no GitHub secret is needed, and nothing is written inside the
repository: `WIZARD_RULES_FILE` is what keeps a walk's `.sops.yaml` in the temporary directory
instead of the repository root, and the smoke asserts the tree stayed clean. CI runs it in the
`backend` job, beside `infra/secrets/decrypt-smoke.sh` and on the same pinned `sops` and `age`.

`shellcheck` is the other gate, already a pre-commit hook.
