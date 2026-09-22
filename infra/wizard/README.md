# infra/wizard

The nine setup steps only a human can do (`docs/research/2026-09-14-setup-automation.md` §4), in
order, in one script. Everything downstream of them is the planned machine root's `tofu apply` (#87)
and the deploy's `docker compose up`, so this is where the setup stops living in the developer's head
and starts living in the repository.

```bash
infra/wizard/setup.sh            # walk the nine steps
infra/wizard/setup.sh --check    # report which are done; non-zero while any checkable one is not
```

Check mode is non-interactive and touches nothing, so the developer can stop halfway and come back.

## The nine steps

1. One age key pair
2. Apple Developer Program enrolment
3. App Store Connect API key
4. Apple Services ID and the two `.p8` keys
5. Google OAuth client
6. Open Exchange Rates App ID
7. SMTP provider account and its first token
8. Domain, and a registrar token scoped to its zone
9. Cloud account, its API token and a deploy key

The keys each step fills in `infra/secrets/prod.yaml`, and where its value comes from, are `STEPS` in
`setup.sh`; step 1 fills no key but creates the file itself and `.sops.yaml`, and step 2 produces
nothing at all.

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
at. It is reported by hand and never counts toward the exit status (spec #61, testing decision 9).

**The values the stack assembles for itself are checked too.** The twelve `keyset.sh stack` names
(`infra/secrets/README.md`, "The key set") come from no step — a password chosen once, and for the
three service databases and the broker a connection string built from it — but the deploy hands whatever they hold to the
machine, so a placeholder left in one reaches it. They get their own row, below the nine, and they
count toward the exit status.

## Where the creation rule goes

Step 1 writes `.sops.yaml` at the repository root, not beside `prod.yaml`. Why, and when the two are
committed, is in [`../secrets/README.md`](../secrets/README.md), "What is here, and what is not".

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
