# Prompt for Claude Design, part 3: Recap and share

Paste everything below the line into Claude Design, in the **same "Kids Money Recap" project**, as a **new file** (`Recap and Share.dc.html`).

This is the one part where all four grounds appear. The four recap screens already exist as prop-driven components in the synced design system (`RecapCoverScreen`, `RecapEarningScreen`, `RecapGoalScreen`, `RecapPersonalityScreen`) and were revised on 2026-09-04 for exactly this brief — the Earned chip is gone, screen 2 is Kept-and-streak, screen 3 rides the `piggy` marker and colours all eight Categories, and screen 4 switches its mascot per Archetype. Start from those components and swap the data; do not redraw them from scratch.

---

Design the monthly Recap and its share flow for Kids Money Recap, as a multi-artboard canvas. This is part 3 of four. Parts 1 and 2 cover sign-in, onboarding, Home and spend logging; part 4 covers Settings. Use the synced "Kids Money Recap" design system for every element.

## Product facts you must respect

- Users are adults (18+). The kids look is a visual theme only. Money words stay plain.
- iPhone 390 × 844 artboards, portrait, except the share image, which is 1080 × 1920.
- **One ground per screen, and here they are fixed by the spec:** recap 1 violet, recap 2 yellow, recap 3 coral, recap 4 green. Yellow is the only ground that inverts text to dark ink and forces an ink header chip.
- Fredoka for headlines and numbers, Nunito for sentences. Two radii: cards 28, pills 999. No shadows.
- Recap screens 1 to 3 carry `PageDots` (4 dots). **Screen 4 has none** — it is the shareable one and screenshots must stay clean.
- A Recap is read-only and computed live on every open. There is no stored copy, no "refresh", and no editing from inside it.
- The Recap covers a **calendar month**, always. Payday only matters for when Money In appears inside it.

## The sample data

**Maya · GBP · salary £2,400 · payday 25 · saving target 20% = £480 · Avatar Otter.** The full dated ledger is in the part 2 prompt; these are its totals.

**September 2026 — the month the main row shows. Target met.**
- Money In £2,400 · Spent £1,760 · **Saved £640** · target £480 · **kept 133%**
- Longest no-spend Streak **9 days, 12 to 20 September** (nothing is dated in that window)
- Splurge **£420, the flight home on 27 September** — 23.9% of Spent, deliberately under the 25% line
- Category split of £1,760: Bills 35% (£620), Rides 26% (£460), Snacks 14% (£250), Clothes 10% (£180), Gifts 7% (£120), Games 5% (£80), Toys 2% (£30), Other 1% (£20)
- Archetype rules in order: Saved £640 ≥ target £480 and Streak 9 ≥ 7 → **Squirrel**

**October 2026 — the month the "missed" variants show. Target missed.**
- Money In £2,400 · Spent £2,090 · **Saved £310** · target £480 · **kept 65%**, £170 short
- Streak **4 days, 8 to 11 October**
- Splurge **£560, a winter coat** — 26.8% of Spent, over the 25% line
- Category split of £2,090: Bills 29% (£610), Clothes 28% (£590), Snacks 16% (£330), Rides 14% (£290), Games 7% (£150), Other 6% (£120)
- Archetype: Saved < target, and the splurge is ≥ 25% of Spent → **Treat-Yourself**

## Screens to draw, in one row, left to right, 80 px gap, titled exactly as below

**C1 · Recap 1, cover.** `RecapCoverScreen`. Violet ground. Chips "September 2026" and the yellow highlight chip "Maya's recap". `Headline` size l "Your September in money". The piggy-bank cover illustration at 310 pt hanging into a white `Card`. `StatHero` eyebrow "Money that came your way", value **£2,400**. `StatChipRow` with **only two** chips: "Saved £640" (green) and "Spent £1,760" (coral). `HelperText` "Swipe to see where it all went." `PageDots` 4, active 1. Annotate: "The Earned chip is gone — Salary is the only income in v1, so Money In and Earned would always be the same number."

**C2 · Recap 2, kept and streak.** `RecapEarningScreen`. **Yellow ground, so all text is dark ink and the first chip is the ink inverse one.** Chips "Kept" (inverse) and "September". `Headline` compact "You kept £640 of £2,400. That's 133% of your target." The saving-target jar illustration. White `Card`: `Eyebrow` "Kept against your target" over `CompareBars` with **`scale="max"`** — "Kept £640" in green at full width, "Target £480" in violet at 75%. Then a violet accent `Card`: `StreakBadge` 9 beside `CardTitle` "Longest no-spend streak" and the line "12 to 20 September. New badge unlocked.", with `StreakDots` for all 30 days, 12 to 20 filled. `PageDots` 4, active 2, `onLight`. Annotate: "`scale=\"max\"` matters here: Target is a benchmark, not a slice of a whole. Badge copy is 'New badge unlocked.' at 7 days or more, otherwise 'Beat it next month.'"

**C3 · Recap 3, target and spending.** `RecapGoalScreen`. Coral ground. Chips "Target" and "Spending" (inverse). `Headline` size m "133% of your saving target", `SubCopy` "You kept a third more than you set out to keep." White `Card`: `CardTitle` "Saving target", right-aligned "£640 of £480", `ProgressBar` full with the **`piggy` marker** riding its end, `HelperText` "Target met. £160 to spare." Then the split `Card` with the vending machine breaking its top edge: `Eyebrow` "Where £1,760 went", `CardTitle` "Mostly bills", then four `CategoryBar`s — Bills 35% ink, Rides 26% sky, Snacks 14% coral, Clothes 10% green — and beneath them one roll-up line "Gifts, Games, Toys, Other · 15%". Last, the coral `Eyebrow` "The splurge" over "£420 on a flight home. Your biggest single buy." `PageDots` 4, active 3. Annotate: "Four bars is the cap — measured. Five overflow the 844pt artboard and push the page dots off screen. The tail rolls into one labelled line rather than being dropped."

**C4 · Recap 4, archetype.** `RecapPersonalityScreen`. Green ground. Chip "Your money personality". The line "This month, Maya, you were", then `Headline` display "The Squirrel". `SubCopy` "You stashed more than you meant to, and you didn't crack once." The squirrel mascot at 340 pt hanging into a white `Card` holding `CompareBars` (default `scale="total"`) "Saved £640" green and "Spent £1,760" coral, then the ink `Button` with the `share` glyph, "Share your Squirrel". Below the card, the `ArchetypeRow` of four pills with Squirrel selected — 2 px white ring, the other three at 60% opacity. **No page dots.**

**C5 · Share sheet.** The iOS share sheet rising over a dimmed C4, with the generated image as its preview thumbnail, and the usual row of targets (Messages, Mail, Copy, Save Image). Annotate: "System UI, drawn for placement only. Rendering the image and presenting `UIActivityViewController` is the app's only job here."

**C6 · Share image.** A single 1080 × 1920 artboard, green ground, drawn as the poster the share sheet produces. The squirrel mascot large and centred, "The Squirrel" in Fredoka display, "Maya · September 2026" beneath it, and one white card with "Saved £640" and "Spent £1,760". A small "Kids Money Recap" wordmark at the bottom. Annotate: "**The share image must never show Salary or Money In.** Archetype, month, Display Name, Saved and Spent only. No page dots, no share button, no chrome."

**C7 · Past recaps.** Violet ground, back chevron, `Headline` size m "Past recaps". A white `Card` holding a `DataRowList` of `ListRow`s, newest first: "October 2026 / Treat-Yourself" and "September 2026 / The Squirrel", each with a chevron. `HelperText` "Months with no money in and no spends are hidden." Annotate: "Reached from Home. The current month is never listed — its recap becomes available on the 1st, when it becomes last month."

## Second row: states and variants, same titles with a suffix

- **C2 · Recap 2 / no target set.** The Member skipped the Saving Target. Headline drops the second sentence: "You kept £640 of £2,400." The compare card is replaced by a prompt: `CardTitle` "Set a target to see how you did" with `HelperText` "Pick a share of your salary to keep each month, and next month's recap will score it." The streak card is unchanged.
- **C3 · Recap 3 / target missed.** October. Headline "65% of your saving target", `SubCopy` "Close, but the coat tipped it." The bar is part-full in coral, `HelperText` "£170 short of your target." Split card: "Where £2,090 went", "Mostly bills", bars Bills 29% ink, Clothes 28% green, Snacks 16% coral, Rides 14% sky, roll-up "Games, Other · 13%". Splurge "£560 on a winter coat. Your biggest single buy."
- **C3 · Recap 3 / no spends.** A month with Money In and no Spends at all. "500% of your saving target", bar full, `HelperText` "Target met. £1,920 to spare." The split card's eyebrow reads "Where it went" and its title "Nothing spent this month", with the helper "You kept every penny that came in." in place of the bars, **and the splurge block is hidden entirely.** In the component this is automatic when `spentTotal` is 0; pass `splurge={null}`, never `undefined`.
- **C4 · Recap 4 / Treat-Yourself.** October. "The Treat-Yourself" with the cat mascot, blurb "One big treat tipped the month. Worth it?", `CompareBars` "Saved £310" and "Spent £2,090", button "Share your Treat-Yourself", the Treat-Yourself pill selected.
- **C4 · Recap 4 / Hustler.** Saved met the target but the streak was under 7 days. The bee mascot, blurb "You hit your number while still living a little."
- **C4 · Recap 4 / Slow Burner.** Target missed with no single big spend. The tortoise mascot, blurb "Lots of little spends added up. Sneaky."
- **C1 · Recap 1 / no salary.** No Salary was ever set, so Money In is £0 and there is no Archetype. The hero value is "£0" and the card carries a prompt "Set your monthly salary to get a recap" with a chevron into Settings. Annotate: "With no Salary in force there is no Archetype at all — screen 4 shows this prompt instead of a mascot."

## Third row: one flow artboard

A single wide artboard titled **"Flow · Recap and share"** with thumbnails of C1 to C7 and labelled arrows: B1 Home → C1 "See last month"; B1 Home → C7 "Past recaps"; C7 → C1 "tap a month"; C1 → C2 → C3 → C4 "swipe, horizontal paging"; C4 → C5 "Share your Squirrel"; C5 → C6 "the rendered image"; C4 → B1 "close". Note on the artboard: "Four horizontally paged screens. Reveal choreography per screen: the card slides up 24 px and fades over 320 ms, the number counts up over 800 ms starting 150 ms later, bars grow from zero over 600 ms staggered 80 ms apart. Reduce Motion replaces all of it with a 150 ms fade."

## Annotations to add beside each screen

Short sticky notes, not inside the artboard: the formula behind every number on that screen (kept% = Saved ÷ target amount, clamped 0 to 999; Streak = longest run of days with no spend, never crossing a month boundary; Splurge = largest single spend, ties to the earliest date); which content disappears in which empty state; and, on C6, the list of figures that must never appear.

## Do not

- Do not put page dots on screen 4 or on the share image.
- Do not show Salary or Money In anywhere on the share image.
- Do not add an "Earned" chip, a chores list, or any income other than the Salary.
- Do not use the skateboard marker. The saving-target bar rides the `piggy` glyph.
- Do not rename the screens' concepts: it is a "saving target", not a goal, and "Saved", not savings.
- Do not add a share button to screens 1, 2 or 3. Only screen 4 shares.
- Do not draw more than four `CategoryBar`s. Four plus a roll-up line is the pattern, and it is what fits.
- Do not make any recap element editable or tappable except the share button and the page swipe.
