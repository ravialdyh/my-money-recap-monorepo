# My Money Recap — Design System Extraction

Status: **Checkpoint 1 resolved.** Extraction plus confirmed design decisions. No v0 built here by choice: the flow continues in Claude Design, where this system is the base for the full end-to-end app.
Companion file: `../my-money-recap-ds/src/tokens/tokens.json` (the token source of truth). Resolved decisions are in §6.

## 0. Source and scope

- **Read from source, not screenshots.** The four artboards were pulled from the live canvas (version `1788501827-1d34`, the one you saved after your edits) and every inline style was tallied programmatically. Counts below are occurrences across the four files.
- **Screens covered:** `1 · Cover`, `2 · Earned + streak`, `3 · Goal + spending`, `4 · Money personality`. Your brief said "both screens" in one place and "1.Cover until 4.Money personality" in another. I read that as all four. Flag it if you meant a pair.
- **Your edits found in the live version** (kept, and reflected here): headline on screen 2 sits 10px below the chips instead of 20px; the archetype pill row on screen 4 has an 18px left inset.
- **Mode:** this is an extraction of a greenfield mockup set. Nothing here is a production codebase, so several states (pressed, disabled) do not exist in the source and are marked **proposed** rather than observed.

---

## 1. Step 2 — Extracted token set (as found)

### 1.1 Colour inventory

| Hex | Uses | Screens | Role as used |
|---|---:|---|---|
| `#FFFFFF` | 57 | all | Card surface, on-colour text, page dots, header chips, sparkle glyphs |
| `#1E1B3A` | 24 | all | Primary ink, inverse chip bg, primary button bg, page dots on light screens |
| `#F7C948` | 15 | 1, 2, 3 | Screen 2 nominal bg, earned chip, streak badge, toys bar, streak dots, coin gold |
| `#6B6890` | 13 | all | Secondary text (eyebrows, helper copy, muted amounts) |
| `#F0665A` | 6 | 1, 3, 4 | Screen 3 nominal bg, spent dot, snacks bar, "The splurge" eyebrow, Treat-Yourself pill |
| `#5B4FE8` | 6 | 2, 3, 4 | Screen 1 nominal bg, streak card bg, games bar, skateboard deck, Slow Burner pill |
| `#3DAA6A` | 3 | 1, 2, 4 | Screen 4 nominal bg, saved dot, saved bar, lemon leaf |
| `#FFE3E0` | 3 | 3, 4 | Coral track tint (goal track, snacks track, spent track) |
| `#C9F0D6` | 2 | 1, 4 | Green tint (saved chip, saved track) |
| `#EFE9FF` | 2 | 2, 3 | Violet tint (paw icon well, games track) |
| `#EFECF7` | 2 | 2 | Row divider |
| `#FFD6D1` | 1 | 1 | Spent chip tint — a second coral tint, differs from `#FFE3E0` |
| `#FFE3EA` | 1 | 2 | Lemonade icon well — a third coral/pink tint |
| `#DFF3FF` | 1 | 2 | Car-wash icon well (sky tint) |
| `#FFF3CC` | 1 | 3 | Yellow track tint (toys) |
| `#2E8FD1` | 1 | 2 (svg) | Droplet icon fill — **off-palette blue** |
| `#F0B429` | 1 | 2 (svg) | Lemon icon fill — near gold, not the token gold |
| `#E8A317` | 1 | 4 | Hustler pill, coin ring gold |
| `#C96A3B` | 1 | 4 | Squirrel pill (rust) |
| `#5153E5` | 1 | 1 | Screen 1 **actual** bg (sampled from the render) |
| `#FCCA3B` | 1 | 2 | Screen 2 actual bg (sampled) |
| `#F0604E` | 1 | 3 | Screen 3 actual bg (sampled) |
| `#41A76A` | 1 | 4 | Screen 4 actual bg (sampled) |

Observations:
- **Two palettes for the screen backgrounds, at extraction time.** The nominal brand hues (`#5B4FE8`, `#F7C948`, `#F0665A`, `#3DAA6A`) were used for data and chips, while the screen backgrounds used slightly shifted tones sampled from the GPT Image 2 renders. **Resolved:** grounds now use the nominal hues (see §6, D1). The sampled tones are recorded above for history only.
- **Coral had three tints** (`#FFE3E0`, `#FFD6D1`, `#FFE3EA`) doing the same job. **Resolved:** consolidated to `#FFE3E0`.
- **The droplet icon was off-palette** (`#2E8FD1`). **Resolved:** a fifth hue, sky, was added and harmonised in oklch (§6, D4).
- **Opacity is used instead of tints on colour backgrounds:** white at 0.28 (inactive streak dots), 0.45 and 0.35 (inactive page dots), 0.6 (unselected archetype pills), 0.85/0.9 (sub-copy on colour), 0.8 (helper copy on violet card).

### 1.2 Typography

Fonts (Google Fonts): **Fredoka** 500/600/700 for display, **Nunito** 600/700/800 for everything else. Fallbacks as written: Fredoka → Arial Rounded MT Bold → Nunito; Nunito → Trebuchet MS → Segoe UI.

| Size | Weight | Face | Line height | Tracking | Count | Where |
|---:|---:|---|---:|---:|---:|---|
| 80 | 700 | Fredoka | 1.0 | −0.02em | 1 | Hero number £36 (1) |
| 68 | 700 | Fredoka | 0.98 | −0.02em | 1 | "The Squirrel" (4) |
| 48 | 700 | Fredoka | 1.02 | −0.01em | 1 | Cover headline (1) |
| 40 | 700 | Fredoka | 1.04 | −0.01em | 1 | Goal headline (3) |
| 36 | 700 | Fredoka | 1.04 | −0.01em | 1 | Earning headline (2) — squeezed for space |
| 30 | 700 | Fredoka | 1.0 | — | 1 | Streak badge number (2) |
| 20 | 600 | Fredoka | 1.1 | — | 3 | Card titles (2, 3) |
| 18 | 600 | Fredoka | — | — | 4 | Row amounts (2), "£41" (3) |
| 16 | 700/800 | Nunito | 1.4 | — | 3 | Sub-copy (3, 4), button label (4) |
| 15 | 800 | Nunito | 1.3 | — | 3 | Eyebrow (1), intro line (4), splurge copy (3) |
| 14 | 800/700 | Nunito | 1.4 | — | 9 | Card eyebrows, helper copy, row labels, amounts |
| 13 | 800/700 | Nunito | — | — | 10 | Chips, bar labels, comparison labels, helper copy |
| 12 | 800 | Nunito | — | 0.08em | 8 | Header chips (uppercase), splurge eyebrow |
| 11 | 800 | Nunito | — | — | — | (previous archetype pills; superseded by 10) |
| 10 | 800 | Nunito | — | 0.02em | 4 | Archetype pills |
| 9 | 800 | Nunito | — | 0.1em | 1 | "DAYS" unit in streak badge |

Weight distribution: 800 ×33, 700 ×11, 600 ×6. Nunito 800 is the workhorse. Headline-to-body ratio on the cover is 80 ÷ 14 ≈ 5.7×, which is where the Wrapped feel comes from.

### 1.3 Spacing

Observed values (px): 2, 4, 5, 6, 8, 9, 10, 12, 14, 16, 18, 20, 22, 24, 28, 54–60.

| Use | Value |
|---|---|
| Screen inset, horizontal | 24 |
| Screen inset, top / bottom | 56 / 28 |
| Card inset | 22 (one card 24) |
| Card inset when an illustration overlaps the top | 54–60 |
| Illustration overlap into card | 42–46 (negative margin) |
| Chip padding | 8 × 14 (header), 9 × 14 (stat), 6 × 10 (archetype), 7 × 11 (older) |
| Gap between chips / dots | 6, 8 |
| Gap inside cards | 12, 14, 16 |
| Row padding (data rows) | 7 × 0 |
| Headline top margin | 10–22 (varies per screen) |
| Vertical section rhythm | 12, 14, 18, 22 |

Not on a strict grid. Even values dominate; 5, 7 and 9 appear as one-offs from fitting.

### 1.4 Radius, borders, elevation

| Token | Value | Count |
|---|---|---:|
| Pill / circle | 999px | 76 |
| Card | 28px | 6 |
| Divider | 1px solid `#EFECF7` | 2 |
| Selection ring | `box-shadow: 0 0 0 2px #FFFFFF` | 1 |
| Drop shadow | none | 0 |

Elevation is achieved with **overlap and colour**, not shadow: illustrations hang over card tops; cards are pure white on saturated ground.

### 1.5 Sizes

| Element | Value |
|---|---|
| Screen | 390 × 844 |
| Primary button height | 52 |
| Streak badge | 64 circle |
| Icon well | 30 circle, 16px glyph |
| Button icon | 20px stroke glyph, 2px stroke |
| Header / stat chip height | ~30–32 (padding-derived) |
| Goal progress track | 22 |
| Comparison track | 18 |
| Category track | 16 |
| Streak dot | 12 |
| Page dot | 6, active 22 × 6 |
| Status dot in chip | 8 |
| Illustration display widths | 310 (cover), 200 (lemonade), 168 (vending), 340 (squirrel) |

### 1.6 Iconography and illustration

- Icons: inline SVG, 16px filled glyphs in wells (paw, droplet, lemon) and one 20px stroked glyph (share). Two styles coexist: filled vs stroked.
- Illustrations: flat vector, no outlines, one darker shading tone per hue, white glint dots, solid single-colour ground. Generated with GPT Image 2 from the shared style block in `illustration-prompts.md`, backgrounds removed, delivered as transparent WebP at roughly 2× display size.
- Drawn SVG: the progress-bar markers, 64 × 30 each. `skateboard` was the original; `piggy` replaced it on the saving-target bar at Checkpoint 2 (PRD D24) and `skateboard` is kept for reference.

---

## 2. Step 2b — Design Read and five dials

```yaml
Design Read:
  artifact: mobile story-card set (Wrapped-style monthly recap), 4 screens, static
  audience: children roughly 8–13 with a parent looking over their shoulder; the share card also reaches friends
  visual-language: warm playful flat-illustration, chunky rounded type, one saturated ground per screen with white cards (Dribbble education-app family, not corporate fintech)
  mode: extension of an existing 4-screen mockup set (preserve)
  visual-variance: 6
  motion-intensity: 5
  information-density: 3
  asset-dependence: 8
  brand-fidelity: 9
```

What each dial does to the work:

- **Variance 6.** One asymmetric move per screen (illustration overlapping a card, vending machine breaking the card edge). Everything else sits on a stable single-column stack. Do not add off-grid moments beyond one per screen.
- **Motion 5 (proposed).** No motion exists in the mockups. Wrapped's genre expectation is sequenced reveals: number count-up, bars growing, card sliding under the illustration. Anything beyond that is out of register for kids' content.
- **Density 3.** One story number plus at most two supporting facts per screen. Any new element must displace one, not join.
- **Assets 8.** The screens fail without the illustrations. Every new screen needs one hero illustration from the same prompt style block. No CSS stand-ins.
- **Fidelity 9.** New work must look native to these four screens: same two fonts, same radii, same chip anatomy.

---

## 3. Step 3a — Four positioning questions per screen

| Screen | Narrative role | Viewing distance | Temperature | Capacity |
|---|---|---|---|---|
| 1 · Cover | Hero / opener. Sets the month and the one big number. | 30cm phone, glance | Celebratory, abundant | Comfortable. Room for the 310px piggy plus a 210px card. |
| 2 · Earned + streak | Data beat. Two facts: earnings breakdown and streak badge. | 30cm, reading | Sunny, industrious | **Tight.** Headline reduced to 36px and rows compressed to 7px to fit. The densest screen. |
| 3 · Goal + spending | Data beat with a hook. Goal progress is the emotional line; spending split is the honest one. | 30cm, reading | Warm-urgent (coral), a little mischievous | Full. Split card with illustration inside is at capacity. |
| 4 · Money personality | Closing / shareable. The one that leaves the app. | 30cm, but designed to be seen at arm's length in a chat thread | Triumphant, cosy | Comfortable. Largest illustration, smallest card. |

Consequence: the system has to tolerate one "dense" screen (2) without changing type roles. That is why the headline scale below has a compact step.

---

## 4. Step 3 — Design Decisions (expanded)

### 4.1 Semantic colour tokens

Primitives are named by hue; everything in the UI references a semantic alias. Full values in `../my-money-recap-ds/src/tokens/tokens.json`.

**Ground (one per screen)**
| Semantic | Meaning | Primitive |
|---|---|---|
| `color.ground.recap` | Cover / overview screens | `palette.violet.500` |
| `color.ground.earning` | Money in, what was kept, streaks | `palette.yellow.500` |
| `color.ground.goal` | Goals and spending | `palette.coral.500` |
| `color.ground.personality` | Archetype / share | `palette.green.500` |

Grounds are the same primitives as the data colours. A screen's ground says what the screen is about; the same hue inside a card says what a number is about. Sky (`palette.sky.500`) is deliberately **not** a ground: five grounds would dilute the one-colour-per-beat rhythm. It is a supporting hue for icons, wells and a fifth data category.

**Surfaces**
| Semantic | Primitive | Where |
|---|---|---|
| `color.surface.card` | `#FFFFFF` | Every white card |
| `color.surface.card.accent` | `palette.violet.500` | Streak card (violet on yellow) |
| `color.surface.chip.onGround` | `#FFFFFF` | Header chips on colour |
| `color.surface.chip.inverse` | `palette.ink.900` | Ink header chips, primary button |
| `color.surface.chip.highlight` | `palette.yellow.500` | "Maya's recap" chip, streak badge |
| `color.surface.well.*` | tints | Icon wells (violet, sky, coral) |
| `color.surface.track.*` | tints | Bar tracks per data category |

**Text**
| Semantic | Primitive | Contrast note |
|---|---|---|
| `color.text.primary` | `palette.ink.900` | On white ≈ 15:1 |
| `color.text.secondary` | `palette.ink.500` (`#6B6890`) | On white ≈ 4.6:1, AA for 14px+ only |
| `color.text.onGround` | `#FFFFFF` | On violet ≈ 5.6:1; on coral ≈ 3.1:1 (large text only); on green ≈ 3.3:1 (large text only) |
| `color.text.onGround.muted` | white @ 0.85 | Sub-copy |
| `color.text.onGround.dark` | `palette.ink.900` | Text on the yellow screen |
| `color.text.accent.coral` | `palette.coral.500` | "The splurge" eyebrow |

**Data (category colours carry meaning, not decoration)**
| Semantic | Primitive | Meaning |
|---|---|---|
| `color.data.earned` | yellow | Money in |
| `color.data.saved` | green | Kept |
| `color.data.spent` | coral | Money out |
| `color.data.category.snacks` | coral | Spending split |
| `color.data.category.games` | violet | Spending split |
| `color.data.category.toys` | yellow | Spending split |
| `color.data.category.rides` | sky | Spending split (added 2026-09-04) |
| `color.data.category.clothes` | green | Spending split (added 2026-09-04) |
| `color.data.category.bills` | ink-500 | Spending split (added 2026-09-04) |
| `color.data.category.gifts` | gold | Spending split (added 2026-09-04) |
| `color.data.category.other` | rust | Spending split (added 2026-09-04) |

The eight Categories are fixed (PRD D20) and no two share a hue. Snacks, Games and Toys keep the hues observed on the Goal artboard; the five added in the PRD build-out needed two new tints, `gold.100` `#FDEFD0` and `rust.100` `#F7E0D3`, derived to match the lightness of the existing 100s.
| `color.data.streak.active` | yellow | Filled streak dot |
| `color.data.streak.inactive` | white @ 0.28 | Empty streak dot |

**Archetype**
| Semantic | Primitive |
|---|---|
| `color.archetype.squirrel` | rust `#C96A3B` |
| `color.archetype.hustler` | gold `#E8A317` |
| `color.archetype.treatYourself` | coral |
| `color.archetype.slowBurner` | violet |

**Feedback / interaction (adopted at Checkpoint 1; only the selected states were drawn)**
| Semantic | Rule |
|---|---|
| `color.state.pressed` | Same hue, lightness −8% (oklch), 120ms |
| `color.state.disabled.surface` | `palette.ink.900` @ 0.12 on white; white @ 0.28 on colour |
| `color.state.disabled.text` | `palette.ink.500` @ 0.6 |
| `color.state.selected.ring` | 2px white ring (as on the Squirrel pill) |

### 4.2 Typography roles

| Role | Face | Size / LH / Tracking | Used for |
|---|---|---|---|
| `type.display.xl` | Fredoka 700 | 80 / 1.0 / −0.02em | The one hero number |
| `type.display.l` | Fredoka 700 | 68 / 0.98 / −0.02em | Archetype name |
| `type.headline.l` | Fredoka 700 | 48 / 1.02 / −0.01em | Cover headline, 2 lines max |
| `type.headline.m` | Fredoka 700 | 40 / 1.04 / −0.01em | Screen headline, 2 lines |
| `type.headline.m.compact` | Fredoka 700 | 36 / 1.04 / −0.01em | Screen headline when 3 lines are unavoidable |
| `type.title.m` | Fredoka 600 | 20 / 1.1 | Card titles |
| `type.title.s` | Fredoka 600 | 18 / 1.1 | Amounts in rows |
| `type.stat.badge` | Fredoka 700 | 30 / 1.0 | Number inside a badge |
| `type.body.m` | Nunito 700 | 16 / 1.4 | Sub-copy, button label (800) |
| `type.body.s` | Nunito 700 | 14 / 1.4 | Helper copy |
| `type.label.m` | Nunito 800 | 14 / 1.2 | Row labels, card eyebrows (secondary colour) |
| `type.label.s` | Nunito 800 | 13 / 1.2 | Chips, bar labels |
| `type.caption.upper` | Nunito 800 | 12 / 1.2 / 0.08em, uppercase | Header chips, coloured eyebrows |
| `type.caption.pill` | Nunito 800 | 10 / 1.2 / 0.02em | Archetype pills |
| `type.caption.unit` | Nunito 800 | 9 / 1.0 / 0.1em, uppercase | Unit under a badge number |

Rule: Fredoka only for numbers, headlines and titles. Nunito for anything a child has to read as a sentence.

### 4.3 Spacing

Base unit 4px with a 2px half-step for chips. Proposed scale, mapped to what the screens already use:

| Token | px | Observed use |
|---|---:|---|
| `space.1` | 4 | Micro gaps (2 rounds up) |
| `space.2` | 6 | Chip gap, dot gap |
| `space.3` | 8 | Chip gap, chip vertical padding |
| `space.4` | 12 | Card internal gap, section gap |
| `space.5` | 14 | Chip horizontal padding, card gap |
| `space.6` | 16 | Vertical rhythm |
| `space.7` | 22 | Card inset (normalise the one 24) |
| `space.8` | 24 | Screen horizontal inset |
| `space.9` | 28 | Screen bottom inset |
| `space.10` | 56 | Screen top inset |
| `space.overlap` | 44 | Illustration hang into a card (normalise 42/46) |

Off-scale values in the source (5, 7, 9, 10, 18, 20) are fitting artefacts. Round each to the nearest step on the next build.

### 4.4 Radius, elevation, motion

- **Radius:** two values only. `radius.card` 28, `radius.pill` 999. Nothing in between.
- **Elevation:** none via shadow. Depth comes from `space.overlap` and ground-vs-card colour. Keep it that way; a shadow would read as a different product.
- **Motion (adopted at Checkpoint 1, applies to the built app, not the static boards):**
  - `motion.enter.screen` 320ms `cubic-bezier(.2,.8,.2,1)`, card slides up 24px and fades.
  - `motion.reveal.number` count-up 800ms ease-out, starts 150ms after card lands.
  - `motion.reveal.bar` width from 0 to value, 600ms, same curve, staggered 80ms per bar.
  - `motion.press` scale 0.97, 120ms ease-out, on any tappable pill or button.
  - Honour `prefers-reduced-motion`: replace reveals with 150ms fades.

### 4.5 Component inventory with states

Legend: **obs** = observed in the mockups, **prop** = proposed rule, not drawn.

| Component | Anatomy | Default (obs) | Pressed (prop unless noted) | Disabled (prop) | Selected (obs unless noted) | Screens |
|---|---|---|---|---|---|---|
| **Screen canvas** | 390×844, ground colour, insets 56/24/28 | one ground per screen | n/a | n/a | n/a | all |
| **Header chip** | pill, caption.upper, 8×14 padding | 3 variants: white on ground, ink inverse, yellow highlight | −8% lightness, scale .97 | fill @ 0.4 | n/a, chips are labels | all |
| **Headline** | headline.l / m / m.compact, text-wrap balance | white on colour, ink on yellow | n/a | n/a | n/a | all |
| **Sub-copy** | body.m, onGround.muted | white @ 0.85–0.9 | n/a | n/a | n/a | 3, 4 |
| **Illustration slot** | flex-grow, centred, negative bottom margin | transparent WebP, hangs `space.overlap` into next card | n/a | n/a | n/a | all |
| **Card** | surface.card, radius.card, inset 22 | white | n/a (cards are static) | n/a | n/a | all |
| **Card, accent** | surface.card.accent (violet), white text | violet | n/a | n/a | n/a | 2 |
| **Card, split** | card with illustration column breaking top edge | white + vending machine | n/a | n/a | n/a | 3 |
| **Stat hero** | eyebrow label.m secondary + display.xl | "Money that came your way / £36" | n/a | n/a | n/a | 1 |
| **Stat chip** | pill, 8px status dot, label.s, 9×14 | tinted by data colour | scale .97 | fill @ 0.4, dot grey | n/a | 1 |
| **Data row** | icon well + label.m + title.s amount, divider below | white row | row bg ink @ 0.04 | text @ 0.4 | n/a | 2 |
| **Icon well** | 30 circle, tint, 16px filled glyph | violet / sky / coral tints | n/a | glyph @ 0.4 | n/a | 2 |
| **Streak badge** | 64 circle, stat.badge + caption.unit | yellow with ink | n/a | n/a | n/a | 2 |
| **Streak dot grid** | 15 columns, 12px dots, 6 gap | active yellow, inactive white @ 0.28 | n/a | n/a | active = filled (obs) | 2 |
| **Progress bar, goal** | 22 track, fill, marker illustration above fill end | coral fill on coral tint | n/a | track only | n/a | 3 |
| **Category bar** | label.s row (name / %) + 16 track | one data colour per row | n/a | fill @ 0.4 | n/a | 3 |
| **Comparison bar** | label + 18 track + value, two rows | green vs coral | n/a | n/a | n/a | 4 |
| **Eyebrow, coloured** | caption.upper in a data colour | "THE SPLURGE" in coral | n/a | n/a | n/a | 3 |
| **Primary button** | ink pill, 52 high, 20px stroked icon + body.m 800 | ink, white text | ink −8%, scale .97 | ink @ 0.32, no icon dim | n/a | 4 |
| **Archetype pill** | pill, caption.pill, archetype colour, white text | unselected @ 0.6 (obs) | scale .97 | @ 0.3 | 2px white ring, opacity 1 (obs) | 4 |
| **Page indicator** | 6px dots, active 22×6 | inactive @ 0.45 (0.35 on yellow) | n/a | n/a | active = elongated (obs) | 1, 2, 3 |
| **Helper text** | body.s secondary | "Swipe to see where it all went." | n/a | n/a | n/a | 1, 3 |
| **Sparkle glyph** | white 4-point SVG | decorative, part of illustration | n/a | n/a | n/a | 4 |

Added at the PRD §13 build-out, 2026-09-04. All nine are **obs** — they render in the synced project and were verified by the design-sync capture pass.

| Component | Anatomy | Default | Pressed | Disabled | Selected | Where |
|---|---|---|---|---|---|---|
| **Sheet** | white surface, `radius.card` on the top corners only, 36×5 drag handle, optional scrim | white on violet | n/a | n/a | n/a | Log a spend, Settings |
| **TextField** | label + 52px pill field on `surface.field`, Nunito 800 ink | field tint | n/a | opacity 0.4 | 2px violet ring on focus; 2px coral on invalid | Email, name, note |
| **AmountField** | `TextField` with a fixed symbol, Fredoka 30px value, minor units per ISO 4217 | as above | n/a | as above | as above | Salary, spend amount |
| **Slider** | goal-height track with a 28px ink knob | violet fill | n/a | opacity 0.4 | n/a | Saving Target |
| **Toggle** | 51×31 pill, 27px white knob | ink-100 off, violet on | scale .97 | opacity 0.4 | knob slides right | Settings |
| **Banner** | full-width coral card, dismiss = `plus` rotated 45° | coral, white text | n/a | n/a | n/a | Target Warning |
| **ListRow** | `DataRow` with no icon well; chevron, or a violet check when selecting | white row | 4% ink wash | opacity 0.4 | violet check | Settings, Currency |
| **AvatarGrid** | 3×2 of round 68px avatars | transparent WebP | scale .97 | n/a | 2px white ring on a violet halo | Onboarding 1, Profile |
| **Avatar** | one 56px round badge | transparent WebP | n/a | n/a | n/a | Home greeting |
| **Button, secondary** | white pill, ink label, same 52px | `surface.button-secondary` | −8% lightness, scale .97 | opacity 0.4 | n/a | Home, currency confirm |
| **Chip, category** | tinted pill + 16px Category glyph | the hue's own 100 tint | scale .97 | opacity 0.4 | ink inverse fill, white glyph | Log a spend |

Notes on states:
- Only two selected states are actually drawn: the elongated page dot and the ringed Squirrel pill. Every pressed and disabled value is a proposal derived from one rule each (−8% lightness, 0.4 opacity) so the set stays coherent.
- Disabled is unlikely to appear in a recap flow at all. It is here so the system has an answer, not because a screen needs it.
- Touch targets: header chips (~30px) and archetype pills (~24px) are below the 44px minimum. **Resolved 2026-09-04:** both take a `tappable` prop that lifts the hit area to 44px with a centred pseudo-element, so the drawn pill keeps its size. Use it wherever the pill is a control; use `asLabel` where it is a caption.

---

## 5. Appears on only one screen

Decide whether each is systemic or a one-off before the build.

| Element | Screen | My read |
|---|---|---|
| Hero stat (eyebrow + 80px number) | 1 | **Systemic.** Any future "big number" screen (total saved, year recap) needs it. Keep as `Stat hero`. |
| Two header chips including a yellow highlight chip | 1 | **One-off.** Highlight chip identifies the child; other screens use one or two neutral chips. Keep the variant, expect it once per flow. |
| "Swipe" helper line | 1 | **One-off.** Onboarding hint for the first card only. |
| Accent (violet) card | 2 | **Systemic candidate.** It is how the yellow screen keeps a secondary card legible. Rule: on a light ground, the second card may use `surface.card.accent`. |
| Data rows with icon wells | 2 | **Systemic.** Any itemised list (chores, purchases) reuses it. |
| Streak badge + 15-column dot grid | 2 | **Systemic pattern, one-off content.** The badge is reusable for any "N days" stat; the 30-dot grid is specific to a month view. |
| Blue droplet icon | 2 | **Resolved:** now `palette.sky.500` (`#3BA6F5`), a harmonised fifth hue. |
| Lemon icon gold (`#F0B429`) | 2 | **Resolved:** snapped to `palette.gold.500`. |
| Goal progress bar with illustrated marker | 3 | **Systemic.** Every goal screen wants it. The marker swaps per goal. |
| Category bars with percentage labels | 3 | **Systemic.** Spending split is a recurring recap beat. |
| Split card (illustration inside the card, breaking its top edge) | 3 | **One-off layout, worth keeping.** It is the only way to fit an illustration and dense data on one card. Use sparingly. |
| Coloured eyebrow ("THE SPLURGE") | 3 | **Systemic.** Cheap way to flag a callout inside a card. |
| Comparison bars (two rows) | 4 | **Systemic.** Saved vs spent will appear on any summary. |
| Primary button with icon | 4 | **Systemic.** The only button in the set defines the button style. |
| Archetype pill row with selection ring | 4 | **One-off in the recap, systemic in the app.** The same pill likely appears in a profile or history view. |
| Screen without page indicator | 4 | Intentional: the share card has no dots so screenshots are clean. Make that a rule for any shareable screen. |
| Dark ink text on ground | 2 | **Rule, not one-off.** Yellow is the only ground light enough for ink text; it also forces the ink header chip. Document as "yellow ground inverts the text colour". |

---

## 6. Decisions resolved at Checkpoint 1

| # | Decision | Outcome | Why |
|---|---|---|---|
| D1 | Ground colours | **Nominal brand hues are the tokens.** Grounds reference `palette.*.500`; the render-sampled tones were dropped. The four artboards were updated to match. | The set grows into a full end-to-end app in Claude Design. Most future screens have no hero illustration, so a ground that only exists to match one render cannot scale. One palette for grounds and data keeps every screen derivable from `tokens.json`. The illustrations are already transparent cut-outs, so they sit on the nominal hues with no visible edge. |
| D2 | Coral tints | **One tint, `#FFE3E0`.** | Most used of the three; the others were fitting drift. |
| D3 | Pressed / disabled | **One rule each:** pressed = same hue, oklch lightness −8%, scale 0.97, 120ms; disabled = opacity 0.4. | Consistent across every component; per-component values can be added later if a screen needs them. |
| D4 | Off-palette blue | **Add a fifth hue, `palette.sky.500` = `#3BA6F5`** (oklch L 0.70, C 0.15, h 245), tint `#DFF3FF`. Never a ground. | A kids-styled finance app needs more than four category colours (chores, categories, badges) and blue is the largest gap in the wheel between violet (h 279) and green (h 154). The original `#2E8FD1` was duller than its neighbours; the new value matches coral and green in lightness and chroma so it reads as family, not as an import. |
| D5 | Motion | **Carry the reveal choreography** in §4.4 as adopted tokens. | Wrapped-genre expectation; keeps the app version from being designed twice. |
| D6 | Scope | All four screens. | Confirmed. |

### Continuing in Claude Design

- The published canvas is the source of truth. Add new artboards there; keep one ground per screen from the four grounds, white cards, the two radii, the two fonts.
- Any new screen type (onboarding, add-a-chore, goal detail, parent view) should be able to name its ground, its card type, its type roles and its components from this document before drawing. If it cannot, that is a new component, and it belongs in §4.5 first.
- New illustrations: reuse the style block in `illustration-prompts.md` with the cover render as the style reference. As of the batch-2 renders, the set is complete for the whole product — nothing in Parts 2 to 4 needs a new one.

### Checkpoint 2, 2026-09-04 — built for PRD §13

The system was extended so the remaining flows can be drawn from real components rather than placeholders: nine components added (table in §4.5), `Icon` taken from 6 glyphs to 18, five Category colours and the sheet/scrim/field/banner/toggle surfaces added, and the four recap screens revised per D2, D23 and D24. Two defects surfaced by rendering were fixed — `CompareBars` scaled rows against the sum while documenting the largest (now an explicit `scale` prop), and `RecapGoalScreen` clipped past five `CategoryBar`s (now capped with a labelled roll-up, which is what density dial 3 in §2 implies). The font-inheritance gap in §5's notes is closed. Full record: `../my-money-recap-ds/.design-sync/NOTES.md`, third sync.

A fourth sync followed, driven by what building the part 2 and 3 canvases exposed: a secondary `Button` variant and Category `Chip` tones (both had been hand-styled on the canvas for want of a prop), an `Illustration` `maxHeight` so the tallest Archetype mascot stops clipping the pill row, and three defects in the recap screens that only measurement caught — `RecapGoalScreen` overflowing 844pt by 90, and two empty states made unreachable by default parameters swallowing `undefined`. Every recap screen and variant is now measured to fit with its page dots visible; `scripts/measure.tsx` is the harness.
