# My Money Recap — agent instructions

Read `README.md` for the layout and `docs/CONTEXT.md` before naming anything.

## Branching

Never commit to `main`. Every change is a short branch off `main` and a pull request; CI must be green before merge. Deployment is a version tag, not a merge (`docs/ARCHITECTURE.md` A15). GitHub cannot enforce this on the repository's plan, so this rule is the gate.

## House rules the hooks enforce

`pre-commit run --all-files` is the gate. Two rules are this repo's own, in
`tools/policy/README.md`:

- **300 lines per source file.** Generated and tool-owned files are excluded.
- **A comment either states a contract or is a directive.** Doc comments on a package, type or
  member are how code cites a `D*`, `A*` or `I*` decision, and they are capped at 255 characters.
  Comments inside a function body are rejected: name the code instead. `TODO`, commented-out code
  and anything restating the code are rejected. `//go:embed`, `// MARK:`, `//nolint` and the other
  directives are untouched.

Coverage has a floor per language that only rises (`tools/coverage/README.md`). When a change adds
tests, raise the floor in the same pull request.

## Agent skills

### Issue tracker

Issues and specs are GitHub Issues on `ravialdyh/my-money-recap-monorepo`, via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

See `docs/agents/triage-labels.md`.

### Domain docs

Single-context. Glossary at `docs/CONTEXT.md`; decisions are the numbered tables in `docs/PRD.md` (D*), `docs/ARCHITECTURE.md` (A*) and `docs/IOS.md` (I*), with no `adr/` folder. See `docs/agents/domain.md`.
