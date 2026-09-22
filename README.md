# My Money Recap

A personal money tracker for adults, styled as a playful kids' app, that turns one month of
income, spending and saving into a Wrapped-style monthly recap.

This repository is the monorepo for it: product docs and design record, illustration source, design system, apps and infrastructure.

## Layout

| Path | What lives here |
| --- | --- |
| `docs/` | Product-level docs: the PRD, the screen-state contract, the domain glossary, the backend architecture, the iOS design and the design handoff |
| `my-money-recap/` | Illustration source: the WebP artwork, the app icon, and the scripts and prompts that produce them |
| `my-money-recap-ds/` | `@my-money-recap/ui` — the React design system: tokens, components, the four Recap screens |
| `apps/` | Runnable products: `apps/backend` (one Go module: the BFF and the `members` and `notify` services) and `apps/ios` (the SwiftUI app). See `apps/README.md` |
| `packages/` | `packages/contract`, the Buf module every client is generated from |
| `infra/` | `infra/compose`, the phase-1 Compose stack; `infra/tofu/realm`, the Keycloak realm in OpenTofu; `infra/nginx`, the edge's configuration and static files; `infra/secrets`, the SOPS-encrypted secret set, committed as its shape; `infra/wizard`, the nine setup steps only a human can do. See `infra/README.md` |

It is an npm workspace with one member today, `my-money-recap-ds` (`apps/*` and `packages/*` are
declared for later JavaScript packages; nothing there has a `package.json`), so one `npm install` at
the root wires it up and there is a single `package-lock.json`. Go, Buf, Xcode, OpenTofu, SOPS and
age are separate toolchains; each area's README says what it needs.

## How the pieces connect

`docs/PRD.md` and `docs/SCREEN-STATES.md` hold the rules the design handoff
(`docs/design/design_handoff_my_money_recap/`) draws. Every other link is a generator; "Single source
of truth" below lists each source, the command and what it writes. The design system reads the WebP
artwork straight out of `my-money-recap/assets/illustrations/` via a relative path, so the two folders
must stay siblings inside this repo.

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

The three `build:` commands write what "Single source of truth" lists. The `design-system` CI job runs
all three and fails if any output is stale, so generated files can never drift from source. The other
CI jobs (`contract`, `backend`, `ios`) are listed under Delivery in `docs/ARCHITECTURE.md` §10.

## Checks

Every check runs on every commit through `pre-commit`, and again in CI. Install it once:

```bash
pipx install pre-commit   # or: pip install pre-commit
pre-commit install --install-hooks
```

`pre-commit` is Python; nothing else in this repo is. It is here because it pins every hook to a
version, keeps them in their own environments, and Dependabot raises those versions.

```bash
pre-commit run --all-files              # CI's pre-commit job runs it, skipping the hooks other jobs run
pre-commit run --hook-stage pre-push    # also regenerates and diffs the generated files
```

| Check | Tool | Config |
| --- | --- | --- |
| Whitespace, JSON and YAML, large files, private keys, case conflicts, commits to `main` | `pre-commit-hooks` | `.pre-commit-config.yaml` |
| Secrets | gitleaks | `.pre-commit-config.yaml` |
| gitleaks, `sops` and OpenTofu each pinned to one version everywhere they are fetched | `tools/hooks/check-pins.sh` | `.pre-commit-config.yaml` |
| File size and comment policy | `tools/policy/check_policy.py` | `tools/policy/README.md` |
| TypeScript lint and format | Biome | `biome.json` |
| Go lint and format | golangci-lint, gofmt | `.golangci.yml` |
| Swift lint and format | SwiftLint, SwiftFormat | `.swiftlint.yml`, `.swiftformat` |
| Shell, Dockerfile, workflows | shellcheck, hadolint, actionlint | `.pre-commit-config.yaml` |
| Protobuf | buf lint | `packages/contract/buf.yaml` |
| OpenTofu format and validity | tofu fmt, tofu validate | `tools/hooks/tofu.sh` |

golangci-lint, SwiftLint, SwiftFormat, hadolint, buf and OpenTofu are per-language toolchains. The hooks skip
with a note where the binary is missing, so a contributor touching only the design system does not
have to install all of them; CI installs every one and runs the same commands as hard gates.
SOPS and age have no hook. The wizard needs them, and so do `infra/secrets/decrypt-smoke.sh` and
`infra/wizard/check-smoke.sh`, which CI's `backend` job runs after installing both, pinned to a
checksum.

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

Coverage gates on a floor per language that only ever rises; `tools/coverage/README.md` has today's
floors and where each came from. The `Security` workflow runs govulncheck, OSV-Scanner (Go, npm and
Swift together), Semgrep, Trivy and a full-tree gitleaks scan on every pull request, every push to
`main` and every Monday.
OSV reads the same GHSA advisories `npm audit` would, so the npm tree is covered once, without a
second `npm ci`.

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
- **Design-system tests.** `vitest.config.ts` runs them in `jsdom` so components can actually be
  rendered and clicked, which a Node environment cannot do. `restoreMocks` and `unstubGlobals` reset
  every spy and global between tests, because the suite replaces `ResizeObserver`,
  `requestAnimationFrame`, `matchMedia` and `document.fonts` — none of which jsdom provides — and a
  leak between files would make a pass depend on test order. React Testing Library drives the
  components through the DOM rather than their internals, so a refactor that keeps the markup keeps
  the tests.
- **Dependabot.** Every ecosystem waits seven days before proposing a newly published version — a
  release hours old can be a compromised one. Security updates ignore the cooldown, so a real
  advisory still arrives at once. Each group is also `update-types: ["minor", "patch"]`, so routine
  bumps arrive batched and quiet while every **major** gets a pull request of its own. Without that
  a wildcard group hides majors inside a routine-looking update: the first npm group PR raised here
  was titled "bump the node group with 5 updates" and was in fact React 18 → 19 *and* TypeScript
  5 → 7. `swift` and `docker` are ungrouped, so their majors already arrive one per pull request.
  The `pre-commit` ecosystem bumps the gitleaks hook's `rev`, but no ecosystem sees the gitleaks
  binary `security.yml` downloads by URL and sha256: a hooks pull request that bumps gitleaks must
  also update that URL and sha256. `tools/hooks/check-pins.sh` fails until the URL carries the
  hook's version; it does not read the sha256, which only the `secrets` job's `sha256sum -c` checks.
  No ecosystem sees the `sops`, `age`, hadolint or OpenTofu downloads either; the same script holds
  the two copies of `sops` and of OpenTofu to one version.
- **Actions are pinned to a commit, never a tag.** A tag is mutable, and whoever owns the action can
  repoint it at code that runs with this job's token. The trailing `# vN` is Dependabot's, and
  `tools/policy` exempts it for that reason.
- **The Compose stack** is documented in `infra/compose/README.md`, including the Keycloak pin, the
  `/dev/tcp` health check and the parallel-worktree ports.

### What CI costs, and why jobs are skipped

This repository is private, so Actions minutes are billed: macOS at **ten times** the Linux rate, and
every job rounded up to the whole minute. Each job of the `CI` workflow therefore runs only when the
files it builds or checks changed, so a change that touches no Swift never starts the macOS job.
`docs/ARCHITECTURE.md` A15 records that design, and why `all green` is the job to require rather than
any single one. The `Security` workflow's jobs are not gated: they all run on Linux, so gating saves
little, and govulncheck reports only the advisories the Go code reaches, so a change to Go code alone
can make it fail while `go.mod` stays the same.

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
- [`docs/design/design_handoff_my_money_recap/README.md`](docs/design/design_handoff_my_money_recap/README.md) — the design handoff: every canvas and artboard
- [`my-money-recap-ds/README.md`](my-money-recap-ds/README.md) — the design system: visual rules and component API
- [`my-money-recap/README.md`](my-money-recap/README.md) — how a rendered illustration becomes a shipped asset
- [`apps/README.md`](apps/README.md) — the Go backend and the iOS app: build, test, run
- [`infra/compose/README.md`](infra/compose/README.md) — the Compose stack: its containers and ports, the `mmr` realm, a phone on the LAN
- [`packages/contract/README.md`](packages/contract/README.md) — the contract: what is generated where, and the breaking-change rule

## Single source of truth

| Thing | Lives in | Everything else is generated |
| --- | --- | --- |
| Design tokens | `my-money-recap-ds/src/tokens/tokens.json` | `build:tokens`: `tokens.css`, `styles/css.generated.ts`, `dist/mmr.css`; `build:swift-tokens`: `apps/ios/DesignSystem/Tokens.swift` (`docs/IOS.md` I5) |
| Illustration artwork | `my-money-recap/assets/illustrations/*.webp` | `build:assets`: `my-money-recap-ds/src/assets/illustrations.ts`. The iOS asset catalog's PNGs are copied by hand (`docs/IOS.md` I5) |
| The API contract | `packages/contract/money_recap/v1/*.proto` | `buf generate`: `apps/backend/gen/`, `packages/contract/swift/Sources/` (`packages/contract/README.md`) |
| Product and design decisions | `docs/PRD.md` (numbered `D*` decisions) | — |
| Backend architecture decisions | `docs/ARCHITECTURE.md` (numbered `A*` decisions) | — |
| iOS app decisions | `docs/IOS.md` (numbered `I*` decisions) | — |
| Visual rules | `my-money-recap-ds/README.md` § Design rules | — |
| Screens | `docs/design/design_handoff_my_money_recap/` | renders (not in git) |
| Domain vocabulary | `docs/CONTEXT.md` | — |
| Sample ledger | `docs/PRD.md` §8, and its answer key in `packages/contract/fixtures/ledger/` | not generated: the prop defaults in `my-money-recap-ds/src/screens/` copy Maya's September by hand for the stories and design-sync previews, and PRD §8 wins where they differ |
