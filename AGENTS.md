# My Money Recap — agent instructions

Read `README.md` for the layout and `docs/CONTEXT.md` before naming anything.

## Branching

Never commit to `main`. Every change is a short branch off `main` and a pull request; CI must be green before merge. Deployment is a version tag, not a merge (`docs/ARCHITECTURE.md` A15). GitHub cannot enforce this on the repository's plan, so this rule is the gate.

## Agent skills

### Issue tracker

Issues and specs are GitHub Issues on `ravialdyh/my-money-recap-monorepo`, via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

The five default labels, each string equal to its role name. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context. Glossary at `docs/CONTEXT.md`; decisions are the numbered tables in `docs/PRD.md` (D*), `docs/ARCHITECTURE.md` (A*) and `docs/IOS.md` (I*), with no `adr/` folder. See `docs/agents/domain.md`.
