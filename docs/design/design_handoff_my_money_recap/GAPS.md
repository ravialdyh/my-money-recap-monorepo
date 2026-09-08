# GAPS — undefined, undrawn and contradictory states

Each item: **what is missing**, **evidence** (PRD section / screen id / note), and **the decision to make**. Severity: 🔴 blocks implementation, 🟠 visible to users, 🟡 polish/consistency.

Ids like `B1` refer to `data-screen-label` values in the canvases and to `screens/*.png`. "Note" = the yellow annotation, quoted in `DESIGN_NOTES_VERBATIM.md`.

---

## 1. Partial-setup states (the salary-set / target-unset family)

**G-01 🔴 Home with Salary set but no Saving Target — not drawn.**
Evidence: only "both skipped" exists (A8/skipped, B1/no salary set). PRD 6.8 says "No Saving Target set: 'Set a saving target' prompt" for the target card but does not say what the hero card, helper sentence, daily allowance, and Banner do. Home logic hardcodes `TGT = 480` (Home file logic class) with no null branch.
Decide: hero card = normal (Spent, Money in, Saved); target slot = 136pt PromptCard; no daily-allowance helper; Banner cannot fire. Confirm and draw.

**G-02 🔴 Home with Saving Target set but no Salary — not drawn, and ambiguous.**
Evidence: A7 note allows setting a % without a salary ("Add a salary to see the amount"). Target amount = Salary × % is undefined. PRD 6.8 hero prompt covers Salary only.
Decide: does the target card show "20% · Set a salary to see the amount", or the Salary prompt only (target hidden)? Does the Streak still count?

**G-03 🔴 Recap 3 (RecapGoalScreen) with no Saving Target — not drawn, not in PRD.**
Evidence: PRD 6.12 defines the no-target Recap 2; 6.13 says nothing; `RecapGoalScreen` has no `noTarget` branch (only `splurge` nullable). C2/no-target note says "next month's recap will score it."
Decide: drop the target card and headline ("Where £1,760 went" becomes the headline?), or a prompt card in the target slot? kept% is undefined → what SubCopy?

**G-04 🔴 Recap 4 with no Saving Target — not drawn, contradicts itself.**
Evidence: PRD 8.6 "Requires Money In > 0 and a Saving Target; otherwise no Archetype." C1/no-salary note: "screen 4 shows this prompt instead of a mascot" — but the drawn prompt (C1/no salary) is on screen 1, and no screen-4 prompt exists. Nothing for the *target-missing* case at all. C7 lists "Month · Archetype" — what value when there is none?
Decide: design a screen-4 "no archetype" state (target prompt vs salary prompt), decide what the Past recaps row and Home recap list show ("—", "No archetype", hidden?), and whether Share is disabled.

**G-05 🟠 Settings with Salary set, Target "Not set" — not drawn.**
Evidence: D1/no salary shows both rows "Not set". Trivial, but confirm "Not set" copy and that the D4 slider opens at 20% default with amount preview.

**G-06 🟠 Recap 1 and Recap 2 with salary set but Payday not yet reached in the *recap* month.**
Evidence: PRD 8.2 "For a past month it is always the full Salary." Fine — but for the **first** month if the member signed up *after* payday, Money In for that month is still the full salary by this rule, yet the Home said "salary lands in n days" and never showed money arriving.
Decide: is Money In for the sign-up month the full salary, prorated, or 0 if payday already passed before sign-up? Affects A8 "Streak day 4" logic too (streak since sign-up vs since 1st — see G-14).

**G-07 🟠 Target percentage of 0%.**
Evidence: Slider range 0–100 (A7, D4). Target amount 0 → kept% = Saved ÷ 0.
Decide: is 0% allowed? If yes, treat as "no target" (`targetAmount <= 0` is already the `noTarget` check in `RecapEarningScreen`) — then the Settings row shows "0%" or "Not set"? Recommend min 5%.

**G-08 🟡 Target percentage of 100%.**
Daily allowance = (Salary − Salary − Spent) ÷ days = negative → floored to 0 → "Keep spending under £0 a day". Banner fires on the first spend.
Decide: acceptable, or special copy?

---

## 2. Onboarding and auth

**G-09 🟠 Onboarding resume state.**
Evidence: PRD 5.5 "Progress is saved screen by screen, so a force-quit resumes at the next unanswered screen." No screen shows re-entering at Step 2/3/4 after a cold launch (which lands on A1 → sign in → ?). Step chips read "Step n of 4" but the back chevron behaviour between onboarding steps is not drawn.
Decide: can the member go back a step? Does the Currency step re-preselect the saved choice?

**G-10 🟠 Skip semantics on A6 are contradictory.**
Evidence: PRD 6.6 "Both required together, or skip both." A6 note: Next disabled until amount > 0; Payday defaults to 25. If a member enters salary and never touches payday, payday = 25 silently.
Decide: fine (default 25 counts as set)? Then the helper "If the month is shorter, we use its last day." from PRD 6.6 is **missing from the A6 design**.

**G-11 🟠 A5 currency search — empty result state not drawn.**

**G-12 🟠 Magic link edge cases not drawn:** expired link (>15 min), already-used link, link opened on a device without the app (PRD 6.3 web fallback page — no design), rate-limit reached (PRD 11: 5/email/hour).
Decide: in-app error screens vs. Keycloak-hosted pages. At minimum the fallback web page needs a design.

**G-13 🟠 A1 error copy.** PRD 6.1 says "Error toast"; design uses an **inline** error inside the card (A1/error note). Design wins? Update PRD.

**G-14 🟠 Streak start for a brand-new member.**
Evidence: A8 note "Streak counts no-spend days since sign-up" and "Streak day 4"; PRD 8.5 streak runs from the 1st of the month. A member who signs up on the 4th would show streak 4 by PRD, and 1 by the note.
Decide: since sign-up or since the 1st? Affects Recap 2 and Archetype for the first month.

**G-15 🟡 Session-expiry mid-use.** PRD 9: SSO idle 30 min / max 24 h. Nothing shows what happens when a backgrounded app's refresh fails (silent bounce to A1? toast?).

**G-16 🟡 Returning member whose onboarding is complete but Home has no data at all** — same as A8 first-run? Confirm A8 covers it.

---

## 3. Home and spends

**G-17 🔴 Hero number contradiction: Spent vs Money In.**
Evidence: PRD 6.8 hero `StatHero` eyebrow "Money in so far", value = Money In; chips Saved + Spent. B1 note: "Spent is the hero number… Money in and Saved sit as chips." The design deliberately flipped it.
Decide: adopt the design (Spent hero) and amend PRD 6.8, or revert. Also affects acceptance criterion 4.

**G-18 🟠 Hero card tap → B6 is design-only.** PRD 6.10 says the **Spent chip** opens the list. Decide one.

**G-19 🟠 Home target card when Saved > target all month.** B1 note asks: "The bar reads full all month because Saved starts at the whole salary on payday and only falls — flag if that reads wrong." Open design question; nobody answered.
Decide: keep ProgressBar (value Saved, max target) or switch to a "spending room" bar that fills as you spend toward the line (Salary − Target).

**G-20 🟠 Negative Saved before payday.** B1/before payday shows "Saved −£900". PRD 8.3 allows negative. But StatChip tone and the C1 cover for a month where payday never arrived (member set payday after it passed) would show negative Saved on a Recap. Not drawn.

**G-21 🟠 Month chip on Home is `tappable` (B1 note) but nothing says what it opens.** Month picker? Past recaps? Decide or make it a label.

**G-22 🟠 Log a spend — amount ≥ some max?** No upper bound, no formatting for 7+ digit amounts (IDR salaries are 7–8 digits). AmountField at Fredoka 48 will overflow at ~9 digits.
Decide: max minor-units value and truncation/scale rule.

**G-23 🟠 Log a spend on a day with no Salary (B1/no salary) — Banner and daily helper impossible; confirm Save still works (note says yes) and what Home hero shows after the first spend (PromptCard still, or Spent hero with Money in £0?).**

**G-24 🟠 Spend detail (B7) for a spend older than 60 days.** Date rule "within the last 60 days" — editing a 90-day-old spend: is the date field locked, or does saving fail?

**G-25 🟠 Spend list scope.** B6 is "September spends" (current month). Past months' spends are reachable only via… nothing. PRD 6.10 implies edit "any time" (D21).
Decide: month switcher on B6, or a spend list inside each past Recap.

**G-26 🟡 B6b sort "newest first" vs B6 "biggest first"** — intentional; confirm.

**G-27 🟡 Note field: 80 chars, but the DataRow shows note or Category name.** Long notes truncation rule (1 line ellipsis?) not specified.

**G-28 🟡 Delete spend (B8) undo?** None specified. Fine for v1; state it.

---

## 4. Recap

**G-29 🔴 Recap 3 bar cap: PRD says 5, design/system say 4.**
Evidence: PRD 6.13 "capped at the five largest"; PRD 13 and C3 note: "default 4, measured… five still overflowed by 6." PRD is internally inconsistent.
Decide: 4. Fix 6.13.

**G-30 🟠 kept% > 999 or Saved negative on Recap 2/3.** Clamp 0–999 covers the top; negative Saved gives kept% 0 and "You kept −£900 of £0." Not drawn.

**G-31 🟠 Recap for a month with Money In but zero spends — Archetype.** PRD 8.6 edge case says Saved ≥ Target is guaranteed → Squirrel (streak = whole month ≥ 7). C3/no spends exists; C4 for it does not, and the splurge line in C6 share image is fine. Confirm Squirrel and that C6 reads "Spent £0".

**G-32 🟠 Recap month with Spends but no Money In (salary set after the month ended, or payday later than the month's last logged day).** Saved negative, no Archetype (Money In = 0). Which prompt? C1/no salary says "Set your monthly salary" — wrong if a salary *is* set now.

**G-33 🟠 Recap paging.** Four horizontally paged screens (PRD 5.8). Swipe only? Tap-to-advance like Stories? Close affordance? Not drawn — C1–C4 have no close/back control; the only tappable element is Share (C4 note).
Decide: dismiss gesture (swipe down? tap outside?) and whether page dots are tappable.

**G-34 🟠 "See last month" when last month has no data but an earlier month does.** Disabled per PRD; then the only way in is Past recaps list. Confirm.

**G-35 🟠 C7 Past recaps — empty state** (first month, nothing yet) is on Home as dashed text but the C7 screen itself has no empty variant. Also: is C7 needed at all if Home lists them (B1/with past recaps)? Two lists of the same thing.

**G-36 🟠 Share image copy per Archetype** — only Squirrel drawn (C6). Confirm the four descriptions and mascots fit the 1080×1920 layout (Hustler bee is taller — C4/Hustler note).

**G-37 🟡 Share when Display Name changed.** D2 note: "past recaps keep the name they were built with" — but the Recap is computed live and has no table (PRD 7). Contradiction: either Recaps snapshot the name (a stored field) or they show the current name.

**G-38 🟡 `RecapEarningScreen` no-target branch unreachable via props** — needs `null` not `undefined` (fixed per PRD 13; C2 note still says "worth fixing"). Verify in the shipped bundle; the DC works around it by composing primitives.

**G-39 🟡 RecapPersonalityScreen mascot size** — bee needs 290pt vs 340pt default; PRD 13 says `maxHeight` cap exists. Verify.

---

## 5. Settings and account

**G-40 🟠 D3 Salary edit mid-month after payday has passed.** New row effective "current month" — Money In for the current month changes retroactively, potentially un-meeting a target and re-arming nothing (warning never un-fires). Also: changing Payday from 25 to 5 on the 20th makes Money In appear instantly. Not addressed.
Decide: effective current month (PRD D19) vs next month for Payday changes.

**G-41 🟠 D4 Target edit mid-month** — same: the Banner may now be "wrong" (fired at the old line). Accept, or re-evaluate MonthFlag?

**G-42 🟠 Currency change with no salary and no spends** — nothing to convert; D6 body copy assumes amounts. Decide: skip D6 and convert directly?

**G-43 🟠 Currency change failure mid-transaction / rate fetch failure (network).** D7 covers unsupported pair; nothing covers "fetch failed, retry".

**G-44 🟠 Notifications "2 on" in D1 when permission denied** (D8/permission denied keeps preferences). Row should probably read "Off in iOS Settings". Decide.

**G-45 🟠 Export — email delivery failure / no verified email (Apple "Hide my email" relay).** Not covered.

**G-46 🟠 Delete account — pending-deletion state.** "Signs out immediately; deletion within 24h." If the member signs in again within 24h, what happens? Reactivate, or a "Your account is being deleted" screen? Not drawn.

**G-47 🟡 Log out (D12) confirmation wording** — PRD lists Log out with no confirm; design adds an action sheet. Pick one.

**G-48 🟡 Profile avatar change** — D2 note says Home greeting updates; C6 share image never shows the avatar, fine. Confirm past-recap C1 chip "Maya's recap" uses current name (see G-37).

---

## 6. Cross-cutting

**G-49 🔴 iPad layout (D40).** No design. Single column capped at 390pt centred on the ground — Sheets and the share sheet anchor where? Landscape?

**G-50 🟠 Dynamic Type.** PRD 11 requires up to accessibility "Large". All artboards are measured to fit 844pt at default size (C3 bar cap exists *because* of this). No design for text growth; Recap screens will overflow.
Decide: scrollable recap screens at large sizes, or cap type scaling on Recap only.

**G-51 🟠 Loading / skeleton states.** PRD 11: "Home renders from cache instantly and refreshes." First launch after sign-in has no cache — no loading state drawn for Home, Recap, or Settings.

**G-52 🟠 Offline / no-connection banner.** PRD D30 "Connection required." Only the write-failed sheet (B2/write failed) exists. What does Home show with no network and no cache?

**G-53 🟠 Generic error toast component.** PRD mentions toasts (6.1, 6.16 export). Design uses inline errors and one ink toast (D9/sent). Define one toast component and where it is allowed.

**G-54 🟡 Currency formatting across the app** — designs are GBP-only except D6. Negative amounts, 0-decimal (IDR) and 3-decimal (BHD) rendering in StatHero/StatChip/DataRow are not shown. BHD "BD 1,234.567" widths untested.

**G-55 🟡 Month names / localisation.** English only; "1 October" date formats hardcoded. State v1 = en-GB only.

**G-56 🟡 Reduce Motion, VoiceOver labels** — required by PRD 11, not annotated anywhere. Add per-component a11y labels to the spec (e.g. CategoryBar "Bills, 35 percent, £616").

**G-57 🟡 Design-system README says `Screen` accepts `ground="recap"` for Settings** — Settings is a Sheet on violet; fine, but D-screens' `Sheet variant="full"` gets a `flex-grow` override in the DC (`fullSheet` prop) that isn't a documented prop.

---

## 7. PRD ↔ design contradictions (summary table)

| # | PRD says | Design does | Resolve |
|---|---|---|---|
| G-17 | Hero = Money In | Hero = Spent | pick |
| G-18 | Spent chip opens list | Hero card opens list | pick |
| G-13 | Error toast on sign-in | Inline error | design |
| G-29 | 5 bars (6.13) / 4 bars (13) | 4 bars | 4 |
| G-14 | Streak from 1st | "since sign-up" (A8) | pick |
| G-37 | Recap computed live, no table | Past recaps keep old name | pick |
| G-47 | Log out, no confirm | Action sheet | pick |
| G-10 | Payday helper copy | Missing on A6 | add |
| 6.7 | Headline "How much do you want to keep?" | Card label "Share of salary to keep", headline differs | align copy |
| 6.8 | Hero eyebrow "Money in so far" | "Spent so far" | follows G-17 |
| 6.8 | Home target helper "Nothing spent yet. Streak day {n}." | present (A8) | ok |
| 6.12 | Streak helper "{start} to {end}. {badge}" | matches | ok |

## 8. Suggested order of work

1. Decide G-17 (hero number) — it changes Home, PRD 6.8 and acceptance criterion 4.
2. Decide the partial-setup matrix (G-01…G-08) and draw the 2×2: {salary, no salary} × {target, no target} for Home, Settings, Recap 2/3/4, Past recaps row.
3. Decide Recap paging/dismiss (G-33) and no-archetype screen 4 (G-04).
4. Fill PRD 6.13 cap (G-29), streak start (G-14), name snapshot (G-37).
5. Loading/offline/iPad/Dynamic Type (G-49…G-52) before implementation starts.

---

# CLOSE-OUT — 2026-09-08

All 57 gaps are resolved. Nothing above is open. The authority is now `docs/PRD.md` (rules, with a decisions log D41–D67) and `docs/SCREEN-STATES.md` (states and copy, keyed to the screen ids used above). This file is kept as the historical record of what was found; **do not work from the "decide" lines above** — work from the resolution here.

Five resolutions were verified against the design-system source rather than decided: G-29, G-38, G-39, G-53 and G-57. In each case the code was already right and a document was stale.

## 1. Partial-setup states

| Gap | Resolution |
|---|---|
| G-01 | Hero card is entirely normal; only the target card becomes a 136pt `PromptCard`. No daily-allowance sentence; the Target Warning cannot fire. → PRD §6.8, new state `B1/no target set`. |
| G-02 | **State forbidden.** D44: a Saving Target requires a Salary. Onboarding step 4 shows a prompt back to step 3. → new state `A7/salary skipped`. |
| G-03 | Recap 3 with no target: headline becomes "Where {Spent} went", `SubCopy` dropped, target card becomes a `PromptCard`. Bars and Splurge unaffected. → PRD §6.13, new state `C3/no target set`. |
| G-04 | Recap 4 with no Archetype shows a `PromptCard` in the mascot slot — "Set a saving target to get an archetype" or "Set your monthly salary to get a recap" — with Share hidden and the four pills at 40%. List rows read "Partial". → PRD §6.14, new state `C4/no archetype, no target`. |
| G-05 | Confirmed. "Not set" is the copy; D4 opens at 20% with the amount preview. |
| G-06 | D52 / PRD §8.11. The sign-up month is **partial** when its effective Payday fell before sign-up: Money In 0, no Archetype, the P1–P4 screens. A member who signs up *before* Payday is not partial. |
| G-07 | 0% means no target. `targetAmount <= 0` is already the component's "not set" branch. PRD §8.4. |
| G-08 | 100% accepted. The daily allowance is `max(0, …)` so it reads £0, never a negative; the warning firing on the first spend is correct at that setting. PRD §8.7. |

## 2. Onboarding and auth

| Gap | Resolution |
|---|---|
| G-09 | 44pt back chevron on every step after the first; each step saves on its primary button; resume re-selects saved answers, including the Currency. PRD §6.4. |
| G-10 | Payday default 25 counts as set. The Draft v1 helper copy is **deleted** rather than added — D51 replaces it with three computed dates. → `A6` redraw. |
| G-11 | → new state `A5/empty search`. |
| G-12 | → four new states: `A3/link expired`, `A3/link already used`, `A3/rate limited`, and the web fallback page. |
| G-13 | Design wins. `InlineError` inside the card, not a toast. PRD §6.1. |
| G-14 | D43. Streak window is `max(1st of M, sign-up date)`. The drawn "Streak day 4" becomes "Streak day 1"; the drawn note was right and the number was wrong. |
| G-15 | D62. 30 days from last use, biometric unlock each launch. → new state `A1/session expired`. |
| G-16 | Confirmed: A8 covers it. |

## 3. Home and spends

| Gap | Resolution |
|---|---|
| G-17 | D42. **Spent is the hero.** The design was right; PRD §6.8 and acceptance criterion 4 were rewritten to match. |
| G-18 | Follows G-17: the hero card opens the Spend list. The "Spent chip opens the list" line is withdrawn — after D42 there is no Spent chip. PRD §6.10. |
| G-19 | Accepted as drawn. The bar measures what has been kept; starting full on payday and falling as the month is spent is the honest signal, not a bug. Before payday the question no longer arises — D41 hides Saved entirely. |
| G-20 | D41. The Saved chip is **removed** while Money In is 0, rather than showing a minus on a green pill. A Recap can no longer show a negative Saved either: those months are partial (§8.11). |
| G-21 | The month chip opens the Spend list's month switcher (D59) — the one place a month is chosen. PRD §6.8. |
| G-22 | D57. 12 digits before the separator, with a refusal message; plus the `StatHero` / `AmountField` step-down rule in §11. |
| G-23 | D54. Save still works; the hero keeps showing Spent and the salary prompt moves below it. |
| G-24 | The date row goes read-only with "Dates older than 60 days can't be changed."; amount, Category and note stay editable. → new state `B7/spend older than 60 days`. |
| G-25 | D59. A month switcher on the Spend list. |
| G-26 | Confirmed intentional: categories biggest-first, spends within a category newest-first. |
| G-27 | One line, trailing ellipsis; the note when there is one, otherwise the Category name. PRD §6.10. |
| G-28 | No undo in v1. Stated in §14. |

## 4. Recap

| Gap | Resolution |
|---|---|
| G-29 | **Four.** Verified in source: `maxCategoryBars` defaults to 4 and five overflows the artboard by 6pt. PRD §6.13's "five" was the error, and the canvas's `slice(0, 5)` is dead code that must not be ported. The drawn roll-up confirms it arithmetically — "Gifts, Games, Toys, Other · 15%" is 7+5+2+1, the tail after four bars. |
| G-30 | kept% clamps 0–999. Negative Saved can no longer reach a Recap: such months are partial (§8.11). |
| G-31 | Confirmed Squirrel; the share image reads Spent 0. |
| G-32 | §8.11. Any month with Money In 0 and at least one Spend is partial and uses P1–P4. The "Set your monthly salary" prompt is only shown when no Salary exists. |
| G-33 | D46. A 44pt close control top-left on all four screens and tappable `PageDots`. |
| G-34 | Confirmed: "See last month" stays disabled and the past-recaps list is the route. |
| G-35 | → new state `C7/empty`. C7 is kept: Home shows recent months, C7 shows all of them. |
| G-36 | → three new share images, one per remaining Archetype. The mascots differ in aspect ratio, so each needs its own 1080×1920 proof. PRD §6.15. |
| G-37 | **The note was wrong.** Recaps show the Member's *current* Display Name and Avatar. §7 says a Recap has no table and Goal 4 forbids frozen copies; a name snapshot would contradict both. |
| G-38 | **Already fixed.** Verified in source: `RecapEarningScreen.targetAmount` is `number \| null` and its null branch renders, dropping the second headline sentence automatically. The canvas note is stale; the hand-composed workaround should be dropped. |
| G-39 | **Already fixed.** Verified in source: `Illustration` has `maxHeight` and the screen passes 300. All four mascots measure 816pt. The hand-composed 290pt Hustler should be dropped. |

## 5. Settings and account

| Gap | Resolution |
|---|---|
| G-40 | D56. Editing Salary after Payday opens a "which salary is this" sheet — a date rule cannot tell a correction from a raise. **Payday changes are always forward-only**, so moving Payday backwards can never make a salary appear for a date that has passed. |
| G-41 | D65. Target edits are forward-only with **no** sheet. Unlike a Salary, a Target cannot be mistyped, so there is no correction case; raising it retroactively would turn a met target into a missed one. |
| G-42 | → new state `D6/nothing to convert`: the confirm sheet is skipped and the code changes directly. |
| G-43 | → new state `D6/rate fetch failed`, with Retry and "Enter it myself". |
| G-44 | → new state `D1/push denied`: the row reads "Off in iOS Settings", not a count. |
| G-45 | → new state `D9/no deliverable address`, with a link to Profile. |
| G-46 | D60. Signing in within 24 hours offers to restore. → new screen `D11b Restore account`. |
| G-47 | Design wins: the action sheet stays. |
| G-48 | Confirmed, and generalised by G-37: the current name and avatar are shown everywhere, including past Recaps. |

## 6. Cross-cutting

| Gap | Resolution |
|---|---|
| G-49 | → `X1` iPad portrait and `X2` iPad landscape. Sheets anchor to the bottom of the 390pt column, not the screen. |
| G-50 | D66. At accessibility text sizes the four Recap screens become one vertical scroll with `PageDots` as section markers. One demonstration artboard, not 62 redraws. |
| G-51 | → new state `B1/first load, no cache`: both cards at their filled heights with `ink.100` blocks, no spinner. |
| G-52 | D61. **Read offline, write online.** The session persists behind biometric unlock and the data is cached, so Home, the Spend list and every past Recap render with no signal. → `B1/offline`, `D1/offline`, `A0b Unlock`, and `D63`'s "Last updated" line for the iPhone/iPad case. |
| G-53 | Resolved by removing the component. D67: `Toast` is not built. `InlineError` is the single error surface and `Banner` is the single persistent-notice surface; the export confirmation is the button's own `Sent ✓` state. |
| G-54 | D45. A `StatHero` / `AmountField` step-down rule in §11, plus `X3`–`X5`: Home, the Recap cover and the share image redrawn in IDR as proof before any new work lands. |
| G-55 | v1 is en-GB copy only. Stated in §11 and §14. |
| G-56 | Per-component VoiceOver labels in §11, including the rule that the hidden Saved chip is absent from the accessibility tree rather than present-and-empty. |
| G-57 | **There is no `fullSheet` prop and no flex-grow prop** on `Sheet` or `Card` — verified in source. The canvas behaviour is an inline `flexGrow: 1`, i.e. a canvas-local hack, not library API. If it is wanted as API it must be added deliberately. |

## 7. The PRD ↔ design contradiction table, resolved

| # | PRD said | Design did | Winner |
|---|---|---|---|
| G-17 | Hero = Money In | Hero = Spent | **Design.** D42. |
| G-18 | Spent chip opens list | Hero card opens list | **Design.** The chip no longer exists. |
| G-13 | Error toast on sign-in | Inline error | **Design.** |
| G-29 | 5 bars (§6.13) / 4 bars (§13) | 4 bars | **4**, verified in source. |
| G-14 | Streak from the 1st | "since sign-up" | **Neither as drawn.** D43: `max(1st, sign-up date)`. The note was right, the drawn number was wrong. |
| G-37 | Recap computed live, no table | Past recaps keep the old name | **PRD.** Current name always. |
| G-47 | Log out, no confirm | Action sheet | **Design.** |
| G-10 | Payday helper copy | Missing on A6 | **Neither.** D51 replaces it with three computed dates. |
| §6.7 | "How much do you want to keep?" | "Share of salary to keep" | Both kept — headline and card label are different strings. |
| §6.8 | Hero eyebrow "Money in so far" | "Spent so far" | **Design**, follows G-17. |

## 8. One thing this file did not catch

`README.md` states that the logic class at the bottom of each canvas is "the reference implementation of the formulas". It is not. Verified: Streak, kept%, Archetype, Splurge and the category roll-up are **not computed anywhere** in the four canvases — they are literal props and sentences in yellow notes. Only Home's `home(d)` derives anything, and it hardcodes a 30-day month (a divide-by-zero on the 31st of a 31-day month), omits `min(Payday, days in M)` (a Payday of 31 is never paid in September), and omits the "never below 0" clamp on the daily allowance (producing "Keep spending under £−40 a day").

**Do not port numbers or formulas out of the `.dc.html` files.** PRD §8 is the only source, and every rule there now carries a worked example against one shared ledger so any number on an artboard can be checked against the rule that produced it.
