# Prompt for Claude Design, part 2: Home and the daily loop

Paste everything below the line into Claude Design, in the **same "Kids Money Recap" project** as the auth flow, as a **new file** (`Home and Core Features.dc.html`). Same project keeps the design-system binding and the uploaded artwork; a separate file keeps the auth canvas stable while this one grows.

The design system was rebuilt and re-synced on 2026-09-04 for exactly this part: `TextField`, `AmountField`, `Sheet`, `Banner`, `ListRow`, `Toggle`, `Slider`, `AvatarGrid`/`Avatar` are real components now, `Icon` carries eighteen glyphs including all eight Categories, and every Category has its own colour token. Nothing here should be drawn as a "NEW" placeholder.

---

Design the Home screen and the whole spend-logging loop for Kids Money Recap, as a multi-artboard canvas. This is part 2 of four; part 1 (`Auth and Onboarding`) ends on the first-run Home, and this file picks it up from there. Use the synced "Kids Money Recap" design system for every element.

## Product facts you must respect

- Users are adults (18+). The kids look is a visual theme only. Money words stay plain: "salary", "saving target", "spend". Never "pocket money", "allowance", "budget" or "transaction".
- Native iOS app, iPhone 390 × 844 artboards, portrait. Fonts Fredoka (headlines, numbers) and Nunito (everything else). Two radii only: cards 28, pills 999. No shadows anywhere; depth comes from white cards on a coloured ground and from illustrations overlapping the card top.
- Home uses the violet ground (`ground.recap`). **The Log-a-spend sheet is not a fifth ground** — it is a `Sheet`, a white surface with `radius.card` on its top corners, sitting over a dimmed Home.
- Primary button: ink pill. Secondary button: white pill with ink label. Pressed = scale 0.97 and brightness 0.92. Disabled = 40% opacity.
- Every tappable element is at least 44 × 44 pt. The month chip and the Settings chip are controls, so they take the `tappable` variant.
- All amounts are integers in the Currency's minor units. Maya's Currency is GBP, so two decimals, and amounts display as `£1,190`.

## The one sample month, and why the numbers are what they are

Every artboard in parts 2, 3 and 4 uses one Member and one ledger. Do not invent other figures.

**Maya · GBP · salary £2,400 · payday 25 · saving target 20% = £480 · Avatar Otter.**

September 2026 has fifteen spends totalling **£1,760**, so Money In £2,400 − Spent £1,760 = **Saved £640**:

| Date | Category | Note | Amount |
|---|---|---|---|
| 2 Sep | Bills | Phone and internet | £280 |
| 3 Sep | Snacks | Coffee and lunch | £60 |
| 5 Sep | Bills | Electricity | £220 |
| 6 Sep | Other | Locksmith | £20 |
| 8 Sep | Rides | Train to Brighton | £40 |
| 9 Sep | Snacks | Groceries top-up | £55 |
| 10 Sep | Clothes | Winter jacket | £180 |
| 11 Sep | Snacks | Bakery run | £45 |
| 21 Sep | Snacks | Dinner out | £50 |
| 22 Sep | Bills | Water | £120 |
| 23 Sep | Gifts | Ana birthday present | £120 |
| 27 Sep | Rides | Flight home | £420 |
| 28 Sep | Snacks | Airport sandwich | £40 |
| 29 Sep | Games | New release | £80 |
| 30 Sep | Toys | Puzzle | £30 |

Consequences you must keep consistent: nothing is dated 12 to 20 September, so the longest no-spend **Streak is 9 days**. The largest single spend is the **£420 flight**, 23.9% of Spent. The target-warning line is Salary − Target = **£1,920**, and September's £1,760 never crosses it — so **the Target Warning cannot fire in September**. Its artboard is an October screen; see the variants row.

**Today, for the main Home artboard, is Saturday 26 September 2026.** Spends dated up to and including the 26th total **£1,190**, so Money In £2,400, Spent £1,190, Saved £1,210, with five days left in the month counting today.

## Screens to draw, in one row, left to right, 80 px gap, titled exactly as below

**B1 · Home.** The screen part 1 lands on, now carrying a month of data. Top to bottom:
- `ChipRow` with a `tappable` month chip "September 2026" and a `tappable` "Settings" chip carrying the `settings` glyph.
- Greeting `Headline` size m "Hi, Maya" with the Otter `Avatar` at 56 pt on the right of the same row.
- Hero `Card`: `StatHero` eyebrow "Money in so far", value **£2,400**. Below it a `StatChipRow` with `StatChip` tone `saved` "Saved £1,210" and tone `spent` "Spent £1,190". The Spent chip is tappable and opens B6.
- Target `Card`: `CardTitle` "Saving target", right-aligned "£1,210 of £480", `ProgressBar` full (Saved is over target, so it clamps at 100%), `HelperText` "Keep spending under £146 a day and you'll make it."
- `Button` "Log a spend" (ink, `plus` glyph) and `Button` "See last month" (white, secondary, **disabled** — August has no data).
- Recap list heading "Past recaps" with a `HelperText` "Your first recap arrives on 1 October."

Annotate: "£146 = (Salary − Target − Spent) ÷ days remaining, floored. Days remaining counts today, so 5 on the 26th. The bar reads full all month because Saved starts at the whole salary on payday and only falls — flag if that reads wrong."

**B2 · Log a spend, empty.** Home dimmed behind a scrim, `Sheet` over it with the drag handle and title "Log a spend". Inside: an `AmountField` showing "£0.00" focused, the eight Category `Chip`s in a wrapping row (Snacks, Games, Toys, Rides, Clothes, Bills, Gifts, Other — each with its Category glyph and its own colour, none selected), a date row reading "Today, 26 September", an optional note `TextField` with placeholder "What was it? (optional)", and an ink `Button` "Save" **disabled**. Annotate: "Save enables only when amount > 0 and a Category is selected."

**B3 · Log a spend, filled.** Same sheet with the iOS numeric keypad shown at the bottom, amount "£12.40", the **Snacks** chip selected (inverse fill, its glyph in white), note "Coffee and lunch", Save enabled. The keypad covers the lower third; the sheet content sits above it.

**B4 · Log a spend, date picker.** The date row expanded into a wheel or calendar limited to the last 60 days and today. Today, 26 September, is selected; 27 September onward is greyed. `HelperText` "You can log up to 60 days back, never ahead."

**B5 · Push permission.** The iOS system alert over a dimmed Home, appearing after the first ever spend is saved. Title "Kids Money Recap would like to send you notifications", body "Recap ready, your saving-line warning, and an optional daily reminder." Buttons "Don't Allow" and "Allow". Annotate: "System alert, drawn for placement only — iOS owns this UI. It fires once, after the first Spend is saved, and never again."

**B6 · September spends.** Opened by tapping the Spent chip. Violet ground, back chevron top-left, `Headline` size m "September spends", a `SubCopy` "£1,190 so far, 11 spends", then a white `Card` holding a `DataRowList` of `DataRow`s, newest first: each row is an `IconWell` in the Category's tone with the Category glyph, the note as label, and the amount in Fredoka. Show 23 Sep Gifts £120, 22 Sep Bills £120, 21 Sep Snacks £50, 11 Sep Snacks £45, 10 Sep Clothes £180, 9 Sep Snacks £55, 8 Sep Rides £40. Annotate: "Rows carry the date as a small secondary line under the note."

**B7 · Spend detail.** Tapping a row. The same fields as the sheet, prefilled from the 10 Sep Clothes £180 jacket, presented on a `Sheet` titled "Edit spend": `AmountField` £180.00, Clothes chip selected, date "10 September 2026", note "Winter jacket". Ink `Button` "Save changes" and, beneath it, a `ListRow` "Delete this spend" with `tone="destructive"`.

**B8 · Delete confirm.** The iOS action sheet over B7: "Delete this spend? £180 on Clothes, 10 September. This changes September's recap." Destructive "Delete" and "Cancel". Annotate: "Editing or deleting a Spend in a past month immediately changes that month's Recap — the Recap is computed live, never stored."

## Second row: states and variants, same titles with a suffix

- **B1 · Home / before payday.** 20 September, payday not yet reached. Spends dated up to the 20th total **£900**. Hero eyebrow "Salary lands in 5 days", value "£0", chips "Saved −£900" (the minus is correct: Saved = Money In − Spent, and Money In is 0 until payday) and "Spent £900". Target card `ProgressBar` at 0 — the bar clamps at zero, it never shows a negative — with `HelperText` "Nothing to keep yet. Your salary lands on the 25th."
- **B1 · Home / no salary set.** The Member skipped salary and target in onboarding. The hero card becomes a prompt card "Set your monthly salary" with a `chevron`; the target card becomes "Set a saving target". Both keep the same height as the filled versions so the layout does not jump.
- **B1 · Home / target warning.** **October 2026**, the 27th — September's target was met, so the warning cannot appear there. Money In £2,400, Spent £2,090, Saved £310. A coral `Banner` with a dismiss control sits directly above the target card: "You've passed your saving line. £170 over." Target card `HelperText` "You're £170 past your line. Next month is a fresh start." Annotate: "Fires once per month, the first time Spent passes Salary − Target (£1,920). Editing or deleting spends never un-fires it. Dismiss lasts for the session only."
- **B1 · Home / with past recaps.** October, with the recap list populated: a `DataRowList` of `ListRow`s, newest first, label "September 2026", value "The Squirrel", chevron. "See last month" is now enabled.
- **B2 · Log a spend / write failed.** Save pressed, no connection. The sheet stays open with every value intact and a coral `HelperText` under the button: "Couldn't save. Check your connection and try again." Annotate: "Every write needs the network. A failed write never closes the sheet and never loses what was typed."
- **B6 · September spends / empty.** No spends yet this month. The card is replaced by a centred prompt "Nothing spent yet. Streak day 4." with the `plus` glyph button beneath.

## Third row: one flow artboard

A single wide artboard titled **"Flow · Home and the daily loop"** showing thumbnails of B1 to B8 connected with labelled arrows: A8 (part 1) → B1 "returning member"; B1 → B2 "Log a spend"; B2 → B3 "type an amount"; B3 → B4 "change the date"; B3 → B1 "Save · numbers animate up"; B3 → B5 "first ever spend"; B5 → B1 "Allow or Don't Allow"; B1 → B6 "tap the Spent chip"; B6 → B7 "tap a spend"; B7 → B8 "Delete"; B8 → B6 "confirmed"; B1 → "Settings (part 4)"; B1 → "Recap (part 3)" via "See last month". Note on the artboard: "Home numbers animate to their new values with `motion.reveal.number` after the sheet closes. Reduce Motion replaces every reveal with a 150 ms fade."

## Annotations to add beside each screen

Short sticky notes, not inside the artboard: the validation rules (amount > 0; a Category required; date within the last 60 days and never in the future; note 80 characters); which buttons are disabled and why; and the one computed sentence per card, with its formula.

## Do not

- Do not add a bottom tab bar, a navigation bar, or a header title. Screens use the chip row and a back chevron only.
- Do not use shadows, gradients, or any radius other than 28 and 999.
- Do not invent a fifth ground colour for the sheet or the spend list. White surfaces are `Sheet`; the spend list keeps the violet ground with a white `Card`.
- Do not add Categories. There are exactly eight and they never change.
- Do not add a savings jar, a "move money to savings" action, or a named goal. **Saved is computed, never logged.**
- Do not draw a Settings screen here — that is part 4. The Settings chip just points off-canvas.
- Do not show a recap screen here — that is part 3.
