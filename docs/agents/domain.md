# Domain Docs

How the engineering skills consume this repo's domain documentation.

## Before exploring, read these

- `docs/CONTEXT.md` — the glossary. Use its terms exactly; never a synonym it lists under "Avoid".
- The decision registers, in place of ADRs:
  - `docs/PRD.md` §4 — product decisions D1–D71, and §8 for the money rules
  - `docs/ARCHITECTURE.md` §11 — backend decisions A1–A15; §9 records the alternatives rejected
  - `docs/IOS.md` §2 — app decisions I1–I9
- `docs/SCREEN-STATES.md` — the per-screen contract and copy, keyed by screen id.

There is no `docs/adr/` folder and none should be created. A new decision is a new numbered row in the matching register, with a Status column entry; a changed decision is the row rewritten in place with the date, as the registers already do.

## Layout

Single-context. One glossary, three registers, all under `docs/`. `packages/contract/` holds the protobuf schema; the ledger fixtures arrive there with slice 5.

## Use the glossary's vocabulary

Issue titles, test names and proposals use the glossary's terms: Member, Salary, Payday, Spend, Category, Saving Target, Saved, Money In, Streak, Splurge, Archetype, Recap. A concept missing from the glossary is either invented language (reconsider) or a real gap (note it for `/domain-modeling`).

## Flag decision conflicts

If your output contradicts a numbered decision, say so explicitly rather than silently overriding:

> _Contradicts A14 (idempotency in the owning service) — but worth reopening because…_
