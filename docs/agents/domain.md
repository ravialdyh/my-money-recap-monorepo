# Domain Docs

How the engineering skills use this repo's domain documentation.

## Before exploring, read these

- `docs/CONTEXT.md`: the glossary. Use its terms exactly; never a word it lists under "Avoid".
- The README of the part you are changing. Each one says what that part is and does.
- `apps/backend/cmd/ledger/MONEY-RULES.md`: the money rules, with worked examples from
  `packages/contract/fixtures/ledger/`.

There is no decisions register and no `docs/adr/` folder; do not create one. The code, its tests and the READMEs
are the record of what the app does. Git history and pull requests hold why.

## Layout

Single-context. One glossary under `docs/`. `packages/contract/` holds the protobuf schema and, under
`fixtures/ledger/`, the answer key for the money rules.

## Use the glossary's vocabulary

Issue titles, test names and proposals use the glossary's terms: Member, Salary, Payday, Spend, Category, Saving
Target, Saved, Money In, Streak, Splurge, Archetype, Recap. A concept missing from the glossary is either invented
language (reconsider) or a real gap (note it for `/domain-modeling`).

## Flag conflicts

If your output contradicts a rule a README or `MONEY-RULES.md` states, say so explicitly rather than silently
changing it:

> _Contradicts the one-key-per-write rule in `apps/README.md`, but worth reopening because…_
