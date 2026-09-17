# My Money Recap

A personal money tracker for adults, styled as a playful kids' app, that turns one month of
income, spending and saving into a Wrapped-style monthly recap.

This repository is the monorepo for it: product docs and design record, illustration source, design system, apps and infrastructure.

## Layout

| Path | What lives here |
| --- | --- |
| `docs/` | Product-level docs: the PRD, the screen-state contract, the domain glossary, the backend architecture, the iOS design and the design handoff (all 111 artboards) |
| `my-money-recap/` | Illustration source: the WebP artwork, the app icon, and the scripts and prompts that produce them |
| `my-money-recap-ds/` | `@my-money-recap/ui` — the React design system: tokens, components, the four Recap screens |
| `apps/` | Runnable products: `apps/backend` (one Go module: the BFF and the `members` service) and `apps/ios` (the SwiftUI app). See `apps/README.md` |
| `packages/` | `packages/contract`, the Buf module every client is generated from |
| `infra/` | `infra/compose`, the phase-1 Compose stack (Keycloak, Postgres, RabbitMQ, Redis, `members`, the BFF) |

It is an npm workspace with one member today, `my-money-recap-ds` (`apps/*` and `packages/*` are
declared for later JavaScript packages; nothing there has a `package.json`), so one `npm install` at
the root wires it up and there is a single `package-lock.json`. Go, Buf and Xcode are separate
toolchains; each area's README says what it needs.

## How the pieces connect

```
docs/PRD.md + docs/SCREEN-STATES.md ──rules──►  docs/design/design_handoff_my_money_recap/  (the drawings)
my-money-recap/assets/illustrations/ ──build:assets──►  my-money-recap-ds/src/assets/illustrations.ts
my-money-recap-ds/src/tokens/tokens.json ──build:tokens──►  tokens.css, styles/css.generated.ts, dist/mmr.css
my-money-recap-ds/src/tokens/tokens.json ──build:swift-tokens──►  apps/ios/DesignSystem/Tokens.swift
```

The design system reads the WebP artwork straight out of `my-money-recap/assets/illustrations/`
via a relative path, so the two folders must stay siblings inside this repo.

## Common commands

From the repo root:

```bash
npm install
```

```bash
npm run build:assets
```

```bash
npm run build:tokens && npm run typecheck
```

```bash
npm run build:swift-tokens
```

`build:assets` re-inlines `my-money-recap/assets/illustrations/*.webp` into the design system;
`build:tokens` regenerates `tokens.css`, `css.generated.ts` and `dist/mmr.css` from `tokens.json`;
`build:swift-tokens` regenerates `apps/ios/DesignSystem/Tokens.swift` from the same `tokens.json` (`docs/IOS.md` I5).
The `design-system` CI job runs all three and fails if any output is stale, so generated files can never
drift from source. The other CI jobs (`contract`, `backend`, `ios`) are listed under Delivery in `docs/ARCHITECTURE.md` §10.

## Checks

Every check runs on every commit through `pre-commit`, and again in CI. Install it once:

```bash
pipx install pre-commit   # or: pip install pre-commit
pre-commit install --install-hooks
```

`pre-commit` is Python; nothing else in this repo is. It is here because it pins every hook to a
version, keeps them in their own environments, and Dependabot raises those versions.

```bash
pre-commit run --all-files              # what CI's pre-commit job runs
pre-commit run --hook-stage pre-push    # also regenerates and diffs the generated files
```

| Check | Tool | Config |
| --- | --- | --- |
| Whitespace, JSON and YAML, large files, private keys, case conflicts, commits to `main` | `pre-commit-hooks` | `.pre-commit-config.yaml` |
| Secrets | gitleaks | `.pre-commit-config.yaml` |
| File size and comment policy | `tools/policy/check_policy.py` | `tools/policy/README.md` |
| TypeScript lint and format | Biome | `biome.json` |
| Go lint and format | golangci-lint, gofmt | `.golangci.yml` |
| Swift lint and format | SwiftLint, SwiftFormat | `.swiftlint.yml`, `.swiftformat` |
| Shell, Dockerfile, workflows | shellcheck, hadolint, actionlint | `.pre-commit-config.yaml` |
| Protobuf | buf lint | `packages/contract/buf.yaml` |

golangci-lint, SwiftLint, SwiftFormat, hadolint and buf are per-language toolchains. The hooks skip
with a note where the binary is missing, so a contributor touching only the design system does not
have to install all of them; CI installs every one and runs the same commands as hard gates.

SwiftFormat and SwiftLint only exist on macOS, so run them on a Mac before pushing Swift:

```bash
swiftformat . && swiftlint --fix && swiftlint lint --strict
```

CI installs SwiftLint with `brew` before running it. That step looks redundant against
`actions/runner-images`, because `images/macos/macos-15-Readme.md` lists SwiftLint 0.65.1 — but that
is the **Intel** image. `macos-latest` resolves to arm64 here (the job log shows Homebrew at
`/opt/homebrew`, not `/usr/local`), and `macos-15-arm64-Readme.md` lists SwiftFormat and no
SwiftLint. Without the install the job dies with exit 127, which is exactly how it failed earlier in
this repo's history. SwiftFormat needs no install on either image.

Coverage does not gate at a fixed number yet. `tools/coverage/` holds a floor per language that can
only go up — `tools/coverage/README.md` has today's floors and why. The `Security` workflow runs
govulncheck, OSV-Scanner (Go, npm and Swift together), Semgrep, Trivy and a full-tree gitleaks scan
on every pull request and every Monday. OSV reads the same GHSA advisories `npm audit` would, so the
npm tree is covered once, without a second `npm ci`.

### Why these settings are what they are

The configs carry no comments, so the reasoning lives here.

- **golangci-lint.** `staticcheck.checks: all` turns on the `ST` group, which already covers package
  comments, naming, error strings and receiver names; `revive` would repeat all of it, so it is off.
  The `comments` exclusion preset drops every "should have a comment" finding, because comment style
  is `tools/policy`'s job and two tools with different opinions is how the old `revive` block ended
  up enabled but silently suppressed. `gosec` G115 is excluded: it fires on every `int` → `int32` at
  the generated protobuf boundary.
- **SwiftLint.** It owns Swift correctness only; SwiftFormat owns every formatting decision and
  `tools/policy` owns file length and comments. `trailing_comma` is **off on purpose**: SwiftFormat's
  `--commas always` adds the comma that rule deletes, so with both on, the two tools undo each other
  every run and `swiftformat --lint` fails in CI. `function_body_length`, `file_length` and `todo`
  are off because `tools/policy` already caps length and bans tracker tags. `identifier_name`
  allows one-character names: the design system's size steps are `l`, `m` and `s` in both languages.
- **Dependabot.** Every ecosystem waits seven days before proposing a newly published version — a
  release hours old can be a compromised one. Security updates ignore the cooldown, so a real
  advisory still arrives at once. Each group is also `update-types: ["minor", "patch"]`, so routine
  bumps arrive batched and quiet while every **major** gets a pull request of its own. Without that
  a wildcard group hides majors inside a routine-looking update: the first npm group PR raised here
  was titled "bump the node group with 5 updates" and was in fact React 18 → 19 *and* TypeScript
  5 → 7. `swift` and `docker` are ungrouped, so their majors already arrive one per pull request.
- **Actions are pinned to a commit, never a tag.** A tag is mutable, and whoever owns the action can
  repoint it at code that runs with this job's token. The trailing `# vN` is Dependabot's, and
  `tools/policy` exempts it for that reason.
- **The Compose stack** is documented in `infra/compose/README.md`, including the Keycloak pin, the
  `/dev/tcp` health check and the parallel-worktree ports.

### What CI costs, and why jobs are skipped

This repository is private, so Actions minutes are billed, and macOS bills at **ten times** the
Linux rate. A run where every job fires costs about 41 billable minutes, of which 30 are the `ios`
job alone — even though the whole run finishes in under four minutes of wall clock.

So each job is gated on whether its inputs changed, using `dorny/paths-filter` at the **job** level
rather than a workflow-level `paths:` filter. That distinction matters: a workflow skipped by
`paths:` never reports a status at all, so a required check would wait for it forever, while a job
skipped by `if:` reports success. The `all green` job gathers the results and is the one worth
requiring. Over the last twenty merges this drops the average run from 41 to about 18 billable
minutes; a docs-only change costs 5.

To move the iOS job onto your own Mac and stop paying for it entirely, register a self-hosted runner
(repository → Settings → Actions → Runners → New self-hosted runner, macOS), then set a repository
**variable** `IOS_RUNNER` to `self-hosted`. The workflow reads
`runs-on: ${{ vars.IOS_RUNNER || 'macos-latest' }}`, so nothing else changes, and unsetting the
variable moves it straight back. Self-hosted runners are only safe on a private repository like this
one — on a public repo a fork's pull request could run its own code on your machine.

## Docs

- [`docs/PRD.md`](docs/PRD.md) — product requirements and the numbered decision log
- [`docs/SCREEN-STATES.md`](docs/SCREEN-STATES.md) — every screen and state, with its copy
- [`docs/CONTEXT.md`](docs/CONTEXT.md) — domain glossary (Member, Salary, Spend, Recap, Archetype…). Read this before naming anything.
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — the backend design: Nginx, the Go BFF, the four services, RabbitMQ, Redis, Postgres, Keycloak, and the numbered `A*` decisions
- [`docs/IOS.md`](docs/IOS.md) — the iOS app's internals: sign-in, session, networking, the SwiftData cache, generated tokens, and the numbered `I*` decisions
- [`docs/design/design_handoff_my_money_recap/README.md`](docs/design/design_handoff_my_money_recap/README.md) — the design handoff: eleven canvases, 111 artboards
- [`my-money-recap-ds/README.md`](my-money-recap-ds/README.md) — the design system: visual rules and component API
- [`my-money-recap/README.md`](my-money-recap/README.md) — how a rendered illustration becomes a shipped asset
- [`apps/README.md`](apps/README.md) — the Go backend and the iOS app: build, test, run
- [`infra/compose/README.md`](infra/compose/README.md) — the local stack: ports, the `mmr` realm, a phone on the LAN
- [`packages/contract/README.md`](packages/contract/README.md) — the contract: what is generated where, and the breaking-change rule

## Single source of truth

| Thing | Lives in | Everything else is generated |
| --- | --- | --- |
| Design tokens | `my-money-recap-ds/src/tokens/tokens.json` | `tokens.css`, `css.generated.ts`, `dist/mmr.css`, `apps/ios/DesignSystem/Tokens.swift` |
| Illustration artwork | `my-money-recap/assets/illustrations/*.webp` | `my-money-recap-ds/src/assets/illustrations.ts` |
| Product and design decisions | `docs/PRD.md` (numbered `D*` decisions) | — |
| Backend architecture decisions | `docs/ARCHITECTURE.md` (numbered `A*` decisions) | — |
| iOS app decisions | `docs/IOS.md` (numbered `I*` decisions) | — |
| Visual rules | `my-money-recap-ds/README.md` § Design rules | — |
| Screens | `docs/design/design_handoff_my_money_recap/` | renders (not in git) |
| Domain vocabulary | `docs/CONTEXT.md` | — |
| Sample ledger | the prop defaults in `my-money-recap-ds/src/screens/` (one file per Recap screen) | stories and design-sync previews |
