# Kids Money Tracker

Monorepo for **Kids Money Recap** — a personal money tracker for adults, styled as a
playful kids' app, that turns one month of income, spending and saving into a
Wrapped-style monthly recap.

## Layout

| Path | What lives here |
| --- | --- |
| `docs/` | Product-level docs: the PRD and the domain glossary |
| `kids-money-recap/` | Design source: the Claude Design canvas, artwork, design tokens |
| `kids-money-recap-ds/` | `@kids-money-recap/ui` — React design system built from the canvas |
| `apps/` | Runnable products (mobile, web, backend). Empty for now |
| `infra/` | Deployment and infrastructure. Empty for now |

## How the pieces connect

```
kids-money-recap/                    kids-money-recap-ds/
  canvas/*.dc.html      ──design──►    src/components, src/screens
  assets/illustrations/ ──build───►    src/assets/illustrations.ts  (base64 data URIs)
  DESIGN.md + tokens.json ─spec───►    src/tokens/tokens.json
```

The design system reads the WebP artwork straight out of `kids-money-recap/assets/illustrations/`
via a relative path, so the two folders must stay siblings inside this repo.

## Common commands

Regenerate the design system's inlined artwork after new illustrations land:

```bash
cd kids-money-recap-ds && npm install && npm run build:assets
```

Rebuild tokens, then typecheck:

```bash
cd kids-money-recap-ds && npm run build:tokens && npm run typecheck
```

## Docs

- [`docs/PRD.md`](docs/PRD.md) — product requirements
- [`docs/CONTEXT.md`](docs/CONTEXT.md) — domain glossary (Member, Salary, Spend, Recap, Archetype…). Read this before naming anything.
- [`kids-money-recap/DESIGN.md`](kids-money-recap/DESIGN.md) — design system spec and the log of resolved decisions
- [`kids-money-recap-ds/README.md`](kids-money-recap-ds/README.md) — component API
