# My Money Recap: how to build with this system

**Root wrapper.** Every screen is `<Screen ground="recap" | "earning" | "goal" | "personality">` (390x844, one ground per screen). `Screen` owns the ground colour, the insets, the text colour (white on colour; dark ink on `earning`) and `font-family`. Styles inject themselves at runtime (`useStyles`); nothing else to import.

**The one exception to "a ground per screen" is `Sheet`.** Log a spend, Settings and every settings sub-screen are a white surface, not a fifth ground: put `<Sheet>` (bottom sheet, `radius.card` on the top corners, drag handle) or `<Sheet variant="full">` (a white card filling the screen) *on top of* the violet ground. Never invent a new ground colour.

**Styling idiom: tokens on plain elements, no class vocabulary.** The `mmr-*` classes are internal. For your own layout glue use plain `div`/`span` with inline styles and these custom properties (all defined in `tokens/` and `_ds_bundle.css`):

| Family | Real names |
|---|---|
| Grounds | `--mmr-color-ground-recap` (violet), `-earning` (yellow), `-goal` (coral), `-personality` (green) |
| Text | `--mmr-color-text-primary`, `--mmr-color-text-secondary`, `--mmr-color-text-on-ground`, `--mmr-color-text-on-ground-dark`, `--mmr-color-text-accent-coral` |
| Surfaces | `--mmr-color-surface-card`, `-card-accent`, `-sheet`, `-scrim`, `-field`, `-banner`, `-divider`, `-button-primary`, `-button-secondary`, `-toggle-on`, `-toggle-off` |
| Data | `--mmr-color-data-earned`, `-saved`, `-spent` |
| Categories | `--mmr-color-data-category-snacks` (coral), `-games` (violet), `-toys` (yellow), `-rides` (sky), `-clothes` (green), `-bills` (ink), `-gifts` (gold), `-other` (rust). Eight fixed Categories, no two share a hue. |
| Space | `--mmr-space-2` … `--mmr-space-10`, `--mmr-space-overlap` (illustration hang) |
| Radius | `--mmr-radius-card` (28px), `--mmr-radius-pill`. Two radii only — nothing in between. |
| Sizes | `--mmr-size-field-height`, `-toggle-width/-height/-knob`, `-slider-knob`, `-avatar-grid-cell` (68), `-avatar-inline` (56), `-touch-target-min` (44) |
| Type | `--mmr-font-family-display` (Fredoka, headlines and numbers), `--mmr-font-family-text` (Nunito, everything else) |
| Motion | `--mmr-motion-duration-press`, `--mmr-motion-easing-standard` |

Never hard-code hex; the palette lives in `--mmr-palette-*`. Depth comes from colour and overlap, never from shadow.

**Composition rules.** Header chips go in `ChipRow`; stat chips in `StatChipRow`; `DataRow`s and `ListRow`s in `DataRowList`; `ArchetypePill`s in `ArchetypeRow`. These four wrappers ship without cards (they are plain flex rows). `Illustration` hangs `--mmr-space-overlap` into the next `Card`, which needs `underIllustration`.

**Tap targets.** `Chip` and `ArchetypePill` are drawn below 44pt. Whenever either is a control rather than a label, pass `tappable` — it lifts the hit area to 44pt without changing the drawn pill. Use `asLabel` when it is purely a caption.

**Buttons come in two.** `<Button>` is the ink pill; `<Button variant="secondary">` is the white pill with an ink label, for the second action in a pair ("See last month", "Cancel" beside "Convert"). There is no third.

**Category selection chips.** The eight Spend Categories are `<Chip tone={CATEGORY_TONE[c]} selected={…} tappable><Icon name={c} />{CATEGORY_LABEL[c]}</Chip>` — each sits on its own hue tint and flips to the ink inverse fill, glyph and all, when selected. Do not hand-style them.

**Illustrations vary in aspect ratio.** Where one slot shows several (the four Archetype mascots), pass `maxHeight` as well as `width`, or a taller subject pushes what follows off the 844pt screen.

**Optional props that hide content take `null`, never `undefined`.** `RecapEarningScreen`'s `targetAmount` and `RecapGoalScreen`'s `splurge` have defaults, so `undefined` is replaced by the default and the empty state never appears. Pass `null` to mean "not set".

**Illustrations** are exported as data URIs on `ILLUSTRATIONS`: `cover`, `lemonade`, `vending`, `squirrel`, `envelope`, `jar`, `hustler`, `treat`, `slowburner`, and the six avatars, which have hyphenated keys read as `ILLUSTRATIONS['avatar-fox' | 'avatar-otter' | 'avatar-panda' | 'avatar-frog' | 'avatar-owl' | 'avatar-rabbit']`. There are no other illustrations; do not invent filenames.

**Icons.** All glyphs come from one component, `<Icon name="…" />`, with seventeen names: the eight Categories (`snacks`, `games`, `toys`, `rides`, `clothes`, `bills`, `gifts`, `other`) plus `paw`, `droplet`, `lemon` are 16px filled glyphs for `IconWell`; `share`, `settings`, `chevron`, `plus` are 20px stroked and take `currentColor`; `sparkle` is decorative; `piggy` is the 64x30 `ProgressBar` marker for the saving-target bar. A cross is `plus` rotated 45 degrees — there is no separate close glyph.

**Forms and settings.** `TextField` (label, `prefix`, `suffix`, `helper`, `state="default" | "focus" | "invalid" | "disabled"`), `AmountField` (currency-aware, minor units per ISO 4217 — GBP 2, IDR 0, BHD 3), `Slider` (the saving-target percentage), `Toggle` (settings switches), `Banner` (the coral Target Warning, `onDismiss` optional), `ListRow` (settings and currency rows: `value`, `trailing="chevron" | "check" | "none"`, `tone="destructive"`), `AvatarGrid` / `Avatar`.

**States are props:** `state="pressed" | "disabled"` on pressables, `selected` on `ArchetypePill` and `AvatarGrid`.

**Start from a finished screen.** `RecapCoverScreen`, `RecapEarningScreen` (kept + streak), `RecapGoalScreen` (target + spending), `RecapPersonalityScreen` are the four reference compositions, fully prop-driven; read their `.prompt.md` for the exact JSX and swap the data. All four are measured to fit 844pt with their page dots visible — `RecapGoalScreen` draws at most four `CategoryBar`s and rolls the rest into one line, which is what makes it fit. Raising `maxCategoryBars` overflows the artboard.

**Where the truth lives.** `styles.css` (imports `_ds_bundle.css`, which holds tokens plus every component rule), `README.md` component index, and `components/<group>/<Name>/<Name>.prompt.md` per component.

**Idiomatic snippet** (from a verified preview):

```tsx
<Screen ground="goal">
  <ChipRow><Chip asLabel>Target</Chip><Chip asLabel variant="inverse">Spending</Chip></ChipRow>
  <Headline size="m" style={{ marginTop: 20 }}>133% of your saving target</Headline>
  <SubCopy style={{ marginTop: 10 }}>You kept a third more than you set out to keep.</SubCopy>
  <Card style={{ marginTop: 22 }}>
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
      <CardTitle>Saving target</CardTitle>
      <span style={{ fontWeight: 800, color: 'var(--mmr-color-text-secondary)' }}>£640 of £480</span>
    </div>
    <div style={{ paddingTop: 14 }}><ProgressBar value={640} max={480} marker={<Icon name="piggy" />} /></div>
    <HelperText>Target met. £160 to spare.</HelperText>
  </Card>
  <div style={{ marginTop: 18 }}><PageDots count={4} active={2} /></div>
</Screen>
```
