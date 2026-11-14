# Design tokens

The values behind every screen live in `tokens.json`: palette, colours, radii, spacing, sizes, the type ramp and
motion. `build-swift-tokens.mjs` turns them into `apps/ios/DesignSystem/Tokens.swift` (docs/IOS.md I5). The SwiftUI
components in `apps/ios/DesignSystem` are the only component set; `../docs/SCREEN-STATES.md` lists the screens. The
product decisions are in `../docs/PRD.md`; the visual rules are below.

## Changing a value

Edit `tokens.json`, then regenerate and commit both files:

```
npm run build:swift-tokens
```

`tools/hooks/check-generated.sh tokens` runs the same script and fails when `Tokens.swift` is stale. CI runs it as the
`tokens` job.

## Design rules

The rules every screen follows. The numbers behind them are tokens in `tokens.json`; the decisions that set them are the numbered `D*` rows of `../docs/PRD.md` §4.

- **One ground per screen, from four.** `recap` violet, `earning` yellow, `goal` coral, `personality` green. A screen's ground says what it is about; the same hue inside a card says what a number is about. Sky is a supporting hue for icons, wells and the Rides category, never a ground.
- **Yellow inverts the text.** It is the only ground light enough for ink text, so the earning screen uses ink type and the ink header chip. Every other ground uses white type.
- **Two fonts, split by job.** Fredoka for numbers, headlines and titles; Nunito for anything a person reads as a sentence.
- **Three radii.** `radius.card` 28, `radius.inset` 16 (a block nested inside a card: `InlineError`, the row press wash), `radius.pill` 999. Nothing in between.
- **No shadows.** Depth comes from overlap (an illustration hangs `space.overlap` 44pt into the card below it) and from ground-versus-card colour. The only strokes that look like shadows are rings: focus, invalid, selected, and the slider knob's 3pt white ring.
- **One story per screen.** One big number and at most two supporting facts; one asymmetric move (an illustration breaking a card edge) and otherwise a single-column stack.
- **Every screen has one hero illustration**, drawn from the style block in `../my-money-recap/docs/illustration-prompts.md`. No stand-ins drawn in code.
- **Shareable screens carry no page dots**, so a screenshot is clean.
- **Pressed and disabled are one rule each**: pressed = scale 0.97 and brightness 0.92 for 120ms; disabled = opacity 0.4. Selected is a 2pt white ring.
- **Tap targets are 44pt.** `Chip` and `ArchetypePill` are drawn smaller; mark one `tappable` wherever it is a control and the hit area grows without the pill changing. `PageDots` is the one exception: each dot is tappable in its own slice of the row (D71).
- **Motion is the `motion.*` tokens**: a screen enters in 320ms sliding up 24pt; the hero number counts up over 800ms starting 150ms after the card lands; bars grow over 600ms, staggered 80ms; a press is 120ms. With Reduce Motion on, every reveal becomes a 150ms fade.
- **Money is an exact decimal plus an ISO 4217 code** (PRD §7). The platform formatter prints it in the viewer's locale with the currency's own decimals, and nothing here knows how many decimals a currency has. `StatHero` and `AmountField` measure the printed string and pick the largest size that fits (80/56/40pt and 48/36/28pt, D71); whole amounts print without a zero fraction.
- **Contrast.** Secondary ink (`ink.500`) on white is 4.6:1, so it is for 14pt and up. White on coral or green passes only as large text; keep body copy on those grounds inside white cards.

## States

| Component | default | pressed | disabled | selected |
|---|---|---|---|---|
| Chip, StatChip, Button, DataRow, ArchetypePill | yes | scale 0.97 + brightness 0.92 (the "lightness −8%" rule) | opacity 0.4 | ArchetypePill only: 2pt white ring, full opacity |
| PageDots, StreakDots | yes | n/a | n/a | active dot |
