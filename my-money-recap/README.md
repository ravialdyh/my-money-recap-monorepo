# my-money-recap — illustration source

The artwork the design system ships, and the scripts and prompts that produce it. Everything else
about the product lives elsewhere: the rules in `../docs/PRD.md`, the drawings in
`../docs/design/design_handoff_my_money_recap/`, the visual rules and components in `../my-money-recap-ds/`.

## Layout

| Path | What lives here |
| --- | --- |
| `assets/illustrations/` | Shipped artwork — transparent WebP cut-outs, roughly 2× display size; the design system inlines them (`../my-money-recap-ds/README.md`, "Rebuilding generated files") |
| `assets/icons/` | `app-icon-1024.png`, the opaque square for the iOS build once the app icon is wired into the asset catalog |
| `assets/illustrations.json` | Manifest: file, sampled background hue, dimensions, weight, source render |
| `renders/` | 4K originals the WebPs were cut from; untracked (`.gitignore`), present locally after a render run |
| `scripts/` | Background-removal scripts that turn a render into a shipped WebP |
| `docs/illustration-prompts.md` | The style block and one subject block per illustration, plus the recipe for adding a new one |

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
