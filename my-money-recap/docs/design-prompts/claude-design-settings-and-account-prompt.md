# Prompt for Claude Design, part 4: Settings and account

Paste everything below the line into Claude Design, in the **same "My Money Recap" project**, as a **new file** (`Settings and Account.dc.html`).

Everything set during onboarding is editable here, plus notifications, export, deletion and log out. The design system gained `ListRow`, `Toggle`, `Sheet` and the `settings` / `chevron` glyphs on 2026-09-04 for this part, and the onboarding cards from part 1 (`TextField`, `AvatarGrid`, `AmountField`, `Slider`) are reused verbatim rather than redrawn.

---

Design Settings and the account flows for My Money Recap, as a multi-artboard canvas. This is part 4 of four. Use the synced "My Money Recap" design system for every element.

## Product facts you must respect

- Users are adults (18+). Money words stay plain.
- iPhone 390 × 844 artboards, portrait. Fredoka for headlines and numbers, Nunito for sentences. Three radii: cards 28, inset blocks 16 (`InlineError`, row press wash), pills 999. No shadows.
- **Settings is a white surface, not a fifth ground.** Use `Sheet variant="full"` over the violet ground — a white card filling the screen with `radius.card`. Every sub-screen follows the same pattern, so Settings reads as one white space sitting on the app's violet.
- Every row is a `ListRow`: label, optional right-aligned value, and a `chevron`. Rows are at least 44 pt tall.
- Each sub-screen **reuses the matching onboarding card** rather than inventing a new layout: Profile reuses the name field and `AvatarGrid`, Salary reuses the `AmountField` and the day picker, Saving target reuses the `Slider`, Currency reuses the searchable list.
- Destructive rows and buttons are coral. Confirmations for destructive actions are iOS action sheets or alerts, drawn for placement only.

## The sample data

**Maya · GBP · salary £2,400 · payday 25 · saving target 20% = £480 · Avatar Otter · maya@example.com.** Today is 2 October 2026, so September is a completed month and October is in progress.

For the currency-change flow, the target Currency is **IDR at a rate of 1 GBP = 20,150 IDR**, which turns the £2,400 salary into **Rp 48,360,000**. Indonesian rupiah has **no minor units**, so converted amounts show no decimals at all — that contrast with GBP's two decimals is the point of the screen.

## Screens to draw, in one row, left to right, 80 px gap, titled exactly as below

**D1 · Settings.** Reached from the Settings chip on Home. `Sheet variant="full"` on the violet ground, back chevron top-left, `Headline` size m "Settings". Then `DataRowList` sections with a small `Eyebrow` label above each group:
- *You* — "Profile" value "Maya", "Currency" value "GBP".
- *Money* — "Salary and payday" value "£2,400 · 25th", "Saving target" value "20%".
- *Alerts* — "Notifications" value "2 on".
- *Your data* — "Export spends", "Delete account" (`tone="destructive"`, no chevron), "Log out" (no chevron).

**D2 · Profile.** `Headline` size m "Profile". A white `Card` with a `TextField` labelled "First name", value "Maya". A second `Card` titled "Your avatar" holding the `AvatarGrid` of six, Otter selected with the 2 px white ring on its violet halo. Ink `Button` "Save". Annotate: "Display Name is 1 to 30 characters — letters, spaces, hyphens and apostrophes. Save is disabled until the name is valid and an avatar is chosen."

**D3 · Salary and payday.** `Headline` size m "Salary and payday". A `Card` with the `AmountField`, fixed "£" prefix, value 2,400.00. A second `Card` titled "Payday" with the horizontal 1-to-31 day picker, 25 selected as an ink pill, and `HelperText` "If the month is shorter, we use its last day." Ink `Button` "Save". Below it, a violet info `Banner`: "Changes apply to October and every month after. September's recap keeps the £2,400 it was built on." Annotate: "Setting rows are append-only. Editing writes a new row effective from the current month; past months are never recomputed."

**D4 · Saving target.** `Headline` size m "Saving target". A `Card` with "20%" large in Fredoka, the `Slider` from 0 to 100 in steps of 5, and beneath it the live preview "£480 of £2,400 each month". Ink `Button` "Save". Same forward-only `Banner` as D3.

**D5 · Currency.** `Headline` size m "Currency". A `Card` with a search `TextField` "Search by name or code", then `ListRow`s with `trailing="check"`: "GBP · British Pound" selected with the violet check, then "IDR · Indonesian Rupiah", "BHD · Bahraini Dinar", "USD · US Dollar", "EUR · Euro", "SGD · Singapore Dollar". Annotate: "The full ISO 4217 list, sorted by name, searchable by code or name. The device locale's currency is preselected at setup."

**D6 · Currency confirm.** IDR tapped. A `Sheet` rises over D5, titled "Convert everything to IDR?". Body: "All your amounts will be converted at today's rate, 1 GBP = 20,150 IDR. Your £2,400 salary becomes Rp 48,360,000. **This can't be undone.**" A small `DataRowList` previewing two conversions — "Salary  £2,400 → Rp 48,360,000" and "September spends  £1,760 → Rp 35,464,000". Ink `Button` "Convert" and `Button variant="secondary"` "Cancel" — the secondary variant is a real prop now, do not hand-style a white pill. Annotate: "The rate is fetched before the sheet is shown. Saving Target percentages are never touched — only amounts convert."

**D7 · Currency, manual rate.** The provider does not support the pair. Same sheet, but with an `AmountField`-style `TextField` labelled "Rate: 1 GBP =" awaiting a typed number, and `HelperText` "We couldn't fetch this rate. Enter it yourself to continue." "Convert" is disabled until a rate above zero is entered.

**D8 · Notifications.** `Headline` size m "Notifications". A `Card` with three rows, each a label, a `HelperText` line, and a `Toggle`:
- "Recap ready" — **on** — "09:00 on the 1st, when last month has data."
- "Daily reminder" — **off** — "Only if you haven't logged a spend that day." When on, a nested time row appears showing "20:00".
- "Target warning" — **on** — "Once a month, the first time you pass your saving line."

**D9 · Export spends.** `Headline` size m "Export spends". A `Card` with `CardTitle` "Email me a CSV of every spend" and `HelperText` "It goes to maya@example.com. Every spend you've ever logged, with its date, category, amount and note." Ink `Button` "Send it".

**D10 · Delete account.** `Headline` size m "Delete account". A `Card` with coral `Eyebrow` "This is permanent", then "Deleting removes your account, every spend you've logged, and every recap. We can't get any of it back." A coral `Button` "Delete my account".

**D11 · Delete account, confirm.** The second step. A `Card` with "Type DELETE to confirm" and a `TextField` holding "DELETE" typed. The coral `Button` "Delete my account" is enabled only on an exact match. Annotate: "Two steps, and the second cannot be a tap. Signing out is immediate; the data and the identity record are gone within 24 hours."

## Second row: states and variants, same titles with a suffix

- **D1 · Settings / no salary.** Salary and Saving target were skipped in onboarding. Their rows read "Not set" in secondary text instead of a value, and the Notifications row reads "2 on".
- **D8 · Notifications / permission denied.** Push was denied at the system prompt. All three toggles are disabled at 40% opacity, and a violet `Banner` sits above the card: "Notifications are off for this app. Turn them on in iOS Settings." with a "Open Settings" link. Annotate: "The toggles deep-link to iOS Settings rather than doing nothing."
- **D8 · Notifications / reminder on.** The Daily reminder toggle is on and the nested time row is expanded into a wheel picker, 20:00 selected.
- **D9 · Export spends / sent.** A confirmation toast over the screen: "On its way to maya@example.com."
- **D12 · Log out.** An iOS action sheet over D1: "Log out of My Money Recap? You'll need your email or Apple or Google account to get back in." Buttons "Log out" and "Cancel". Annotate: "There is no persisted session by design, so every cold launch already starts at sign-in. Log out just ends this one early."

## Third row: one flow artboard

A single wide artboard titled **"Flow · Settings and account"**, thumbnails of D1 to D12 with labelled arrows: B1 Home → D1 "Settings chip"; D1 → D2, D3, D4, D5, D8, D9, D10 "tap a row"; D5 → D6 "pick a currency"; D6 → D7 "rate unavailable"; D6 → D5 "Convert · every amount rewritten"; D10 → D11 "Delete my account"; D11 → A1 Sign in "signed out immediately"; D1 → D12 → A1 Sign in "Log out". Note on the artboard: "Salary and Saving Target edits apply to the current month and forward. A currency change rewrites every stored amount once, in a single transaction, and records what rate was used."

## Annotations to add beside each screen

Short sticky notes, not inside the artboard: what each edit changes and from when; which fields gate which buttons; and, on the currency screens, exactly what converts (salary rows and spend amounts) versus what does not (the Saving Target percentage).

## Do not

- Do not invent a fifth ground. Settings and every sub-screen are a white `Sheet variant="full"` over the violet ground.
- Do not add a bottom tab bar or a navigation bar. A back chevron and the screen headline are the whole chrome.
- Do not add settings the product does not have: no theme switch, no language picker, no biometric lock, no session length, no data-sharing controls, no linked accounts.
- Do not offer a one-tap account deletion. It is always two steps, and the second is typed.
- Do not let the currency confirm sheet appear before the rate is known.
- Do not convert the Saving Target percentage. Percentages are currency-free.
- Do not draw the onboarding screens again — reference part 1's cards and reuse them.
- Do not hand-style a white secondary button or a tinted Category chip. Both are design-system props now (`Button variant="secondary"`, `Chip tone={…} selected`), added after part 2 flagged them missing.
