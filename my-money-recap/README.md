# my-money-recap — design source

The Claude Design canvas, the illustration artwork, and the design spec that the
`my-money-recap-ds` package is built from.

## Layout

| Path | What lives here |
| --- | --- |
| `DESIGN.md` | Design system spec: type scale, colour, spacing, components, and the log of resolved decisions |
| `canvas/` | Claude Design canvas: `canvas.json` layout plus one `.dc.html` per artboard, and `preview.html` to view all four side by side |
| `assets/illustrations/` | Shipped artwork — transparent WebP cut-outs, roughly 2× display size |
| `assets/icons/` | `app-icon-1024.png`, the opaque square for the iOS build |
| `assets/illustrations.json` | Manifest: file, sampled background hue, dimensions, weight, source render |
| `renders/` | 4K originals the WebPs were cut from. **Not in git** — large, and kept locally |
| `scripts/` | Background-removal scripts that turn a render into a shipped WebP |
| `docs/` | Illustration prompts, the Claude Design prompts for parts 2–4, and the raw prompt log |
| `published/` | `my-money-recap.html`, the canvas published as a standalone artifact page, plus its preview |

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
cd ../my-money-recap-ds && npm run build:assets
```

Full prompt recipes and the style block are in [`docs/illustration-prompts.md`](docs/illustration-prompts.md).

## Viewing the canvas

Open `canvas/preview.html` in a browser for the four artboards side by side, or
`published/my-money-recap.html` for the interactive published canvas.

## Note on artboard asset paths

The `.dc.html` artboards reference artwork as `../assets/illustrations/ill-*.webp` so the repo
keeps one copy of each file. Claude Design resolves project-uploaded assets by bare filename, so
if you re-import an artboard there, re-point its `<img src>` at the uploaded asset.
