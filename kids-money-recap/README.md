# kids-money-recap — design source

The Claude Design canvas, the illustration artwork, and the design spec that the
`kids-money-recap-ds` package is built from.

## Layout

| Path | What lives here |
| --- | --- |
| `DESIGN.md` | Design system spec: type scale, colour, spacing, components, and the log of resolved decisions |
| `tokens.json` | Design token source that `DESIGN.md` documents |
| `canvas/` | Claude Design canvas: `canvas.json` layout plus one `.dc.html` per artboard, and `preview.html` to view all four side by side |
| `assets/illustrations/` | Shipped artwork — transparent WebP cut-outs, roughly 2× display size |
| `assets/icons/` | `app-icon-1024.png`, the opaque square for the iOS build |
| `assets/illustrations.json` | Manifest: file, sampled background hue, dimensions, weight, source render |
| `renders/` | 4K originals the WebPs were cut from. **Not in git** — large, and kept locally |
| `scripts/` | Background-removal scripts that turn a render into a shipped WebP |
| `docs/` | Illustration prompts, the Claude Design prompts for parts 2–4, and the raw prompt log |
| `published/` | `kids-money-recap.html`, the canvas published as a standalone artifact page, plus its preview |

## Turning a render into a shipped asset

Run from this folder. Both scripts print a JSON block to paste into
`assets/illustrations.json`, and write the WebP into `assets/illustrations/`.

```bash
python3 scripts/cut_flat_bg.py renders/5-envelope.png envelope 640
```

```bash
python3 scripts/cut_avatars.py
```

Then inline the new artwork into the design system:

```bash
cd ../kids-money-recap-ds && npm run build:assets
```

Full prompt recipes and the style block are in [`docs/illustration-prompts.md`](docs/illustration-prompts.md).

## Viewing the canvas

Open `canvas/preview.html` in a browser for the four artboards side by side, or
`published/kids-money-recap.html` for the interactive published canvas.

## Known duplication

`tokens.json` here is a snapshot of the canvas tokens. `kids-money-recap-ds/src/tokens/tokens.json`
is the same set plus 49 tokens added while building the design system, and is the
source of truth for anything built. Reconcile or drop this copy when convenient.
