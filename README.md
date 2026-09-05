# My Money Recap

A personal money tracker for adults, styled as a playful kids' app, that turns one month of
income, spending and saving into a Wrapped-style monthly recap.

This repository is the monorepo for it: design source, design system, apps and infrastructure.

## Layout

| Path | What lives here |
| --- | --- |
| `docs/` | Product-level docs: the PRD and the domain glossary |
| `my-money-recap/` | Design source: the Claude Design canvas, artwork, design tokens |
| `my-money-recap-ds/` | `@my-money-recap/ui` — React design system built from the canvas |
| `apps/` | Runnable products (mobile, web, backend). Empty for now |
| `packages/` | Shared libraries beyond the design system. Empty for now |
| `infra/` | Deployment and infrastructure. Empty for now |

It is an npm workspace: `my-money-recap-ds`, `apps/*` and `packages/*` are members, so one
`npm install` at the root wires everything up and there is a single `package-lock.json`.

## How the pieces connect

```
my-money-recap/                    my-money-recap-ds/
  canvas/*.dc.html      ──design──►    src/components, src/screens
  assets/illustrations/ ──build───►    src/assets/illustrations.ts  (base64 data URIs)
  DESIGN.md + tokens.json ─spec───►    src/tokens/tokens.json
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
`build:tokens` regenerates `tokens.css`, `tokens/index.ts` and `dist/mmr.css` from `tokens.json`.
CI runs both and fails if either output is stale, so generated files can never drift from source.

## Docs

- [`docs/PRD.md`](docs/PRD.md) — product requirements
- [`docs/CONTEXT.md`](docs/CONTEXT.md) — domain glossary (Member, Salary, Spend, Recap, Archetype…). Read this before naming anything.
- [`my-money-recap/DESIGN.md`](my-money-recap/DESIGN.md) — design system spec and the log of resolved decisions
- [`my-money-recap-ds/README.md`](my-money-recap-ds/README.md) — component API

## Single source of truth

| Thing | Lives in | Everything else is generated |
| --- | --- | --- |
| Design tokens | `my-money-recap-ds/src/tokens/tokens.json` | `tokens.css`, `tokens/index.ts`, `dist/mmr.css` |
| Illustration artwork | `my-money-recap/assets/illustrations/*.webp` | `my-money-recap-ds/src/assets/illustrations.ts` |
| Product decisions | `docs/PRD.md` (numbered `D*` decisions) | — |
| Domain vocabulary | `docs/CONTEXT.md` | — |
| Sample ledger | the defaults in `my-money-recap-ds/src/screens/RecapScreens.tsx` | stories and design-sync previews |
