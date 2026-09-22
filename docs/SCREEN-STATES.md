# Screen states — the design contract

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
| A1 | session expired | After 30 days of no use the app erases its saved sign-in (D62, the 30-day rule of IOS I2). Same as A1 plus one line above the buttons: `It's been a while. Sign in again to pick up where you left off.` |
| A1 | sign-in refused | The realm refused to renew a saved sign-in before those 30 days, for example after its sessions were ended. Same as A1 plus `InlineError` inside the card: `Your session expired. Sign in again.` |
| A2 | Enter email | `TextField`, Send disabled until valid. |
| A2 | invalid | Coral border and helper, only after blur or Send. |
| A3 | Check your inbox | Envelope 240pt, 15-minute line, Resend countdown. |
| A3 | resend ready | Countdown becomes `Resend`. |
| A3 | link expired | Returns to A2 with `InlineError`: `That link has expired. Links work for 15 minutes.` |
| A3 | link already used | `InlineError`: `That link has already been used. We can send another.` |
| A3 | rate limited | `InlineError`: `That's five links in an hour. Try again later.` Send disabled. |
| A3 | web fallback page | Not an app screen. A single web page: `Open My Money Recap on your phone to finish signing in.` Violet ground, envelope illustration, no controls. |
| A4 | Name and avatar | `TextField` + `AvatarGrid`, Next disabled until both valid. |
| A4 | empty | Apple/Google shared no name. |
| A5 | Currency | Searchable ISO list as `ListRow trailing="check"`, device locale preselected. |
| A5 | empty search | `No currency matches "{query}".` plus a `Clear search` link. |
| A6 | Salary and payday | D51. `DayPicker` with a `Last day` option after 31 and a computed helper below it: `Next paydays: 25 Sep · 25 Oct · 25 Nov`, updating on every tap. No other helper line. |
| A6 | amount too long | D57. At 11 digits the field stops accepting keystrokes and shows `That's more than the app can store.` |
| A7 | Saving target | Slider, live amount preview. |
| A7 | salary skipped | D44. The slider card is replaced by a `PromptCard`: `Set a salary first` / `A target is a share of your salary, so we need the salary first.` Tapping returns to A6. The primary button reads `Skip for now`. A percentage can never be set without a salary. |
| A8 | Home, first run | D43. `Streak day 1` for a member who signed up today: the streak window starts at sign-up, not the 1st. |
| A8 | skipped | D54. The hero card keeps the `StatHero` showing Spent; the salary `PromptCard` sits below it, then the target `PromptCard`. Home scrolls in this state (D69). |
| A8 | after payday | Money in fills, Saved chip appears, daily line shows. |
| — | Flow · Launch to Home | Includes the A0b unlock branch and the A7-blocked edge. |

---

## B · Home and the daily loop

| ID | State | What it must show |
|---|---|---|
| B1 | Home | Spent hero, Money in and Saved chips, target card, two buttons, recap list. |
| B1 | before payday | D41. Before Money In > 0 the chip row carries only `Money in £0` — no Saved chip. Spent hero, countdown eyebrow, bar clamped at zero, `Nothing to keep yet. Your salary lands on the 25th.` |
| B1 | no salary set | D54. The hero keeps the `StatHero` with Spent; the salary `PromptCard` sits below it, then the target `PromptCard`. Home scrolls in this state (D69). |
| B1 | no target set, salary set | The hero card is entirely normal — Spent, `Money in`, `Saved`. Only the target card becomes a 136pt `PromptCard` `Set a saving target`. No daily-allowance sentence, and the Target Warning cannot fire. |
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
| B4 | Date picker | Last 60 days, future at 40%. |
| B5 | Push permission | iOS alert, once, after the first spend. |
| B6 | Spend list | D59. Title `Spends` with a **month switcher** under it: current month by default; tapping lists every month with data, newest first, as `ListRow trailing="check"`. |
| B6 | empty | D43. Centred `Nothing spent yet. Streak day {n}.` and a plus button; the streak counts from sign-up (23 on 26 September for a 4 September sign-up). |
| B6b | Category detail | One category, newest first, share of month. |
| B7 | Spend detail | Prefilled fields, destructive Delete row. |
| B7 | spend older than 60 days | §6.10. The date row becomes read-only text with `Dates older than 60 days can't be changed.` Amount, Category and Note stay editable. |
| B8 | Delete confirm | iOS action sheet. |
| — | Flow · Home and the daily loop | Includes the offline branch, the month switcher and the preview entry. |

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
| C3 | no target set | `RecapGoalScreen` with `target={null}`. Headline is `Where £1,760 went` and the `SubCopy` is dropped — there is no percentage to state. The saving-target card becomes a `PromptCard` `Set a saving target`. Bars and Splurge unchanged. |
| C4 | Recap 4, archetype | D46 close control. Still no `PageDots` — this is the screenshotted screen. |
| C4 | Treat-Yourself / Hustler / Slow Burner | `RecapPersonalityScreen` with the `archetype` prop; the mascot's `maxHeight` keeps the Hustler clear of the pill row. |
| C4 | no archetype, no target | §6.14. `PromptCard` in place of the mascot: `Set a saving target to get an archetype`. Share button hidden. The four pills sit at 40%. |
| C5 | Share sheet | iOS `UIActivityViewController`, placement only. |
| C6 | Share image, Squirrel | 1080×1920, Archetype/month/name/Saved/Spent only. |
| C6 | Share image, Hustler / Treat-Yourself / Slow Burner | §6.15. Three more, one per Archetype. The four mascots differ in aspect ratio and the Hustler is the tallest, so each composition needs proving at 1080×1920. |
| C7 | Past recaps | `ListRow`s newest first, current month never listed. |
| C7 | empty | No months with data yet: a centred line `Your first recap arrives on 1 October.` and a link to the preview. |
| — | Recap at accessibility text size | D66. One demonstration artboard: the four screens as a single vertical scroll, `PageDots` as section markers. Not four artboards — one, showing the rule. |
| — | Flow · Recap and share | Includes the close control, the partial-recap branch and the preview entry. |

---

## P · Month so far and Partial Recap (new group)

Four artboards, three situations (D52, D64). The layout is identical in all three; only the screen-1 headline and the screen-4 line change. Read-only. Close control and tappable dots per D46.

| ID | What it must show |
|---|---|
| P1 | Cover. `StatHero` `What you've spent` = Spent. Chip `{n} spends`. **No Money In figure anywhere.** Headline by situation: `{Month} so far` · `Your first {n} days` · `Your {Month} spending`. |
| P2 | Streak. `StreakBadge`, `StreakDots`, the streak sentence. No `CompareBars` — there is no Money In to compare against. |
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
| D2 | Profile | Name field + `AvatarGrid`. Note that past recaps show the **current** name — nothing about a Recap is stored. |
| D3 | Salary and payday | D51 `DayPicker` with `Last day` and the three-date helper. D56: a line under the payday card reading `Your new payday starts in {next month}.` |
| D3 | which salary is this | D56. Sheet shown only when the current month's payday has passed. Two `ListRow`s: `The £2,800 I was paid on 25 September` and `My new salary, from 25 October`. Cancel returns without saving. |
| D4 | Saving target | D65. Slider and preview, then the forward-only line `Your new target starts in {next month}.` No sheet. |
| D5 | Currency | `ListRow trailing="check"` list, tapping opens D6. |
| D6 | Currency confirm | Rate, before/after rows, Convert + Cancel. |
| D6 | rate fetch failed | `InlineError`: `We couldn't get today's rate.` with `Try again` and `Enter it myself` (→ D7). |
| D6 | nothing to convert | No salary and no spends: the confirm sheet is skipped entirely and the currency changes directly, with a `Sent ✓`-style confirmation on the row. |
| D7 | Manual rate | Rate field, Convert disabled until > 0, live preview. |
| D8 | Notifications | Three toggles writing immediately. |
| D8 | permission denied | All three at 40%, deep-link to iOS Settings. |
| D8 | reminder on | Nested time row, 20:00. |
| D9 | Export spends | One `Send it` button. |
| D9 | sent | D67. The button itself becomes `Sent ✓` and dims for 60 seconds. |
| D9 | no deliverable address | `InlineError`: `We can't reach {email}.` with a link to Profile. |
| D10 | Delete account | Coral-100 card, explanation, coral button. |
| D11 | Delete confirm | Typed `DELETE` enables the button. |
| D11b | Restore account | D60. Shown after sign-in when deletion was requested under 24 hours ago. `Sheet variant="full"`, coral tint, headline `Welcome back`, body naming the deletion time and what is lost. `Keep my account` (ink) / `Continue with deletion` (secondary). |
| D12 | Log out | iOS action sheet. |
| — | Flow · Settings and account | Includes the which-salary sheet, the restore branch and the offline states. |

---

## Cross-cutting

| ID | What it must show |
|---|---|
| X1 | **iPad portrait.** D40 as amended by D68: the single column, capped at 600pt, centred on a full-bleed ground. One artboard, Home. |
| X2 | **iPad landscape.** Same column, same centring, `Sheet`s anchored to the bottom of the column rather than the screen. One artboard, Log a spend over Home. |
| X3 | **IDR proof · Home.** D45. B1 with Salary Rp 15.000.000, Spent Rp 9.400.000. Proves that `StatHero` shrinks to fit (80 → 56 → 40pt) and that separators follow the device language (D71). |
| X4 | **IDR proof · Recap cover.** C1 with the same ledger. |
| X5 | **IDR proof · Share image.** C6 with the same ledger, at 1080×1920. |

---

## Copy that must be identical everywhere

One string, one place. Where two screens show the same idea, they show the same words.

| Situation | The string |
|---|---|
| Before payday, target helper | `Nothing to keep yet. Your salary lands on the {payday}.` |
| Nothing spent yet | `Nothing spent yet. Streak day {n}.` |
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
| Amount cap | `That's more than the app can store.` |
| Payday preview | `Next paydays: {d1} · {d2} · {d3}` |
| No salary prompt | `Set your monthly salary` |
| No salary prompt, helper | `We'll show what you kept once we know what came in.` |
| No target prompt | `Set a saving target` |
| Target blocked in onboarding | `Set a salary first` |
