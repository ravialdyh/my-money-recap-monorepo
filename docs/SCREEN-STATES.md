# Screen states: the design contract

Every screen and every state the app can be in, keyed to the ids in `design/design_handoff_my_money_recap/` so this file can be diffed against what already exists. `PRD.md` owns the rules; this file owns the states and the copy.

**Totals:** 102 states, every one drawn at least once across the 111 artboards of the eleven canvases in `design/design_handoff_my_money_recap/` (live: <https://claude.ai/design/p/5f615747-48cd-4708-9369-becee9793f4e> and <https://claude.ai/code/artifact/c2140f85-a23f-4979-a0a5-b1bdd1d2947c>). Every row says what its screen must show, so this file can be diffed against the canvases.

**Rules that are not screen states**, but that the design agent must apply everywhere: every onboarding step after the first carries a 44pt back chevron and re-selects the saved answer on resume (PRD §6.4); the Home month chip opens the Spend list's month switcher (PRD §6.8); a Recap always shows the Member's *current* Display Name and Avatar, because nothing about a Recap is stored (PRD §7).

**Components these screens lean on:** `PromptCard` (every empty state), `DayPicker` (Payday), `InlineError` (the only error surface) and `ListRow trailing="check"` (the currency list).

---

## A · Auth and onboarding

Violet ground throughout.

| ID | State | What it must show |
|---|---|---|
| A0 | Launch | Cover illustration, ≤800ms. |
| A0b | Unlock | D61/D62. A saved session exists: cover illustration plus a Face ID prompt. Fallback copy when biometrics are unavailable: `Enter your passcode to open My Money Recap.` Works with no network. Failure three times → A1. |
| A1 | Sign in | Three buttons, Apple first. |
| A1 | loading | Tapped button swaps its glyph for a spinner; others at 40%. |
| A1 | error | `InlineError` inside the card: `We couldn't sign you in. Try again.` |
| A1 | session expired | After 30 days of no use the app erases its saved sign-in (D62, the app's own 30-day erase of IOS I2). Same as A1 plus one line above the buttons: `It's been a while. Sign in again to pick up where you left off.` |
| A1 | sign-in refused | The realm refused to renew a saved sign-in before those 30 days, for example after its sessions were ended. Same as A1 plus `InlineError` inside the card: `Your session expired. Sign in again.` |
| A2 | Enter email | `TextField`, Send disabled until valid. |
| A2 | invalid | Coral border and helper, only after blur or Send. |
| A3 | Check your inbox | Envelope 240pt, 15-minute line, Resend countdown. |
| A3 | resend ready | Countdown becomes `Resend`. |
| A3 | link expired | Returns to A2 with `InlineError`: `That link has expired. Links work for 15 minutes.` |
| A3 | link already used | `InlineError`: `That link has already been used. We can send another.` |
| A3 | rate limited | `InlineError`: `That's five links in an hour. Try again later.` Send disabled. |
| A3 | web fallback page | Not an app screen. A single web page: `Open My Money Recap on your phone to finish signing in.` Violet ground, envelope illustration, no controls. |
| A4 | Name and avatar | `Chip` label `Step 1 of 4`, no back chevron. `TextField` `First name`, which drops digits, emoji and other punctuation as they are typed and stops at 30 characters, and `AvatarGrid` under `Pick your avatar`, each Avatar named beneath it. Next disabled until both valid. |
| A4 | empty | Apple/Google shared no name. |
| A4 | prefilled | The name the provider shared, the token's `given_name`, already in `First name`; no Avatar chosen. |
| A5 | Currency | A 44pt back chevron to step 1, then `Chip` label `Step 2 of 4`. A search field `Search by name or code` over the platform's common ISO 4217 currencies as `ListRow trailing="check"`, each labelled `{code} · {name}` and sorted by its British English name. The device region's currency is preselected, so Next is usually enabled. Held sideways, the back chevron, the question and Next sit left of the card, so the list keeps its height. |
| A5 | no region currency | The device region names no currency: no row is checked and Next stays disabled until one is tapped. |
| A5 | empty search | `No currency matches "{query}".` plus a `Clear search` link. The chosen row stays chosen underneath, so Next keeps its state. |
| A4–A7 | no connection | Setup needs a connection, since every answer is saved as it is given. A step whose save cannot reach the server shows `InlineError` in its card: `No connection. We need one to save your answers.` The answers stay and Next tries again. A launch that cannot learn the next step shows the same `InlineError` in a card on the violet ground, with `Try again`. A Member who finished setup never meets it: Home opens from the stored answer (B1 offline). |
| A6 | Salary and payday | A 44pt back chevron to step 2, then `Chip` label `Step 3 of 4`. `AmountField` labelled `Amount, after tax`, empty, with the Currency's symbol fixed in front and at most the Currency's decimals; Next stays disabled until the amount is above zero. D51: `DayPicker` on 25, with a `Last day` option after 31 and a computed helper below it: `Next paydays: 25 Sep · 25 Oct · 25 Nov`, updating on every tap, today counting. No other helper line. `Skip for now` under Next records the skip. |
| A6 | amount too long | D57. At 11 digits the field stops accepting keystrokes and shows `That's more than the app can store.` |
| A7 | Saving target | A 44pt back chevron to step 3, then `Chip` label `Step 4 of 4`. A card: `Share of salary to keep`, the percentage at 80pt, the `Slider` 0 to 100 in steps of 5 on 20, `0%` and `100%` under its ends, and the live preview `£480 of £2,400 each month`, worked out on the phone for display only (ARCHITECTURE A2). `Done`, then `Skip for now` under it. |
| A7 | salary skipped | D44. The slider card is replaced by a `PromptCard`: `Set a salary first` / `A target is a share of your salary, so we need the salary first.` Tapping returns to A6. The primary button reads `Skip for now`, the only one on the screen. A percentage can never be set without a salary. |
| A8 | Home, first run | The target card's helper is §8.7's `Nothing to keep yet. Your salary lands on the 25th.` until the first Payday (spec #92). `Streak day 1` for a Member who signed up today is the empty Spend list's line (B6, D43), not Home's. |
| A8 | skipped | D54. As B1 no salary set. |
| A8 | after payday | Money in fills, Saved chip appears, daily line shows. |
| none | Flow · Launch to Home | Includes the A0b unlock branch and the A7-blocked edge. |

---

## B · Home and the daily loop

| ID | State | What it must show |
|---|---|---|
| B1 | Home | Spent hero, Money in and Saved chips, target card, two buttons, recap list. |
| B1 | before payday | D41. Before Money In > 0 the chip row carries only `Money in £0`, no Saved chip. Spent hero, countdown eyebrow, bar clamped at zero, `Nothing to keep yet. Your salary lands on the 25th.` |
| B1 | no salary set | D54. The hero keeps the `StatHero` with Spent; the salary `PromptCard` (eyebrow `Salary`, with its helper) sits below it at 191pt and opens A6, then the target `PromptCard` (with its helper) at 136pt, which opens A7. Back from a step opened here returns to Home. Home scrolls in this state (D69). |
| B1 | no target set, salary set | The hero card is entirely normal: Spent, `Money in`, `Saved`. Only the target card becomes a 136pt `PromptCard` `Set a saving target`, which opens A7. No daily-allowance sentence, and the Target Warning cannot fire. |
| B1 | target warning | Coral `Banner`, dismiss × session-only, target card moves up. |
| B1 | target warning, before payday | D53. Same `Banner`, different copy: `You've spent £2,000 of the £1,920 you meant to spend this month.` |
| B1 | with past recaps | `DataRowList` of months, See last month enabled. |
| B1 | partial month in the list | §8.11. A `ListRow` whose value reads `Partial` instead of an Archetype name. |
| B1 | offline | D58/D63. Grey `Banner` at top: `No connection. You can look, but not log.` Second line under the month chip: `Last updated Saturday 14:20.` `Log a spend` at 40%. All numbers render from cache. |
| B1 | first load, no cache | D49. Both cards at their filled heights (191pt, 136pt) with `ink.100` blocks in place of numbers. No spinner. |
| B1 | preview card | D47. The dashed `Your first recap arrives on 1 October.` card is tappable, with a chevron well, and opens P1. |
| B2 | Log a spend, empty | Sheet, eight category chips, Save disabled. |
| B2 | write failed | `InlineError`, sheet open, values intact, Save enabled. |
| B2 | amount too long | D57. Same rule as A6. |
| B3 | Log a spend, filled | £12.40 + Snacks, Save enabled. |
| B4 | Date picker | Last 60 days, never before the 1st of the sign-up month (D72), future at 40%. Helper `You can log up to 60 days back, never ahead.` |
| B5 | Push permission | iOS alert, once, after the first spend. |
| B6 | Spend list | D59. Title `Spends` with a **month switcher** under it: current month by default; tapping lists every month with data, newest first, as `ListRow trailing="check"`. Under the switcher the month line, in the empty states too. One `DataRow` per Category, largest first: `IconWell`, the name over `{n} spends` (`1 spend`), the total on the right. |
| B6 | empty | D43. Centred `Nothing spent yet. Streak day {n}.` and a plus button; the streak counts from sign-up (23 on 26 September for a 4 September sign-up). |
| B6 | empty, past month | A past month with Money In and no Spend: `Nothing spent in {Month}.`, no streak; the plus button still opens Log a spend (owner, 2026-09-24). |
| B6b | Category detail | One category, newest first, each Spend by its note or else its Category on one line, the date under it; sub-line `{amount} · {n}% of {Month} so far` (`of {Month}` for a past month). Back returns to B6. |
| B7 | Spend detail | Prefilled fields, destructive Delete row. |
| B7 | spend older than 60 days | §6.10. The date row becomes read-only text with `Dates older than 60 days can't be changed.` Amount, Category and Note stay editable. |
| B8 | Delete confirm | iOS action sheet. |
| none | Flow · Home and the daily loop | Includes the offline branch, the month switcher and the preview entry. |

---

## C · Recap and share

| ID | State | What it must show |
|---|---|---|
| C1 | Recap 1, cover | `RecapCoverScreen` with the D46 chrome: a 44pt close control top-left (`plus` rotated 45°) and tappable `PageDots`. |
| C1 | no salary | §8.11. A month with no Salary is a **partial month** and uses P1–P4 (situation 3), not a variant of C1. |
| C2 | Recap 2, kept and streak | `RecapEarningScreen` with the D46 chrome. |
| C2 | no target set | `RecapEarningScreen` with `targetAmount={null}`: the headline drops its second sentence and the compare card becomes the target prompt. |
| C3 | Recap 3, target and spending | `RecapGoalScreen` with the D46 chrome: three bars plus the roll-up line (D71). |
| C3 | target missed | Coral part-full bar, `{short} short of your target.` |
| C3 | no spends | `splurge={null}`: the split card reads `Nothing spent this month`, the splurge block is hidden. |
| C3 | no target set | `RecapGoalScreen` with `target={null}`. Headline is `Where £1,760 went` and the `SubCopy` is dropped: there is no percentage to state. The saving-target card becomes a `PromptCard` `Set a saving target`. Bars and Splurge unchanged. |
| C4 | Recap 4, archetype | D46 close control. Still no `PageDots`: this is the screenshotted screen. |
| C4 | Treat-Yourself / Hustler / Slow Burner | `RecapPersonalityScreen` with the `archetype` prop; the mascot's `maxHeight` keeps the Hustler clear of the pill row. |
| C4 | no archetype, no target | §6.14. `PromptCard` in place of the mascot: `Set a saving target to get an archetype`. Share button hidden. The four pills sit at 40%. |
| C5 | Share sheet | iOS `UIActivityViewController`, placement only. |
| C6 | Share image, Squirrel | 1080×1920, Archetype/month/name/Saved/Spent only. |
| C6 | Share image, Hustler / Treat-Yourself / Slow Burner | §6.15. Three more, one per Archetype. The four mascots differ in aspect ratio and the Hustler is the tallest, so each composition needs proving at 1080×1920. |
| C7 | Past recaps | `ListRow`s newest first, current month never listed. |
| C7 | empty | No months with data yet: a centred line `Your first recap arrives on 1 October.` and a link to the preview. |
| none | Recap at accessibility text size | D66. One demonstration artboard: the four screens as a single vertical scroll, `PageDots` as section markers. One artboard, not four, showing the rule. |
| none | Flow · Recap and share | Includes the close control, the partial-recap branch and the preview entry. |

---

## P · Month so far and Partial Recap (new group)

Four artboards, three situations (D52, D64). The layout is identical in all three; only the screen-1 headline and the screen-4 line change. Read-only. Close control and tappable dots per D46.

| ID | What it must show |
|---|---|
| P1 | Cover. `StatHero` `What you've spent` = Spent. Chip `{n} spends`. **No Money In figure anywhere.** Headline by situation: `{Month} so far` · `Your first {n} days` · `Your {Month} spending`. |
| P2 | Streak. `StreakBadge`, `StreakDots`, the streak sentence. No `CompareBars`: there is no Money In to compare against. |
| P3 | Spending. `Where {Spent} went`, three `CategoryBar`s plus roll-up, the Splurge card. Before payday the target card is replaced by `Your salary lands on the 25th.` (D55). |
| P4 | Closing. No mascot, no Archetype, no Share. The four `ArchetypePill`s at 40%. Line by situation: `Your recap arrives on 1 October.` · `Your first full recap arrives on 1 November.` · `Set your monthly salary to get a full recap.` |

**Worked example for P1's second situation:** a member signs up 27 September with payday 25 and logs £200 across four days. P1 reads `Your first 4 days` over `£200`.

---

## D · Settings and account

White `Sheet variant="full"` on violet.

| ID | State | What it must show |
|---|---|---|
| D1 | Settings | Sections, `ListRow`s, destructive Delete, Log out. |
| D1 | no salary | Both money rows read `Not set`. |
| D1 | push denied | The Notifications row reads `Off in iOS Settings` instead of a count. |
| D1 | offline | Offline `Banner` at top; every row that writes sits at 40%. |
| D2 | Profile | Name field + `AvatarGrid`. Note that past recaps show the **current** name: nothing about a Recap is stored. |
| D3 | Salary and payday | D51 `DayPicker` with `Last day` and the three-date helper. D56: a line under the payday card reading `Your new payday starts in {next month}.` |
| D3 | which salary is this | D56. Sheet shown only when the current month's payday has passed. Two `ListRow`s: `The £2,800 I was paid on 25 September` and `My new salary, from 25 October`. Cancel returns without saving. |
| D4 | Saving target | D65. Slider and preview, then the forward-only line `Your new target starts in {next month}.` No sheet. |
| D5 | Currency | `ListRow trailing="check"` list, tapping opens D6. |
| D6 | Currency confirm | Rate, before/after rows, Convert + Cancel. |
| D6 | rate fetch failed | `InlineError`: `We couldn't get today's rate.` with `Try again` and `Enter it myself` (→ D7). |
| D6 | nothing to convert | No salary and no spends: the confirm sheet is skipped entirely and the currency changes directly, with a `Sent ✓`-style confirmation on the row. |
| D7 | Manual rate | Rate field, Convert disabled until > 0, live preview. Before a rate is typed the preview reads `All your amounts will be converted at the rate you enter.` |
| D8 | Notifications | Three toggles writing immediately. |
| D8 | permission denied | All three at 40%, deep-link to iOS Settings. |
| D8 | reminder on | Nested time row, 20:00. |
| D9 | Export spends | One `Send it` button. |
| D9 | sent | D67. The button itself becomes `Sent ✓` and dims for 60 seconds. |
| D9 | no deliverable address | `InlineError`: `We can't reach {email}.` with a link to Profile. |
| D10 | Delete account | Coral-100 card, explanation, coral button. |
| D11 | Delete confirm | Typed `DELETE` enables the button. |
| D11b | Restore account | D60. Shown after sign-in when deletion was requested under 24 hours ago. `Sheet variant="full"`, coral tint, headline `Welcome back`, body naming the deletion time and what is lost. `Keep my account` (ink) / `Continue with deletion` (secondary). No back control. Built in #105; until the Settings slice builds Delete account, only a test reaches it. |
| D11b | keep failed | `InlineError` in the sheet, above the buttons, on a white well because the sheet is already coral, with both buttons live: `We couldn't keep your account. Try again.` The same line for every failure, a lost connection included. |
| D12 | Log out | iOS action sheet. |
| none | Flow · Settings and account | Includes the which-salary sheet, the restore branch and the offline states. |

---

## Cross-cutting

| ID | What it must show |
|---|---|
| X1 | **iPad portrait.** D40 as amended by D68: the single column, capped at 600pt, centred on a full-bleed ground. One artboard, Home. |
| X2 | **iPad landscape.** Same column, same centring, `Sheet`s anchored to the bottom of the column rather than the screen. One artboard, Log a spend over Home. |
| X3 | **IDR proof · Home.** D45. B1 with Salary Rp15.000.000, Spent Rp9.400.000. Proves that `StatHero` shrinks to fit (80 → 56 → 40pt) and that separators follow the device language (D71). |
| X4 | **IDR proof · Recap cover.** C1 with the same ledger. |
| X5 | **IDR proof · Share image.** C6 with the same ledger, at 1080×1920. |

---

## Copy that must be identical everywhere

One string, one place. Where two screens show the same idea, they show the same words.

| Situation | The string |
|---|---|
| Before payday, target helper | `Nothing to keep yet. Your salary lands on the {payday}.` |
| Nothing spent yet | `Nothing spent yet. Streak day {n}.` |
| Nothing spent, past month | `Nothing spent in {Month}.` |
| Spend list title | `Spends` |
| Spend list month line, current month | `{spent} so far, {n} spends` / `{spent} so far, 1 spend` / `{spent} so far, no spends` |
| Spend list month line, past month | `{spent}, {n} spends` / `{spent}, 1 spend` / `{spent}, no spends` |
| Category row sub-line | `{n} spends` / `1 spend` |
| Category detail sub-line | `{amount} · {n}% of {Month} so far` (past month: `{amount} · {n}% of {Month}`) |
| On track | `Keep spending under {amount} a day and you'll make it.` |
| Over the line, current month | `You're {over} past your line. Next month is a fresh start.` |
| Past month, met | `Target met. {spare} to spare.` |
| Past month, missed | `{short} short of your target.` |
| Warning, before payday | `You've spent {spent} of the {line} you meant to spend this month.` |
| Warning, on or after payday | `You've passed your saving line. {over} over.` |
| Streak badge, 7+ | `New badge unlocked.` |
| Streak badge, under 7 | `Beat it next month.` |
| Offline | `No connection. You can look, but not log.` |
| Stale cache | `Last updated {day} {time}.` |
| Avatar grid label | `Pick your avatar` |
| Currency search | `Search by name or code` |
| Offline, setup | `No connection. We need one to save your answers.` |
| Amount label, Salary | `Amount, after tax` |
| Amount cap | `That's more than the app can store.` |
| Payday preview | `Next paydays: {d1} · {d2} · {d3}` |
| No salary prompt, eyebrow | `Salary` |
| No salary prompt | `Set your monthly salary` |
| No salary prompt, helper | `We'll show what you kept once we know what came in.` |
| No target prompt | `Set a saving target` |
| No target prompt, helper | `Pick a share of your salary to keep each month.` |
| Target preview, onboarding | `{amount} of {salary} each month` |
| Target blocked in onboarding | `Set a salary first` |
| Restore account, the request | `You asked us to delete this account {day} at {time}. Nothing has gone yet.` ({day} is `today` or `yesterday`) |
| Restore account, what is lost | `Keep it and every spend, every recap and every setting stays exactly as it was. Continue and it all goes at {time} {day}, for good.` ({day} is `today` or `tomorrow`, 24 hours after the request) |
| Restore account, keep failed | `We couldn't keep your account. Try again.` |
