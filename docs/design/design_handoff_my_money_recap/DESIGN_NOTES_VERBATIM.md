# Design notes, verbatim from the canvases

> **Historical transcript — quoted, not maintained.** These are the yellow annotation boxes as they
> were written on the four `.dc.html` canvases, reproduced word for word so nobody has to parse the
> HTML. They are **not** the current spec and are deliberately never edited: where a note disagrees
> with `../../PRD.md` or `../../SCREEN-STATES.md`, those files are right and this one is a record of
> what was thought at the time.
>
> `ListRow` with `trailing="check"` and `selected` already is that row (PRD D50). Any note that
> describes a component the library now owns has been re-specified in PRD §13.


## Auth and Onboarding.dc.html
- [A0 · Launch] Shows for at most 800 ms, then A1. No spinner, no tap target.
- [A0 · Launch] A saved session under 30 days old unlocks with Face ID or the passcode and goes straight to Home (A0b, D61); otherwise A1.
- [A1 · Sign in] Apple button uses Apple's mandated black style, white logo, SF label — the one allowed exception to the system.
- [A1 · Sign in] Secondary button: Button with white surface, ink label, 2 px ink-100 stroke (so it reads on a white card). Google "G" is Google's own mark.
- [A1 · Sign in] Apple / Google → A4 on first sign-in, → A8 for a returning member. Email → A2.
- [A2 · Enter email] "Send me a link" is disabled (40%) until the field holds a valid email: something@domain.tld. Validates on every keystroke; try typing in the field.
- [A2 · Enter email] TextField: 52 pt pill, ink-100 fill and 2 px stroke, Nunito 700 16. Label above in secondary. Email keyboard, autocorrect off.
- [A2 · Enter email] Back chevron (44 × 44) returns to A1. No header title.
- [A3 · Check your inbox] Resend is disabled for 30 s after each send and shows a live m:ss countdown; then it reads "Resend" (see row 2). Each resend restarts the timer.
- [A3 · Check your inbox] Envelope illustration at 240 pt, same style as the cover.
- [A3 · Check your inbox] Tapping the link opens A4 on first sign-in, A8 for a returning member. "Use a different email" → A2 with the field prefilled.
- [A4 · Name and avatar] Name: 1 to 30 characters; letters, spaces, hyphens and apostrophes only. Next is disabled until the name is valid and an avatar is picked. Tap avatars to try.
- [A4 · Name and avatar] Six avatars at 68 pt: Fox, Otter, Panda, Frog, Owl, Rabbit. Selected avatar gets a 2 px white ring on a violet halo.
- [A4 · Name and avatar] AvatarGrid: 68 pt circles in a 3 × 2 grid, 88 pt hit area. Selected = 2 px white ring on a violet halo (ArchetypePill's selected treatment).
- [A5 · Currency] GBP is preselected from the device region. One row is always selected, so Next is always enabled. Search filters by name or ISO code.
- [A7 · Saving target] Slider 0 to 100 in steps of 5, default 20. Preview recalculates live: target% × salary from A6. If salary was skipped the preview reads "Add a salary to see the amount." Drag the knob to try.
- [A7 · Saving target] Slider: ProgressBar track (violet tone) plus a 28 pt ink knob with a 3 px white ring; the hit area is the full 44 pt row.
- [A7 · Saving target] Done → A8. Skip for now → A8 with the target prompt card.
- [A1 · Sign in / loading] While the Google sheet is open: the tapped button keeps its label and swaps the G for a 20 pt spinner; the other two drop to 40%. No timeout: the provider sheet owns cancel.
- [A1 · Sign in / error] Provider cancelled or failed: all three buttons re-enable and an inline error appears under them inside the card, pushing the card taller. Clears on the next tap of any button.
- [A1 · Sign in / error] Inline error: coral-100 fill, 16 radius, coral-500 dot + coral text Nunito 800 13. Inline so it can never overlap the buttons.
- [A2 · Enter email / invalid] Error shows only after the field loses focus or Send is attempted; never while typing. Coral 2 px border plus coral helper; button stays disabled.
- [A3 · Check your inbox / resend ready] After 30 s the countdown text becomes "Resend" at full opacity with an underline, matching the other link. Tapping sends again and restarts the 30 s timer.
- [A4 · Name and avatar / empty] Initial state for Apple/Google sign-ins that share no name. Field shows a placeholder, no avatar is selected, Next is disabled. Apple/Google can prefill the name when they share it.
- [A8 · Home / after payday] Payday has passed: the eyebrow drops its countdown, Money in fills to the salary, saved = salary − spent, the bar is capped at full once saved ≥ target. The daily figure is floor((Salary − Target − Spent) ÷ days remaining, today included).

## Home and Core Features.dc.html
- [B1 · Home] {{ m.dailyNote }} Days remaining counts today, so {{ m.daysLeft }} on the {{ m.dayOrd }}. The bar reads full all month because Saved starts at the whole salary on payday and only falls.
- [B1 · Home] Spent is the hero number — it is the one figure that changes every day and the one the Member can act on. Money in and Saved sit as chips; Saved = Money In − Spent, computed, never logged. Tapping the hero card opens B6. "See last month" is disabled: August has no data.
- [B1 · Home] Month and Settings chips are controls, so they take the tappable variant: the 30 pt pill keeps its size, an invisible 44 pt hit area is added. No tab bar, no nav bar, no header title.
- [B2 · Log a spend, empty] Save enables only when amount > 0 and a Category is selected. Date defaults to today; note is optional, 80 characters max.
- [B2 · Log a spend, empty] Sheet, not a ground: white surface, radius 28 on the top corners, drag handle, over the scrim. Eight Category chips, each on its own 100 tint with its glyph in the Category colour; 44 pt tall.
- [B3 · Log a spend, filled] Live: type in the amount field and tap Category chips. Save is enabled here because amount is £12.40 and Snacks is selected; clear either and it drops to 40%.
- [B3 · Log a spend, filled] The keypad is iOS's numeric keyboard, drawn for placement only. The sheet content sits above it; the selected chip is inverse ink with its glyph in white.
- [B4 · Log a spend, date picker] Date rule: within the last 60 days and never in the future. Today is selected in ink; tomorrow onward is 40% and not selectable. The note field scrolls under the wheel while it is open.
- [B5 · Push permission] System alert, drawn for placement only — iOS owns this UI. It fires once, after the first Spend is saved, and never again. Either answer returns to B1 with the new numbers already animated in.
- [B6b · Category detail] One Category, newest first; rows carry the date as a small secondary line under the note. The sub-line gives this Category's share of the month so far. Tapping a row opens B7.
- [B6b · Category detail] Follows whichever Category was tapped in B6 — Snacks by default. Back returns to B6, not Home.
- [B7 · Spend detail] Same fields as the log sheet, prefilled from the 10 Sep Clothes £180 jacket. Same validation: amount > 0, a Category, date in the last 60 days, note 80 characters. Delete is a destructive ListRow beneath the primary button.
- [B8 · Delete confirm] iOS action sheet, placement only. Editing or deleting a Spend in a past month immediately changes that month's Recap — the Recap is computed live, never stored. Confirming returns to B6 with the row gone.
- [B1 · Home / target warning] October 27. September met its target, so the warning cannot fire there: meeting the target means Spent ≤ Salary − Target, which is the warning line. Money In £2,400, Spent £2,090, Saved £310; £2,090 − £1,920 = £170 over.
- [B1 · Home / target warning] Fires once per month, the first time Spent passes Salary − Target (£1,920). Editing or deleting spends never un-fires it. Dismiss lasts for the session only — tap the × to try it; the target card moves up to close the gap.
- [B1 · Home / with past recaps] 1 October, nothing spent yet: Saved is the whole salary, so the bar is full and the daily line is (2,400 − 480 − 0) ÷ 31 = £61. The recap list is a DataRowList of ListRows, newest first; "See last month" is enabled and opens the September recap (part 3).
- [B2 · Log a spend / write failed] Every write needs the network. A failed write never closes the sheet and never loses what was typed. Save stays enabled so the same tap retries.

## Recap and Share.dc.html
- [C5 · Share sheet] System UI, drawn for placement only. Rendering the image and presenting UIActivityViewController is the app's only job here. The preview thumbnail is the generated C6 poster.
- [C6 · Share image] The share image must never show Salary or Money In. Archetype, month, Display Name, Saved and Spent only. No page dots, no share button, no chrome.
- [C6 · Share image] Never on this image: £2,400 (Salary / Money In), the 133% kept figure, the £480 target, the streak, the splurge, the category split, the avatar, the email.
- [C7 · Past recaps] Reached from Home. The current month is never listed — its recap becomes available on the 1st, when it becomes last month. Newest first; the value is the month's Archetype.

## Settings and Account.dc.html
- [D1 · Settings] Sheet variant="full" on the violet ground: a white card filling the screen with radius 28. Back chevron and headline are the whole chrome — no nav bar, no tab bar.
- [D1 · Settings] Every row is a ListRow, 44 pt minimum. Values show the current setting; "2 on" counts enabled notifications. Delete account is the destructive tone; Delete and Log out carry no chevron because they act, they don't navigate.
- [D2 · Profile] Display Name is 1 to 30 characters — letters, spaces, hyphens and apostrophes. Save is disabled until the name is valid and an avatar is chosen. Live: edit the name or tap an avatar.
- [D2 · Profile] Reuses A4's name field and AvatarGrid unchanged. Saving updates the Home greeting and every recap, past ones included — nothing about a recap is stored (PRD §7). Cards inside the white sheet keep the default white surface with a soft shadow so fields and tracks stay legible.
- [D5 · Currency] The full ISO 4217 list, sorted by name, searchable by code or name. The device locale's currency is preselected at setup. Reuses A5's searchable list; tapping a different row opens D6, never converts directly.
- [D6 · Currency confirm] The rate is fetched before the sheet is shown. Saving Target percentages are never touched — only amounts convert: every salary row and every spend amount, once, in a single transaction, with the rate recorded.
- [D6 · Currency confirm] IDR has no minor units, so converted amounts show no decimals — the contrast with GBP's two is the point. 2,400 × 20,150 = Rp 48.360.000; 1,760 × 20,150 = Rp 35.464.000, with id-ID separators.
- [D7 · Currency, manual rate] The provider does not support the pair, so the Member types the rate. Live: Convert stays disabled until a rate above zero is entered; the preview line recomputes from it. The typed rate is recorded with the conversion like a fetched one.
- [D8 · Notifications] Three toggles, no Save: each writes immediately. Live: switch Daily reminder on and the nested time row appears. Recap ready and Target warning are on by default after the system prompt is allowed.
- [D9 · Export spends] One action, no options. The CSV is built server-side and emailed to the account address; nothing is written to the device. Amounts export in the current Currency's minor units.
- [D10 · Delete account] Step one of two. Nothing happens here except moving to D11. Coral is the only colour used for destructive controls; the card sits on the coral 100 tint to match.
- [D11 · Delete account, confirm] Two steps, and the second cannot be a tap. Live: the button enables only on an exact "DELETE". Signing out is immediate; the data and the identity record are gone within 24 hours.
- [D1 · Settings / no salary] Salary and Saving target were skipped in onboarding: both rows read "Not set" in secondary text. Notifications still counts 2 on — the target warning is armed but can never fire without a target.
- [D8 · Notifications / permission denied] Push was denied at the system prompt. All three toggles sit disabled at 40%; the toggles deep-link to iOS Settings rather than doing nothing. Preferences are kept, so allowing later restores them as they were.
- [D8 · Notifications / reminder on] Daily reminder on: the nested time row expands into a wheel, 20:00 selected, in the Member's local time. The reminder is skipped on any day a spend was already logged.
- [D12 · Log out] iOS action sheet, placement only. Log out erases the saved session and the cache on the device and lands on A1 (D61).
