# Design handoff: My Money Recap

The design record for the whole v1 product: 111 artboards on eleven canvases (`../../SCREEN-STATES.md` counts them). Build from `../../PRD.md` (the rules) and `../../SCREEN-STATES.md` (every screen and state, with its copy); this folder is where the drawings live.

| Canvas | Prefix | File | Screens |
|---|---|---|---|
| Auth and Onboarding | A | `design-files/Auth and Onboarding.dc.html` | A0–A5, A7 and their variants, A8 after payday |
| Home and the daily loop | B | `design-files/Home and Core Features.dc.html` | B1 and two variants, B2–B5, B6b–B8 |
| Recap and Share | C | `design-files/Recap and Share.dc.html` | C5, C6, C7 |
| Settings and Account | D | `design-files/Settings and Account.dc.html` | D1, D2, D5–D12 and their variants |
| The seven 2026-09-11 canvases (71 artboards) | A–D, P, X | [`design-files/2026-09-11/`](design-files/2026-09-11/README.md) | The IDR proofs, the month-so-far recap, every other state, iPad, the four flow maps |

The four canvases above are live at <https://claude.ai/design/p/5f615747-48cd-4708-9369-becee9793f4e>; the `2026-09-11` set at <https://claude.ai/code/artifact/c2140f85-a23f-4979-a0a5-b1bdd1d2947c>. Renders are not in git.

## About the design files

The `.dc.html` files are **design references built in HTML** on top of the real `@my-money-recap/ui` React design system. They are not production code to ship. Each screen is a `<div data-screen-label="…">` wrapping a 390×844 `MMR.Screen`. Yellow note boxes under each screen carry the behavioural rules; those notes are extracted verbatim into `DESIGN_NOTES_VERBATIM.md` so you do not have to parse the HTML. Some screens are live (sliders, fields, chips respond to input; a `today` tweak on the Home canvas changes the date).

> **The logic class at the bottom of each file (`class Component extends DCLogic`) is layout glue and sample values, not a reference implementation of the formulas.** Streak, kept %, Archetype, Splurge and the Category roll-up are hand-written literal values there, and the Home canvas hardcodes a 30-day month so its September sample keeps working. **`../../PRD.md` §8 is the only source of the formulas**, and every rule there carries a worked example. The canvases are for layout, copy and composition.

The design-system bundle (`_ds/…/_ds_bundle.js`, `styles.css`) is **not** included; the HTML will not render as a local copy without it. **To see the designs, open the project itself:** <https://claude.ai/design/p/5f615747-48cd-4708-9369-becee9793f4e>, which is where they are live and current. Component and token names below refer to that published library and are the ones to use.

## Fidelity

**High-fidelity.** Colours, type, spacing, radii and copy are final and come from the design system tokens. Recreate pixel-exact using the design system's own SwiftUI/React equivalents. Exceptions drawn "for placement only" (iOS-owned UI): push-permission alert (B5), action sheets (B8, D12), share sheet (C5), numeric keyboard (B3), date wheel (B4, D8 time). Use the native controls there.

## Screen index

What each screen on the four canvases above is for. Screens drawn in `2026-09-11/` are listed there.

Each entry: purpose · composition · notable rules. Exact copy is in the HTML. Variants are listed under their parent.

### A · Auth and onboarding (violet ground throughout)

- **A0 Launch**: cover illustration only, ≤800ms, then A1 (or the A0b unlock in `2026-09-11/` when a session is saved, D61).
- **A1 Sign in**: Headline l "Your month in money.", SubCopy, card with 3 buttons: Sign in with Apple (Apple's black style, the one allowed exception), Continue with Google (secondary white button, 2px ink-100 stroke, Google G), Continue with email. Apple/Google → A4 first sign-in, → Home returning. *Variants:* **loading** (tapped button swaps its glyph for a 20pt spinner, others 40%, no timeout) · **error** (inline coral-100 error block under the buttons inside the card; clears on next tap).
- **A2 Enter email**: back chevron 44×44, Headline m, card with TextField (52pt pill, ink-100 fill, 2px stroke, Nunito 700 16, email keyboard). "Send me a link" disabled until RFC-valid. *Variant:* **invalid**: coral 2px border + coral helper, shown only after blur or Send attempt.
- **A3 Check your inbox**: envelope illustration 240pt, "We sent a link to {email}. It works for 15 minutes.", "Resend" disabled 30s with live m:ss countdown, "Use a different email" → A2 prefilled. *Variant:* **resend ready**.
- **A4 Name and avatar**: Chip "Step 1 of 4", Headline "What should we call you?", TextField, AvatarGrid 3×2 at 68pt (selected = 2px white ring on violet halo), Next. Name 1–30 chars letters/spaces/hyphens/apostrophes; both required. *Variant:* **empty** (Apple/Google shared no name).
- **A5 Currency**: Chip "Step 2 of 4", search field, full ISO 4217 list as `ListRow trailing="check"` rows with the violet check; device-locale currency preselected so Next is always enabled.
- **A6 Salary and payday**: in `2026-09-11/`.
- **A7 Saving target**: Chip "Step 4 of 4", card: "Share of salary to keep", 80px percentage, Slider 0–100 step 5 default 20 (28pt ink knob, 3px white ring, 44pt hit row), 0%/100% labels, live amount preview (target% × salary; "Add a salary to see the amount." if salary skipped), Done, "Skip for now".
- **A8 Home / after payday**: Money in £2,400, Saved = salary − spent, "Keep spending under £61 a day and you'll make it." The first-run and skipped Homes are in `2026-09-11/`.
- **Flow · Launch to Home**: in `2026-09-11/`.

### B · Home and the daily loop

- **B1 Home**: the canonical filled Home. ChipRow: month chip (tappable) … Settings chip (settings glyph, tappable). Greeting + Avatar. Hero Card: StatHero "Spent so far" = **Spent**, StatChipRow Money in / Saved. Target Card: "Saving target" + "{Saved} of {target}", ProgressBar, HelperText daily allowance sentence. Buttons Log a spend / See last month (secondary, disabled when no previous month). "Past recaps" section. Tapping hero → B6. Live `today` tweak 21–30. *Variants here:* **target warning** (27 Oct: coral Banner "You've passed your saving line. £170 over." above target card, dismiss × session-only) · **with past recaps** (1 Oct: DataRowList of ListRows "September · Squirrel", See last month enabled). The other B1 states are in `2026-09-11/`.
- **B2 Log a spend, empty**: Sheet over scrim: title, AmountField, eight Category Chips each on its own tint with glyph (44pt), Date row (today), Note (80 chars), Save disabled until amount > 0 and Category. *Variant:* **write failed**: inline error, sheet stays open, values kept, Save stays enabled.
- **B3 Log a spend, filled**: £12.40 + Snacks selected (chip flips to ink inverse, white glyph), numeric keyboard placement, Save enabled.
- **B4 Log a spend, date picker**: wheel; last 60 days incl. today; future at 40% and not selectable.
- **B5 Push permission**: iOS alert after the **first ever** Spend saved; fires once. Either answer → B1 with numbers animated.
- **B6 Spend list**: in `2026-09-11/`.
- **B6b Category detail**: one Category, newest first, date as secondary line, share of month; tap → B7. Back → B6.
- **B7 Spend detail**: same fields as B2 prefilled (10 Sep Clothes £180 jacket), Save, destructive ListRow "Delete".
- **B8 Delete confirm**: iOS action sheet. Deleting a past-month spend recomputes that Recap live.
- **Flow · Home and the daily loop**: in `2026-09-11/`.

### C · Recap and share

- **C1–C4, the four Recap screens** and their variants: in `2026-09-11/`.
- **C5 Share sheet**: iOS UIActivityViewController over C4, thumbnail = C6.
- **C6 Share image**: 1080×1920 poster: Archetype, month, Display Name, Saved and Spent only. **Never** Salary/Money In, kept %, target, streak, splurge, category split, avatar, email.
- **C7 Past recaps**: violet, ListRows newest first "Month · Archetype"; current month never listed.
- **Flow · Recap and share**: in `2026-09-11/`.

### D · Settings and account (white `Sheet variant="full"` on violet)

- **D1 Settings**: back chevron + Headline "Settings". Sections (Eyebrow + DataRowList of ListRows): Profile (Maya · Otter); Money: "Salary and payday · £2,400 · 25th", "Saving target · 20%"; Currency · GBP; Notifications · "2 on"; Export spends; Delete account (destructive, no chevron); Log out (no chevron). *Variant:* **no salary**: both Money rows "Not set" in secondary; Notifications still "2 on".
- **D2 Profile**: A4's name field + AvatarGrid, Save disabled until valid. Every recap shows the current name; nothing about a recap is stored.
- **D3 Salary and payday**, **D4 Saving target**: in `2026-09-11/`.
- **D5 Currency**: A5's list; tapping a different row → D6, never converts directly.
- **D6 Currency confirm**: Sheet with scrim: "Convert everything to IDR?", body with rate, ListRows Salary and September spends before → after, Convert (ink) + Cancel (secondary). IDR shows 0 decimals.
- **D7 Currency, manual rate**: pair unsupported: TextField for rate; Convert disabled until > 0; preview recomputes; rate recorded like a fetched one.
- **D8 Notifications**: three Toggles writing immediately: Recap ready (on), Daily reminder (off, nested time row appears when on), Target warning (on) with helper "Once a month, the first time you pass your saving line." *Variants:* **permission denied** (all toggles disabled 40%, tap deep-links to iOS Settings, banner explaining) · **reminder on** (time wheel 20:00; reminder skipped on days a spend was logged).
- **D9 Export spends**: one Button "Send it"; CSV built server-side, emailed. The sent state is in `2026-09-11/`.
- **D10 Delete account**: coral-100 tinted card, explanation, coral Button "Delete my account" → D11.
- **D11 Delete account, confirm**: TextField must equal "DELETE" exactly to enable the coral button. Sign-out immediate; data gone within 24h.
- **D12 Log out**: iOS action sheet; lands on A1.
- **Flow · Settings and account**: in `2026-09-11/`.

## Assets

Illustrations and avatars are data URIs on `MMR.ILLUSTRATIONS` in the design system; the canvases additionally use `uploads/ill-avatar-otter.webp` for the Home avatar (same otter as the system's `avatar-otter`); like `_ds/…`, `uploads/…` resolves only inside the Claude Design project. Apple and Google marks are the providers' own. Category glyphs, settings/chevron/plus/share/sparkle/piggy come from `Icon`.

## Files in this package

- `README.md`: this file.
- `DESIGN_NOTES_VERBATIM.md`: the yellow annotations of the four canvases, quoted, so nobody has to parse the HTML.
- `design-files/*.dc.html`: the four canvases; `design-files/2026-09-11/`: the seven 2026-09-11 canvases, with their build sources.
- Screen renders are not in git: they are live in the two projects linked above and re-exportable from the canvases.
