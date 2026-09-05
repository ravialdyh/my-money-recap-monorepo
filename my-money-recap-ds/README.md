# My Money Recap — design system package

Tokens and React components for My Money Recap, the adult personal-finance recap app in a kids' visual theme, extracted from the four
Wrapped-style artboards. Companion documents: `../my-money-recap/DESIGN.md` (the decisions) and
`src/tokens/tokens.json` (the values).

## Layout

```
src/
  tokens/        tokens.json (source of truth) → tokens.css + index.ts (generated)
  styles/        components.css → css.generated.ts (generated); useStyles() injects it once
  components/    one component per file, states via data-state / :active / :disabled
  icons/         one <Icon name=…> component; glyph drawings live in glyphs.tsx
  screens/       the four recap screens as components (RecapCoverScreen, RecapEarningScreen, RecapGoalScreen, RecapPersonalityScreen)
  stories/       CSF3 stories: one file per component + Screens/Monthly recap (the four artboards rebuilt)
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

The four recap illustrations ship as data URIs in `src/assets/illustrations.ts` (`ILLUSTRATIONS.cover`,
`.lemonade`, `.vending`, `.squirrel`), so stories and previews need no asset pipeline. For a production app
you may prefer to serve the WebP files from `../my-money-recap/` and pass their URLs to `<Illustration src=… />`.

## States

| Component | default | pressed | disabled | selected |
|---|---|---|---|---|
| Chip, StatChip, Button, DataRow, ArchetypePill | yes | `:active` or `state="pressed"` → scale 0.97 + brightness 0.92 (the "lightness −8%" rule) | `disabled` or `state="disabled"` → opacity 0.4 | ArchetypePill only: `selected` → 2px white ring, full opacity |
| PageDots, StreakDots | yes | n/a | n/a | active dot |

## Adding an icon

Draw the glyph in `src/icons/glyphs.tsx`, add its name to the `IconName` union and `GLYPHS` map in `src/icons/Icon.tsx`, and add a story in `src/stories/Icons.stories.tsx`. One `Icon` component keeps the Claude Design pane to a single icon card however many glyphs there are.

## Rebuilding generated files

```
npm run build:tokens   # tokens.json → tokens.css, tokens/index.ts, styles/css.generated.ts
npm run build:assets   # ../my-money-recap/ill-*.webp → src/assets/illustrations.ts
npm run typecheck
```

## Syncing to Claude Design

Open this folder in Claude Code and type `/design-sync` at the prompt yourself. The sync reads the
components and stories here and publishes them as cards under Design systems for your org.
