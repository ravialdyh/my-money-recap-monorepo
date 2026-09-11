# Handoff: My Money Recap — full app design (4 canvases, 62 screens)

Read this first. Then `GAPS.md` (what is missing or contradictory — the reason this package exists), then `PRD.md`, then `DESIGN_NOTES_VERBATIM.md`.

## Overview

My Money Recap is a personal money tracker for adults with a playful kids'-app visual style. A Member sets a monthly **Salary**, **Payday** and percentage **Saving Target**, logs each **Spend** by hand, and gets a four-screen Wrapped-style **Recap** each calendar month. Design is done for the whole v1 product across four canvases:

| Canvas | Prefix | File | Screens |
|---|---|---|---|
| Auth and Onboarding | A | `design-files/Auth and Onboarding.dc.html` | A0–A8 + 7 state variants + flow map |
| Home and the daily loop | B | `design-files/Home and Core Features.dc.html` | B1–B8 + 6 variants + flow map |
| Recap and Share | C | `design-files/Recap and Share.dc.html` | C1–C7 + 7 variants + flow map |
| Settings and Account | D | `design-files/Settings and Account.dc.html` | D1–D12 + 5 variants + flow map |

Two purposes for this package: (1) **complete the spec** — find every state the PRD and the canvases leave undefined and decide it; (2) **implement** the app (native iOS 17+, SwiftUI per PRD D11) once the spec is complete.

> **Status, 2026-09-11.** The remaining 56 states are drawn in [`design-files/2026-09-11/`](design-files/2026-09-11/README.md) (live: <https://claude.ai/code/artifact/c2140f85-a23f-4979-a0a5-b1bdd1d2947c>). **Nineteen screens in the four canvases below are superseded by that folder and must not be built from:** A6, A8 first run, A8 skipped, Flow A · B1 before payday, B1 no salary set, B1 preview card, B6, B6 empty, Flow B · C1, C1 no salary (retired — now the Partial Recap P1–P4), C2, C2 no target, C3, C3 target missed, C3 no spends, C4 and its three variants, Flow C · D3, D4, D9 sent, Flow D. `SCREEN-STATES.md` marks each with "was REDRAW". Everything else in these four canvases still stands.
>
> **Status, 2026-09-08.** Purpose (1) is done: the design review closed all 57 items in `GAPS.md` and the decisions live in `../../PRD.md` (D41–D67) and `../../SCREEN-STATES.md`. **Read those two, not this package**, for anything you intend to build. Everything here — this file, `GAPS.md`, `DESIGN_NOTES_VERBATIM.md`, `PRD.md` (a pointer stub) and the canvases — is the record of how the design got here.

## About the design files

The `.dc.html` files are **design references built in HTML** on top of the real `@my-money-recap/ui` React design system. They are not production code to ship. Each screen is a `<div data-screen-label="…">` wrapping a 390×844 `MMR.Screen`. Yellow note boxes under each screen carry the behavioural rules; those notes are extracted verbatim into `DESIGN_NOTES_VERBATIM.md` so you do not have to parse the HTML. Some screens are live (sliders, fields, chips respond to input; a `today` tweak on the Home canvas changes the date).

> **The logic class at the bottom of each file (`class Component extends DCLogic`) is NOT a reference implementation of the formulas — do not copy from it.** An earlier version of this paragraph said it was, and that sentence sent readers to broken arithmetic. Verified 2026-09-08: Streak, kept %, Archetype, Splurge and the Category roll-up are computed **nowhere** in the canvases — `archetype="squirrel"` and `splurge="£420 on a flight home…"` are hand-written literal attribute values. The Home canvas hardcodes a 30-day month (`daysLeft = 30 - d + 1`), never computes days-in-month, omits `min(Payday, days in month)`, omits the Saved clamp, and clamps its own `today` prop to 21–30 so the September sample keeps working. **`../../PRD.md` §8 is the only source of the formulas.** What the canvases are good for is layout, copy and composition — not arithmetic.

The design-system bundle (`_ds/…/_ds_bundle.js`, `styles.css`) is **not** included; the HTML will not render as a local copy without it. **To see the designs, open the project itself:** <https://claude.ai/design/p/5f615747-48cd-4708-9369-becee9793f4e>, which is where they are live and current. Component and token names below refer to that published library and are the ones to use.

## Fidelity

**High-fidelity.** Colours, type, spacing, radii and copy are final and come from the design system tokens. Recreate pixel-exact using the design system's own SwiftUI/React equivalents. Exceptions drawn "for placement only" (iOS-owned UI): push-permission alert (B5), action sheets (B8, D12), share sheet (C5), numeric keyboard (B3), date wheel (B4, D8 time). Use the native controls there.

## Design system (binding)

Library `@my-money-recap/ui` 0.1.0. Tokens are CSS custom properties `--mmr-*`; never hard-code hex.

- **Grounds** (one per screen): `ground.recap` violet (Home, auth, onboarding, recap cover, settings backdrop), `ground.earning` yellow (Recap 2 — text inverts to dark ink), `ground.goal` coral (Recap 3), `ground.personality` green (Recap 4). No fifth ground.
- **Sheet** is the only non-ground surface: white, `radius.card` (28) top corners, drag handle. `Sheet variant="full"` = white card filling the screen (Settings and all sub-screens, Spend detail).
- **Type:** Fredoka 700 for headlines and numbers (`--mmr-font-family-display`), Nunito 700/800 for everything else (`--mmr-font-family-text`). Headline sizes l / m / compact. StatHero value 80px Fredoka. Yellow notes/canvas chrome are not part of the product.
- **Radii:** 28 (card) and pill. Nothing in between. **No shadows**; depth from colour and overlap.
- **Spacing:** `--mmr-space-2…10`; card stack rhythm on Home is 18 / 12 / 12 / 20.
- **Sizes:** field height 52, button 52 full-width pill, chip 30 (44 hit area when `tappable`), archetype pill 24, avatar grid cell 68 (88 hit), inline avatar 56, min touch 44, StreakBadge 64, ProgressBar 22, CategoryBar 16, CompareBars 18.
- **Data colours:** earned / saved (green) / spent (coral). Eight fixed Category hues: Snacks coral, Games violet, Toys yellow, Rides sky, Clothes green, Bills ink, Gifts gold, Other rust.
- **Buttons:** `Button` (ink pill) and `Button variant="secondary"` (white pill, ink label). Disabled = 40% opacity. Pressed = scale 0.97, brightness 0.92.
- **Icons:** one `Icon` component, 17 names. Cross = `plus` rotated 45°.
- **Illustrations:** cover, lemonade, vending, squirrel, envelope, jar, hustler, treat, slowburner, six avatars (fox, otter, panda, frog, owl, rabbit).
- **Motion:** `motion.duration.press`, `motion.easing.standard`; number reveals on Home after save; Reduce Motion → 150ms fades.

## Screen index

Each entry: purpose · composition · notable rules. Exact copy is in the PNG and the HTML. Variants are listed under their parent.

### A · Auth and onboarding (violet ground throughout)

- **A0 Launch** — cover illustration only, ≤800ms, then A1. No persisted session (D9): every cold launch lands here.
- **A1 Sign in** — Headline l "Your month in money.", SubCopy, card with 3 buttons: Sign in with Apple (Apple's black style — the one allowed exception), Continue with Google (secondary white button, 2px ink-100 stroke, Google G), Continue with email. Apple/Google → A4 first sign-in, → Home returning. *Variants:* **loading** (tapped button swaps its glyph for a 20pt spinner, others 40%, no timeout) · **error** (inline coral-100 error block under the buttons inside the card; clears on next tap).
- **A2 Enter email** — back chevron 44×44, Headline m, card with TextField (52pt pill, ink-100 fill, 2px stroke, Nunito 700 16, email keyboard). "Send me a link" disabled until RFC-valid. *Variant:* **invalid** — coral 2px border + coral helper, shown only after blur or Send attempt.
- **A3 Check your inbox** — envelope illustration 240pt, "We sent a link to {email}. It works for 15 minutes.", "Resend" disabled 30s with live m:ss countdown, "Use a different email" → A2 prefilled. *Variant:* **resend ready**.
- **A4 Name and avatar** — Chip "Step 1 of 4", Headline "What should we call you?", TextField, AvatarGrid 3×2 at 68pt (selected = 2px white ring on violet halo), Next. Name 1–30 chars letters/spaces/hyphens/apostrophes; both required. *Variant:* **empty** (Apple/Google shared no name).
- **A5 Currency** — Chip "Step 2 of 4", search field, full ISO 4217 list as `ListRow trailing="check"` rows with the violet check (the `SelectRow` this once named was not built — PRD D50); device-locale currency preselected so Next is always enabled.
- **A6 Salary and payday** — Chip "Step 3 of 4", Headline "Your monthly salary", AmountField (fixed prefix, Fredoka 48, violet caret), DayPicker 1–31 horizontal 44pt pills default 25, Next, "Skip for now". Amount > 0 required, or skip both.
- **A7 Saving target** — Chip "Step 4 of 4", card: "Share of salary to keep", 80px percentage, Slider 0–100 step 5 default 20 (28pt ink knob, 3px white ring, 44pt hit row), 0%/100% labels, live amount preview (target% × salary; "Add a salary to see the amount." if salary skipped), Done, "Skip for now".
- **A8 Home, first run** — month chip + Settings chip, "Hi, Maya" + 56pt avatar, hero card (StatHero "Spent so far · salary lands in 21 days", £0; chips Money in £0, Saved £0), target card ("£0 of £480", green ProgressBar, "Nothing spent yet. Streak day 4."), Log a spend, See last month (disabled), dashed "Your first recap arrives on 1 October." *Variants:* **skipped** (both A6 and A7 skipped: hero card → PromptCard "Set your monthly salary", target card → PromptCard "Set a saving target"; both keep 191/136pt heights; whole card taps to the matching step) · **after payday** (Money in £2,400, Saved = salary − spent, "Keep spending under £61 a day and you'll make it.").
- **Flow · Launch to Home** — node map with labelled edges.

### B · Home and the daily loop

- **B1 Home** — the canonical filled Home. ChipRow: month chip (tappable) … Settings chip (settings glyph, tappable). Greeting + Avatar. Hero Card: StatHero eyebrow "Spent so far" value = **Spent** (the hero number is Spent, not Money In — see GAPS G-01), StatChipRow Money in / Saved. Target Card: "Saving target" + "{Saved} of {target}", ProgressBar value Saved max target, HelperText daily allowance sentence. Buttons Log a spend (plus icon) / See last month (secondary, disabled when no previous month). "Past recaps" section. Tapping hero → B6. Live `today` tweak 21–30. *Variants:* **before payday** (20 Sep: Money In 0, Saved −£900 shown with minus, bar clamps 0, eyebrow countdown, helper "Nothing to keep yet. Your salary lands on the 25th.") · **no salary set** (both PromptCards, HelperText copy differs slightly from A8/skipped) · **target warning** (27 Oct: coral Banner "You've passed your saving line. £170 over." above target card, dismiss × session-only; target card "£310 of £480") · **with past recaps** (1 Oct: bar full, £61/day, DataRowList of ListRows "September · Squirrel", See last month enabled).
- **B2 Log a spend, empty** — Sheet over scrim: title, AmountField, eight Category Chips each on its own tint with glyph (44pt), Date row (today), Note (80 chars), Save disabled until amount > 0 and Category. *Variant:* **write failed** — inline error, sheet stays open, values kept, Save stays enabled.
- **B3 Log a spend, filled** — £12.40 + Snacks selected (chip flips to ink inverse, white glyph), numeric keyboard placement, Save enabled.
- **B4 Log a spend, date picker** — wheel; last 60 days incl. today; future at 40% and not selectable.
- **B5 Push permission** — iOS alert after the **first ever** Spend saved; fires once. Either answer → B1 with numbers animated.
- **B6 September spends** — Sheet full: one DataRow per Category with spends (IconWell in Category tone), biggest first, sub-line "{spent} so far, n spends"; tap row → B6b. *Variant:* **empty** — centred prompt + plus button → B2; streak note.
- **B6b Category detail** — one Category, newest first, date as secondary line, share of month; tap → B7. Back → B6.
- **B7 Spend detail** — same fields as B2 prefilled (10 Sep Clothes £180 jacket), Save, destructive ListRow "Delete".
- **B8 Delete confirm** — iOS action sheet. Deleting a past-month spend recomputes that Recap live.
- **Flow · Home and the daily loop.**

### C · Recap and share

- **C1 Recap 1, cover** (violet) — `RecapCoverScreen`: chips "September 2026" / "Maya's recap", Headline l "Your September in money", cover illustration hanging into card, StatHero "Money that came your way" £2,400, chips Saved £640 / Spent £1,760, PageDots 4/1. *Variant:* **no salary** — Money In £0, chips dropped, accent prompt "Set your monthly salary to get a recap · Opens Settings" inside the card. No Archetype exists in this case.
- **C2 Recap 2, kept and streak** (yellow) — `RecapEarningScreen`: chips Kept / September, Headline compact "You kept £640 of £2,400. That's 133% of your target.", jar illustration, Card "Kept against your target" with CompareBars scale=max (Kept £640 full width, Target £480 at 75%), accent Card StreakBadge 9 + "Longest no-spend streak" + "12 to 20 September. New badge unlocked." + StreakDots 30, PageDots onLight. *Variant:* **no target set** — headline drops second sentence; compare card → prompt card "Saving target / Set a target to see how you did / Pick a share of your salary to keep each month, and next month's recap will score it."; streak card unchanged.
- **C3 Recap 3, target and spending** (coral) — `RecapGoalScreen`: chips Target / Spending, Headline "133% of your saving target", SubCopy, Card "Saving target £640 of £480" with ProgressBar + piggy marker + "Target met. £160 to spare.", Card "Where £1,760 went / Mostly Bills" with **max 4** CategoryBars (Bills 35, Rides 26, Snacks 14, Clothes 10) + roll-up line "Gifts, Games, Toys, Other · 15%", accent Card "The splurge — £420 on Flight home. Your biggest single buy.", PageDots. *Variants:* **target missed** (Oct: 65%, coral part-full bar, "£170 short of your target.", splurge £560 coat 26.8%) · **no spends** (500%, bar clamps full, "£1,920 to spare", split card shows "Nothing spent this month.", splurge hidden).
- **C4 Recap 4, archetype** (green) — `RecapPersonalityScreen`: "This month, Maya, you were" + Archetype name + one-liner, mascot illustration (maxHeight capped), CompareBars Saved vs Spent scale=total, Button "Share your Squirrel", ArchetypeRow with 4 pills (current selected). **No PageDots.** *Variants:* **Treat-Yourself** (cat), **Hustler** (bee, drawn at 290pt), **Slow Burner** (tortoise).
- **C5 Share sheet** — iOS UIActivityViewController over C4, thumbnail = C6.
- **C6 Share image** — 1080×1920 poster: Archetype, month, Display Name, Saved and Spent only. **Never** Salary/Money In, kept %, target, streak, splurge, category split, avatar, email.
- **C7 Past recaps** — violet, ListRows newest first "Month · Archetype"; current month never listed.
- **Flow · Recap and share.**

### D · Settings and account (white `Sheet variant="full"` on violet)

- **D1 Settings** — back chevron + Headline "Settings". Sections (Eyebrow + DataRowList of ListRows): Profile (Maya · Otter); Money: "Salary and payday · £2,400 · 25th", "Saving target · 20%"; Currency · GBP; Notifications · "2 on"; Export spends; Delete account (destructive, no chevron); Log out (no chevron). *Variant:* **no salary** — both Money rows "Not set" in secondary; Notifications still "2 on".
- **D2 Profile** — A4's name field + AvatarGrid, Save disabled until valid. Past recaps keep their old name.
- **D3 Salary and payday** — A6's AmountField + DayPicker in tinted cards, Save disabled until changed and > 0. Append-only setting rows, effective current month forward.
- **D4 Saving target** — A7's Slider + preview (target% × salary). Forward-only. Percentage untouched by currency change.
- **D5 Currency** — A5's list; tapping a different row → D6, never converts directly.
- **D6 Currency confirm** — Sheet with scrim: "Convert everything to IDR?", body with rate, ListRows Salary and September spends before → after, Convert (ink) + Cancel (secondary). IDR shows 0 decimals.
- **D7 Currency, manual rate** — pair unsupported: TextField for rate; Convert disabled until > 0; preview recomputes; rate recorded like a fetched one.
- **D8 Notifications** — three Toggles writing immediately: Recap ready (on), Daily reminder (off, nested time row appears when on), Target warning (on) with helper "Once a month, the first time you pass your saving line." *Variants:* **permission denied** (all toggles disabled 40%, tap deep-links to iOS Settings, banner explaining) · **reminder on** (time wheel 20:00; reminder skipped on days a spend was logged).
- **D9 Export spends** — one Button "Send it"; CSV built server-side, emailed. *Variant:* **sent** — ink toast pinned above bottom inset 4s, button 40% for 60s.
- **D10 Delete account** — coral-100 tinted card, explanation, coral Button "Delete my account" → D11.
- **D11 Delete account, confirm** — TextField must equal "DELETE" exactly to enable the coral button. Sign-out immediate; data gone within 24h.
- **D12 Log out** — iOS action sheet; lands on A1.
- **Flow · Settings and account.**

## Interactions and behaviour (system-wide)

- Navigation chrome: **no nav bar, no tab bar, no header titles.** Chip row (Home) or a 44×44 back chevron + Headline (sub-screens) is the whole chrome.
- Header `Chip` and `ArchetypePill` are drawn at 30/24pt; when tappable they get an invisible 44pt hit area (`tappable`).
- Disabled controls: 40% opacity, never hidden. Buttons enable/disable live on every keystroke; **errors show only after blur or submit**, never while typing.
- Validation errors are inline (coral-100 block or coral field border + coral helper), pushing content down rather than overlaying.
- Every write requires the network; a failed write keeps the sheet open with values intact and Save enabled for retry.
- Empty-state cards keep the height of the filled card they replace (191 hero, 136 target) so layout never jumps.
- Prompt cards (empty states) are whole-card tap targets with a chevron well and open the matching setup screen.
- Session: tokens in memory only; force-quit or reboot → A1. Background → resumes.
- Number changes on Home animate in (`motion.reveal.number`) after a save.

## Domain rules (consolidated from PRD §8 and the canvas notes — they agree unless flagged in GAPS.md)

All computations are for one calendar month M in device-local dates. Amounts are integers in the Currency's minor units (GBP 2, IDR 0, BHD 3).

- **Effective payday** = min(Payday, days in M).
- **Money In** = Salary in force for M if effective payday ≤ today (or M is past), else 0. No Salary → 0.
- **Spent** = Σ non-deleted Spends dated in M. **Saved** = Money In − Spent (may be negative; displayed "−£900"; bars clamp at 0).
- **Target amount** = round-half-up(Salary × percent ÷ 100). **kept%** = round(Saved ÷ Target × 100) clamped 0–999.
- **Splurge** = largest single Spend; ties → earliest date. Hidden when Spent = 0.
- **Streak** = longest run of consecutive spend-free days in M, 1st→last day (past) or 1st→today (current). Never crosses months. Month with no spends → elapsed days.
- **Archetype** (needs Money In > 0 AND a Saving Target; first match): Saved ≥ Target & Streak ≥ 7 → Squirrel; Saved ≥ Target & Streak < 7 → Hustler; Saved < Target & Splurge ≥ 25% of Spent → Treat-Yourself; else Slow Burner.
- **Daily allowance** (Home helper, current month, on track) = floor((Salary − Target − Spent) ÷ days remaining incl. today), min 0 → "Keep spending under £{n} a day and you'll make it." Over the line → "You're {over} past your line. Next month is a fresh start." Past month met → "Target met. {spare} to spare." Missed → "{short} short of your target." Before payday → "Nothing to keep yet. Your salary lands on the {payday}." No spends yet → "Nothing spent yet. Streak day {n}."
- **Target Warning line** = Salary − Target amount. Fires once per month (MonthFlag) the first time Spent exceeds it after a save; push if enabled, Home Banner regardless; dismiss is session-only; edits/deletes never un-fire. Cannot fire without Salary and Target.
- **Settings are append-only** (SalarySetting, SavingTargetSetting with effective_from year-month). Edits apply current month forward; past months keep old values.
- **Spend edit rules:** amount > 0, Category required, date within last 60 days incl. today, note ≤ 80 chars. Backdated edits recompute past Recaps live (no stored Recap).
- **Currency change:** rate fetched (or typed if unsupported) → confirm → one transaction rewrites every Salary row and Spend amount at round-half-up(amount × rate), records a CurrencyConversion row. Percentages untouched. Not reversible.
- **Past recaps list:** newest first, months with no Spend and no Money In hidden, current month never listed; a month's Recap opens on the 1st of the next month.
- **Notifications:** Recap ready 09:00 on the 1st if previous month has data; Daily reminder at chosen time only if no spend logged today; Target Warning per rule above. Permission asked once after first spend; denied → toggles disabled and deep-link to iOS Settings.
- **Share image:** Archetype, month, Display Name, Saved, Spent only.

## State the app needs

Member (display_name, avatar_key, currency_code) · SalarySetting[] · SavingTargetSetting[] · Spend[] · MonthFlag (target_warning_sent_at) · NotificationPreference · session (in-memory tokens) · UI-only: `today`, selected month, Banner dismissed-this-session, resend countdown, export cooldown. Recap is a pure function of the above.

## Assets

Illustrations and avatars are data URIs on `MMR.ILLUSTRATIONS` in the design system; the canvases additionally use `uploads/ill-avatar-otter.webp` for the Home avatar (same otter as the system's `avatar-otter`). Apple and Google marks are the providers' own. Category glyphs, settings/chevron/plus/share/sparkle/piggy come from `Icon`.

## Files in this package

- `README.md` — this spec.
- `GAPS.md` — every undefined, undrawn or contradictory state, framed as decisions, cross-referenced to PRD sections and screen ids.
- `PRD.md` — the product requirements document, Draft v1 (2026-09-04).
- `DESIGN_NOTES_VERBATIM.md` — every yellow annotation from the four canvases, keyed by screen id.
- `design-files/*.dc.html` — the four canvases (source of truth for copy, layout and the reference formulas in each file's logic class).
- Screen renders are **not** in git. The 65 PNGs were 5.3 MB of binaries that nothing linked to individually and that git would keep forever; the screens are live in the design project (link above) and re-exportable from the canvas. They remain in this folder on the machine that produced the handoff, ignored via `docs/design/**/screens/`.
