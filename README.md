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

Coverage does not gate at a fixed number yet. `tools/coverage/` holds a floor per language that can
only go up — `tools/coverage/README.md` has today's floors and why. The `Security` workflow runs
govulncheck, `npm audit`, OSV-Scanner (Go, npm and Swift together), Semgrep and Trivy on every pull
request and every Monday.

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
