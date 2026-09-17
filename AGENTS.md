# My Money Recap — agent instructions

Read `README.md` for the layout and `docs/CONTEXT.md` before naming anything.

## Branching

Never commit to `main`. Every change is a short branch off `main` and a pull request; CI must be green before merge. Deployment is a version tag, not a merge (`docs/ARCHITECTURE.md` A15). GitHub cannot enforce this on the repository's plan, so this rule is the gate.

One exception, decided 2026-09-18: when GitHub refuses to start the jobs at all (the check-run annotation reads "The job was not started because recent account payments have failed or your spending limit needs to be increased"), CI has not judged the change, so it does not block the merge. The gate is then the same commands run locally on the rebased branch: `pre-commit run --all-files`, and for the contract `buf lint`, `buf breaking` against `origin/main`, `buf generate` with a clean tree, `go build ./...` and `go vet ./...`. The pull request records which jobs were refused and what ran locally. A job that started and failed is never covered by this.

A branch is named `<type>/<kebab-summary>`, where `<type>` is the change's Conventional Commits type, matching the commit subject: `feat/members-service`, `chore/ci-fold-and-tokens`, `build/react-19`. `dependabot/*` is the one exception, because the tool owns those names.

A session that opens on a generated branch, `claude/confident-hypatia-bssrd9` and the like, renames it before the first commit (`git branch -m <type>/<summary>`) and pushes only the renamed name. That name carries no information into the pull request list, and `claude/clever-babbage-vf9f3y` in PR #32 is the one that got away. This paragraph is the standing permission to rename; do not stop to ask for it.

## House rules the hooks enforce

`pre-commit run --all-files` is the gate. Two rules are this repo's own, in
`tools/policy/README.md`:

- **300 lines per source file.** Generated and tool-owned files are excluded.
- **A comment either states a contract or is a directive.** Doc comments on a package, type or
  member are how code cites a `D*`, `A*` or `I*` decision, and they are capped at 255 characters.
  Comments inside a function body are rejected: name the code instead. `TODO`, commented-out code
  and anything restating the code are rejected. `//go:embed`, `//go:build`, `//nolint` and the other
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
