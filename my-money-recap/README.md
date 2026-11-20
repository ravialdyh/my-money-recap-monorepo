# my-money-recap

The app's illustrations, and the prompts and scripts that make them.

## From a render to the app

```mermaid
flowchart LR
    prompt["Prompt"]
    render["4K render"]
    script["Cut-out script"]
    webp["WebP asset"]
    app["iOS asset catalog"]
    prompt --> render
    render --> script
    script --> webp
    webp -- "by hand" --> app
```

| Path | What lives here |
| --- | --- |
| `assets/illustrations/` | The shipped artwork: transparent WebP cut-outs, about twice display size |
| `assets/illustrations.json` | One entry per file: background hue, size, weight, source render |
| `assets/icons/` | `app-icon-1024.png`, the square app icon |
| `renders/` | The 4K originals, kept out of git |
| `scripts/` | The two background-removal scripts |
| `docs/illustration-prompts.md` | The style block, one prompt per illustration, and how to add one |

## Run it

From this folder. Each script writes the WebP into `assets/illustrations/` and prints the entry to paste into
`assets/illustrations.json`.

```bash
python3 scripts/cut_flat_bg.py renders/5-envelope.png envelope 640
```

```bash
python3 scripts/cut_avatars.py
```

Then add the WebP to `apps/ios/Resources/Assets.xcassets` by hand.
