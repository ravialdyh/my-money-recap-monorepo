# My Money Recap: agent instructions

Read `README.md` for the layout and `docs/CONTEXT.md` before naming anything.

## Branching

Never commit to `main`. Every change is a short branch off `main` and a pull request; CI must be green before merge. Deployment is a version tag, not a merge (`docs/ARCHITECTURE.md` A15). A ruleset on `main` enforces it: no direct push, no force push, and the `ci` and `security` checks must pass, with no one exempt.

A branch is named `<type>/<kebab-summary>`, where `<type>` is the change's Conventional Commits type, matching the commit subject: `feat/members-service`, `chore/ci-fold-and-tokens`, `build/react-19`. `dependabot/*` is the one exception, because the tool owns those names.

A session that opens on a generated branch, `claude/confident-hypatia-bssrd9` and the like, renames it before the first commit (`git branch -m <type>/<summary>`) and pushes only the renamed name. That name carries no information into the pull request list, and `claude/clever-babbage-vf9f3y` in PR #32 is the one that got away. This paragraph is the standing permission to rename; do not stop to ask for it.

## One Compose stack per laptop

The laptop runs one Compose stack, the stock `mmr` project on the default ports
(`infra/compose/README.md`, "One stack per laptop"). An agent, a subagent or a worktree never starts a
second one: no other `COMPOSE_PROJECT_NAME`, no other host ports. Before `up`, run
`docker compose -f infra/compose/compose.yaml --profile mail ps` and reuse what is running. Only one
session at a time runs `up`, `--build`, `down` or the Go tests against it; parallel agents take work that
needs no stack. Boot at most one iOS simulator. When the work is done, shut it down
(`xcrun simctl shutdown all`) and stop the stack (`--profile mail down`).

## House rules the hooks enforce

`pre-commit run --all-files` is the gate. Three rules are this repo's own, in
`tools/policy/README.md`:

- **A line cap on every text file** outside the exclusions the policy README lists.
- **No em dash anywhere**: not in code, comments, documents or canvases, excluded paths included; use a full
  stop, a comma or a colon.
- **A comment either states a contract or is a directive.** Doc comments on a package, type or
  member are how code cites a `D*`, `A*` or `I*` decision, and their length is capped. Comments
  inside a function body are rejected: name the code instead. `TODO` and commented-out code are
  rejected too. `//go:embed`, `//go:build`, `//nolint` and the other directives are untouched.

A comment that restates the code is out as well, but no hook can see one: that rule is review's.

Coverage has a floor per language that only rises, up to the target (`tools/coverage/README.md`).
When coverage rises while a floor is still below the target, raise it in the same pull request.

## Agent skills

### Issue tracker

Issues and specs are GitHub Issues on `ravialdyh/my-money-recap-monorepo`, via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

See `docs/agents/triage-labels.md`.

### Domain docs

Single-context. Glossary at `docs/CONTEXT.md`; decisions are the numbered tables in `docs/PRD.md` (D*), `docs/ARCHITECTURE.md` (A*) and `docs/IOS.md` (I*), with no `adr/` folder. See `docs/agents/domain.md`.
