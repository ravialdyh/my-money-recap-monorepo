# My Money Recap — design system package

Tokens and React components for My Money Recap, the adult personal-finance recap app in a kids' visual theme, built for the eleven design
canvases in `../docs/design/design_handoff_my_money_recap/`. The product decisions are in `../docs/PRD.md`; the visual rules are
below; the values are in `src/tokens/tokens.json`.

## Layout

```
src/
  tokens/        tokens.json (source of truth) → tokens.css (generated)
  styles/        components.css → css.generated.ts (generated); useStyles() injects it once
  components/    one component per file, states via data-state / :active / :disabled
  icons/         one <Icon name=…> component; glyph drawings live in glyphs.tsx
  screens/       the four recap screens as components (RecapCoverScreen, RecapEarningScreen, RecapGoalScreen, RecapPersonalityScreen)
  money.ts       formatMoney and friends: the platform formatter, nothing hand-written
  index.ts       public exports
```

## Using it

```tsx
import { Screen, Headline, Card, StatHero } from '@my-money-recap/ui';

<Screen ground="recap">
  <Headline size="l">Your September in money</Headline>
  <Card underIllustration>
    <StatHero eyebrow="Money that came your way" value={36} animate />
  </Card>
</Screen>
```

Styles are injected by the components themselves (`useStyles`), so no CSS import is required.
If you prefer a stylesheet, import `@my-money-recap/ui/tokens.css` and `src/styles/components.css`.
Fonts: Fredoka and Nunito from Google Fonts, loaded by the injected stylesheet.

Every illustration ships as a data URI in `src/assets/illustrations.ts` (`ILLUSTRATIONS.cover`, `.lemonade`,
`.vending`, `.squirrel`, `.envelope`, `.jar`, `.hustler`, `.treat`, `.slowburner` and the six `avatar-*`), so
previews need no asset pipeline. For a production app
you may prefer to serve the WebP files from `../my-money-recap/` and pass their URLs to `<Illustration src=… />`.

## Design rules

The rules every screen follows. The numbers behind them are tokens in `src/tokens/tokens.json`; the decisions that set them are PRD D1–D71.

- **One ground per screen, from four.** `recap` violet, `earning` yellow, `goal` coral, `personality` green. A screen's ground says what it is about; the same hue inside a card says what a number is about. Sky is a supporting hue for icons, wells and the Rides category, never a ground.
- **Yellow inverts the text.** It is the only ground light enough for ink text, so the earning screen uses ink type and the ink header chip. Every other ground uses white type.
- **Two fonts, split by job.** Fredoka for numbers, headlines and titles; Nunito for anything a person reads as a sentence.
- **Three radii.** `radius.card` 28, `radius.inset` 16 (a block nested inside a card: `InlineError`, the row press wash), `radius.pill` 999. Nothing in between.
- **No shadows.** Depth comes from overlap (an illustration hangs `space.overlap` 44px into the card below it) and from ground-versus-card colour. The only `box-shadow`s in the CSS are rings: focus, invalid, selected.
- **One story per screen.** One big number and at most two supporting facts; one asymmetric move (an illustration breaking a card edge) and otherwise a single-column stack.
- **Every screen has one hero illustration**, drawn from the style block in `../my-money-recap/docs/illustration-prompts.md`. No CSS stand-ins.
- **Shareable screens carry no page dots**, so a screenshot is clean.
- **Pressed and disabled are one rule each**: pressed = scale 0.97 and brightness 0.92 for 120ms; disabled = opacity 0.4. Selected is a 2px white ring.
- **Tap targets are 44pt.** `Chip` and `ArchetypePill` are drawn smaller; pass `tappable` wherever one is a control and the hit area grows without the pill changing. `PageDots` is the one exception: each dot is tappable in its own slice of the row (D71).
- **Motion is the `motion.*` tokens**: a screen enters in 320ms sliding up 24px; the hero number counts up over 800ms starting 150ms after the card lands; bars grow over 600ms, staggered 80ms; a press is 120ms. With Reduce Motion on, every reveal becomes a 150ms fade.
- **Money is an exact decimal plus an ISO 4217 code** (PRD §7). The platform formatter prints it in the viewer's locale with the currency's own decimals, and nothing in this package knows how many decimals a currency has. `StatHero` and `AmountField` measure the printed string and pick the largest size that fits (80/56/40px and 48/36/28px, D71); whole amounts print without a zero fraction.
- **Contrast.** Secondary ink (`ink.500`) on white is 4.6:1, so it is for 14px and up. White on coral or green passes only as large text; keep body copy on those grounds inside white cards.

## States

| Component | default | pressed | disabled | selected |
|---|---|---|---|---|
| Chip, StatChip, Button, DataRow, ArchetypePill | yes | `:active` or `state="pressed"` → scale 0.97 + brightness 0.92 (the "lightness −8%" rule) | `disabled` or `state="disabled"` → opacity 0.4 | ArchetypePill only: `selected` → 2px white ring, full opacity |
| PageDots, StreakDots | yes | n/a | n/a | active dot |

## Adding an icon

Draw the glyph in `src/icons/glyphs.tsx` and add it to the `GLYPHS` map in `src/icons/Icon.tsx`; `IconName` and `ICON_NAMES` derive from that map. Show it in `.design-sync/previews/Icon.tsx`. One `Icon` component keeps the Claude Design pane to a single icon card however many glyphs there are.

## Rebuilding generated files

```
npm run build:tokens   # tokens.json → tokens.css, styles/css.generated.ts, dist/mmr.css
npm run build:assets   # ../my-money-recap/assets/illustrations/ill-*.webp → src/assets/illustrations.ts
npm run typecheck
```

## Syncing to Claude Design

Open this folder in Claude Code and type `/design-sync` at the prompt yourself. The sync reads the
components, `.design-sync/docs/` and `.design-sync/previews/` here and publishes them as cards under Design systems for your org.
