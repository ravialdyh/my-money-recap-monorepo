# My Money Recap

A personal money tracker for adults, styled as a playful kids' app, that turns one month of
income, spending and saving into a Wrapped-style monthly recap.

This repository is the monorepo for it: product docs and design record, illustration source, design system, apps and infrastructure.

## Layout

| Path | What lives here |
| --- | --- |
| `docs/` | Product-level docs: the PRD, the screen-state contract, the domain glossary, the backend architecture and the design handoff (all 118 screens) |
| `my-money-recap/` | Illustration source: the WebP artwork, the app icon, and the scripts and prompts that produce them |
| `my-money-recap-ds/` | `@my-money-recap/ui` — the React design system: tokens, components, the four Recap screens |
| `apps/` | Runnable products (mobile, web, backend). Empty for now |
| `packages/` | Shared libraries beyond the design system. Empty for now |
| `infra/` | Deployment and infrastructure. Empty for now |

It is an npm workspace: `my-money-recap-ds`, `apps/*` and `packages/*` are members, so one
`npm install` at the root wires everything up and there is a single `package-lock.json`.

## How the pieces connect

```
docs/PRD.md + docs/SCREEN-STATES.md ──rules──►  docs/design/design_handoff_my_money_recap/  (the drawings)
                                                          │ built from
my-money-recap/assets/illustrations/ ──build:assets──►  my-money-recap-ds/src/assets/illustrations.ts
my-money-recap-ds/src/tokens/tokens.json ──build:tokens──►  tokens.css, styles/css.generated.ts, dist/mmr.css
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

`build:assets` re-inlines `my-money-recap/assets/illustrations/*.webp` into the design system;
`build:tokens` regenerates `tokens.css`, `css.generated.ts` and `dist/mmr.css` from `tokens.json`.
CI runs both and fails if either output is stale, so generated files can never drift from source.

## Docs

- [`docs/PRD.md`](docs/PRD.md) — product requirements and the numbered decision log
- [`docs/SCREEN-STATES.md`](docs/SCREEN-STATES.md) — every screen and state, with its copy
- [`docs/CONTEXT.md`](docs/CONTEXT.md) — domain glossary (Member, Salary, Spend, Recap, Archetype…). Read this before naming anything.
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — the backend design: Nginx, the Go BFF, the four services, RabbitMQ, Redis, Postgres, Keycloak, and the numbered `A*` decisions
- [`docs/IOS.md`](docs/IOS.md) — the iOS app's internals: sign-in, session, networking, the SwiftData cache, generated tokens, and the numbered `I*` decisions
- [`docs/design/design_handoff_my_money_recap/README.md`](docs/design/design_handoff_my_money_recap/README.md) — the design handoff: eleven canvases, 118 screens
- [`my-money-recap-ds/README.md`](my-money-recap-ds/README.md) — the design system: visual rules and component API
- [`my-money-recap/README.md`](my-money-recap/README.md) — how a rendered illustration becomes a shipped asset

## Single source of truth

| Thing | Lives in | Everything else is generated |
| --- | --- | --- |
| Design tokens | `my-money-recap-ds/src/tokens/tokens.json` | `tokens.css`, `css.generated.ts`, `dist/mmr.css` |
| Illustration artwork | `my-money-recap/assets/illustrations/*.webp` | `my-money-recap-ds/src/assets/illustrations.ts` |
| Product and design decisions | `docs/PRD.md` (numbered `D*` decisions) | — |
| Backend architecture decisions | `docs/ARCHITECTURE.md` (numbered `A*` decisions) | — |
| iOS app decisions | `docs/IOS.md` (numbered `I*` decisions) | — |
| Visual rules | `my-money-recap-ds/README.md` § Design rules | — |
| Screens | `docs/design/design_handoff_my_money_recap/` | renders (not in git) |
| Domain vocabulary | `docs/CONTEXT.md` | — |
| Sample ledger | the defaults in `my-money-recap-ds/src/screens/RecapScreens.tsx` | stories and design-sync previews |
