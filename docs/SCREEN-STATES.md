# Screen states: the design contract

Every screen and every state the app can be in, by screen id. `PRD.md` owns the rules; this file owns the states and the copy.

**Canvases:** the ids name the artboards of the 2026-09 design canvases, which left the repo on 2026-11-13 (`git log --diff-filter=D -- docs/design` finds them). Every row says what its screen must show, so it stands without them.

**Rules that are not screen states**, but that the design agent must apply everywhere: every onboarding step after the first carries a 44pt back chevron and re-selects the saved answer on resume (PRD §6.4); the Home month chip opens the Spend list's month switcher (PRD §6.8); a Recap always shows the Member's *current* Display Name, because nothing about a Recap is stored (PRD §7), and no Avatar: no Recap screen draws one (owner, 2026-09-26).

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
| B1 | with past recaps | Under the `Past recaps` label, a `DataRowList` of months, newest first, each row opening its month's Recap: this list is C7. `See last month` is enabled only when the month before Home's is in the list, and opens it (owner, 2026-09-26). |
| B1 | partial month in the list | §8.11. A `ListRow` whose value reads `Partial` instead of an Archetype name. |
| B1 | offline | D58/D63. Grey `Banner` under the chips: `No connection. You can look, but not log.` `Log a spend` at 40%. All numbers, and a Target Warning already shown, render from cache. The grey `Banner` and the dimming go when the path returns, with no relaunch. D63's second line under the month chip, `Last updated Saturday 14:20.`, comes with slice 8. |
| B1 | first load, no cache | D49. Both cards at their filled heights (191pt, 136pt) with `ink.100` blocks in place of numbers. No spinner. |
| B1 | preview card | D47. Under the `Past recaps` label, the dashed `Your first recap arrives on 1 October.` card is tappable, with a chevron well, and opens P1. |
| B2 | Log a spend, empty | Sheet, eight category chips, Save disabled. |
| B2 | write failed | `InlineError` `We couldn't save this spend. Try again.`, sheet open, values intact, Save enabled. |
| B2 | amount too long | D57. Same rule as A6. |
| B3 | Log a spend, filled | £12.40 + Snacks, Save enabled. |
| B4 | Date picker | Last 60 days, never before the 1st of the sign-up month (D72), future at 40%. Helper `You can log up to 60 days back, never ahead.` |
| B5 | Push permission | iOS alert, once, after the first spend. |
| B6 | Spend list | D59. Title `Spends` with a **month switcher** under it: current month by default; tapping lists every month with data, newest first, as `ListRow trailing="check"`. Under the switcher the month line, in the empty states too. One `DataRow` per Category, largest first: `IconWell`, the name over `{n} spends` (`1 spend`), the total on the right. |
| B6 | empty | D43. Centred `Nothing spent yet. Streak day {n}.` and a plus button, at 40% offline (D58); the streak counts from sign-up (23 on 26 September for a 4 September sign-up). |
| B6 | empty, past month | A past month with Money In and no Spend: `Nothing spent in {Month}.`, no streak; the plus button still opens Log a spend (owner, 2026-09-24). |
| B6b | Category detail | One category, newest first, each Spend by its note or else its Category on one line, the date under it; sub-line `{amount} · {n}% of {Month} so far` (`of {Month}` for a past month). Back returns to B6. |
| B7 | Spend detail | Opened from a B6b row: the log sheet titled `Edit spend`, prefilled, `Save changes`, and the destructive row `Delete this spend`. A failed save keeps the sheet and its values with `We couldn't save this spend. Try again.`; a failed delete keeps the Spend with `We couldn't delete this spend. Try again.` After either write, Home and every month it changed are refreshed from the answers. |
| B7 | spend older than 60 days | §6.10. The date row becomes read-only text with `Dates older than 60 days can't be changed.` Amount, Category and Note stay editable. |
| B7 | offline | D58. No network path: every field stays readable; `Save changes` and `Delete this spend` at 40% until the path returns, with no relaunch. |
| B8 | Delete confirm | A centred iOS alert (owner, 2026-09-24): `Delete this spend?`, `You can't undo this.`, both buttons shown, `Cancel` and a red `Delete`. A confirmed delete returns to the list with the row gone; a Category left empty returns to B6. |
| none | Flow · Home and the daily loop | Includes the offline branch, the month switcher and the preview entry. |

---

## C · Recap and share

| ID | State | What it must show |
|---|---|---|
| C1 | Recap 1, cover | Violet ground, cover illustration. The D46 chrome: a 44pt close control top-left (`plus` rotated 45°, named `Close`), white here, with the first chip row moved 34pt right to clear it, and `PageDots`, each dot a control in its own 44pt-tall slice read as `Page {n} of 4`. Chips `{Month} {Year}` and `{Display Name}'s recap`; `Your {Month} in money`; `Money that came your way` = Money In, stepping down to fit (D45); Saved (a negative one as it is) and Spent chips; `Swipe to see where it all went.` Every Recap opens here, and held sideways every page scrolls while a swipe still turns it. |
| C1 | loading | PRD §6.19. Before the first answer, with nothing stored: the ground and the illustration only, the rest fading in when the answer arrives. With no signal it stays, and `Close` still works. A Recap opened before draws its stored copy at once. |
| C1 | no salary | §8.11. A month with no Salary is a **partial month** and uses P1–P4 (situation 3), not a variant of C1. |
| C2 | Recap 2, kept and streak | Yellow ground, ink type, the close control and dots ink. Chips `Kept` and `{Month}`, the headline `You kept {Saved} of {Money In}. That's {kept%} of your target.`, the jar, the card `Kept against your target` with two bars, Kept and Target, scaled to the larger, and the streak card: the badge (`1 day`, `{n} days`), `Longest no-spend streak`, `{start} to {end}. {badge copy}` and one dot per day of the window with the longest run lit. A one-day run prints one date; with no run the badge copy stands alone (PRD §6.12). |
| C2 | overspent | The negative Saved printed as it is: `You kept −£200 of £2,400. That's 0% of your target.` (owner, 2026-09-26). |
| C2 | no target set | The headline stops after its first sentence and the bars card becomes the target card, `Saving target` / `Set a target to see how you did` / its helper, which opens D4 over the Recap. |
| C3 | Recap 3, target and spending | Coral ground. Chips `Target` and `Spending`; `{kept%} of your saving target` with no line under it; the `Saving target` card, `{Saved} of {target}` over a coral bar with the piggy marker and §8.7's sentence under it; the spending card with the vending illustration: `Where {Spent} went`, `Mostly {category}`, three `CategoryBar`s (D71) each read `{Category}, {n} percent, {amount}`, the roll-up line and the splurge. |
| C3 | target missed | Coral part-full bar, `{short} short of your target.` |
| C3 | no spends | The spending card reads `Where it went` and `Nothing spent this month`, adds `You kept every penny that came in.` when Money In is above 0, and has no splurge. |
| C3 | no target set | Headline `Where £1,760 went` (`Where IDR 0 went` with nothing spent, owner, 2026-09-28), no line under it: there is no percentage to state. The saving-target card becomes the 136pt target card `Set a saving target`, which opens D4 over the Recap. Bars and Splurge unchanged. |
| C4 | Recap 4, archetype | Green ground, the close control, no `PageDots`: this is the screenshotted screen. Chip `Your money personality`, `This month, {Display Name}, you were`, the title (`The Squirrel`), the one-line description, the mascot capped at 300pt, a card with Saved against Spent as bars and `Share your {Archetype}`, and the four pills with the month's picked out. |
| C4 | Treat-Yourself / Hustler / Slow Burner | The same page with its own title and mascot; the cap keeps the tallest, the Hustler's, clear of the pill row. |
| C4 | no archetype, no target | §6.14. The target card `Set a saving target to get an archetype` at the hero card's height in place of the mascot, which opens D4 over the Recap; the Saved and Spent bars stay; the four pills sit at 40%; no Share. |
| C5 | Share sheet | The system `UIActivityViewController` holding C6's picture. |
| C6 | Share image, Squirrel | 1080×1920 poster on the green ground (PRD §6.15), built from Archetype, month, Display Name, Saved and Spent only: the mascot, `This month, {Display Name}, you were`, the title, `{Display Name} · {Month} {Year}`, a white card with `Saved` and `Spent`, `My Money Recap` at the foot. Its own composition, not C4 captured. |
| C6 | Share image, Hustler / Treat-Yourself / Slow Burner | §6.15. Three more, one per Archetype, each proved at 1080×1920. Only `The Hustler` fits one title line; the Slow Burner's wide mascot takes a 429px slot. |
| C7 | Past recaps | Home's list (B1 with past recaps): no screen of its own (owner, 2026-09-26). Rows newest first, current month never listed, each opening its month's Recap. |
| C7 | empty | Home's dashed preview card under `Past recaps` (B1 preview card), which opens P1. |
| none | Recap at accessibility text size | D66. One demonstration artboard: the four screens as a single vertical scroll, `PageDots` as section markers. One artboard, not four, showing the rule. |
| none | Flow · Recap and share | Includes the close control, the partial-recap branch and the preview entry. |

---

## P · Month so far and Partial Recap (new group)

Four artboards, three situations (D52, D64). The layout is identical in all three; only the screen-1 headline and the screen-4 line change. Read-only. Close control and tappable dots per D46.

| ID | What it must show |
|---|---|
| P1 | Cover, as C1's ground and chips. `StatHero` `What you've spent` = Spent. Chip `{n} spends` (`1 spend`, `No spends`). `Swipe to see where it all went.` **No Money In figure anywhere.** Headline by situation: `{Month} so far` · `Your first {n} days` · `Your {Month} spending`. |
| P2 | Streak, C2's yellow ground. Chips `Streak` and `{Month}`, `Your longest no-spend streak`, the jar, the streak card with the badge, the run's dates as its title and the badge copy under it, and the dots, which stop at today in the current month. No `CompareBars`: there is no Money In to compare against. |
| P3 | Spending, C3's coral ground and chips. `Where {Spent} went` (`Where IDR 0 went` with nothing spent, owner, 2026-09-28), then C3's spending card. While `ledger`'s days until Payday is above 0, a `Saving target` card reads `Your salary lands on the 25th.` (D55); otherwise there is no target card (owner, 2026-09-26). |
| P4 | Closing, C4's green ground with the page dots. Chip `Your money personality`; no mascot, no Archetype, no Share. The four `ArchetypePill`s at 40% under `Archetypes`. The headline is the line by situation: `Your recap arrives on 1 October.` · `Your first full recap arrives on 1 November.` · `Set your monthly salary to get a full recap.` |

**Worked example for P1's second situation:** a member signs up 27 September with payday 25 and logs £200 across four days. P1 reads `Your first 4 days` over `£200`.

---

## D · Settings and account

White `Sheet variant="full"` on violet.

| ID | State | What it must show |
|---|---|---|
| D1 | Settings | Sections, `ListRow`s, destructive Delete, Log out. The Notifications row reads `{n} on`, or `All off` when none is on. |
| D1 | no salary | Both money rows read `Not set`. |
| D1 | push denied | The Notifications row reads `Off in iOS Settings` instead of a count. |
| D1 | offline | Offline `Banner` at top; every row that writes sits at 40%. |
| D1 | no preferences | `notify` did not answer and nothing is stored: the Notifications row shows no count and opens nothing. An answer without them keeps the stored ones. |
| D2 | Profile | Name field + `AvatarGrid`. Note that past recaps show the **current** name: nothing about a Recap is stored. |
| D3 | Salary and payday | Headline `Salary and payday`. Setup's `AmountField` labelled `Amount, after tax`, filled with the Salary in force, at most the Currency's decimals and 11 digits (D57). D51 `DayPicker` with `Last day` and `Next paydays: {d1} · {d2} · {d3}` counted from today. `Save`, on once the amount is above zero and it or the Payday changed. A save returns to D1, its stored Settings replaced from the answer; a failed one keeps what was typed over `We couldn't save this. Try again.` Held sideways it scrolls. |
| D3 | first Salary | None in force and none waiting: the field starts empty on Payday 25, and Save asks nothing; the Salary counts from this month (PRD §8.10). |
| D3 | payday changed | D56. Once the picked day differs from the saved one, a line under the payday card: `Your new payday starts in {Month}.`, or with the amount changed too `Your new salary and payday start in {Month}.` No sheet: both start next month. |
| D3 | waiting salary | A later Salary is saved: a line under the amount card, `{amount} from your next payday, {date}.` (owner, 2026-09-29). Home shows nothing of it. |
| D3 | which salary is this | D56. A changed amount with the same Payday, on or after this month's effective Payday (the answer's `payday_passed_this_month`), or a save answered `failed_precondition`, opens the sheet `Which salary is this?`: `You were paid on {date}. A date alone can't tell a correction from a raise, so tell us which this is.`, then two `ListRow`s in a tinted card, `The {amount} I was paid on {date}` over `Fixes {Month}` (this month) and `My new salary, from {date}` over `{Month} stays at {amount}` (next month); the dates are this and next month's effective Payday. `Cancel` saves nothing. Held sideways it scrolls. |
| D4 | Saving target | D65. Headline `Saving target`. Setup's card: `Share of salary to keep`, the percentage, the slider from 0 to 100 in steps of 5 and `{amount} of {salary} each month`, over the last Target saved (20% with none). Then the forward-only line `Your new target starts in {Month}.`, only when ledger's `new_saving_target_from` says a save now starts next month, and `Save`. No sheet. A save returns to D1, its stored Settings replaced from the answer; a failed one keeps the slider where it was over `We couldn't save this. Try again.` Held sideways it scrolls. |
| D4 | first target | None saved since setup: the save counts from this month (PRD §8.10), so there is no line. |
| D4 | no salary | D44. Setup's `PromptCard` `Set a salary first` / `A target is a share of your salary, so we need the salary first.` replaces the slider and Save, and opens D3 in D4's place. |
| D4 | over a Recap | The target cards on C2, C3 and C4 open D4 over the Recap; Save or back returns to the same page. |
| D5 | Currency | PRD §8.9. Headline `Currency`. A5's search `Search by name or code` over its `ListRow trailing="check"` list in one card, 336pt tall and scrolling inside it, the Currency held checked. Tapping the checked row does nothing; another asks the API for a preview, which opens D6, D6 rate fetch failed or D7, or changes the Currency at once (nothing to convert). A failed direct change shows `We couldn't save this. Try again.` over the list. Offline the Settings row sits at 40%. Held sideways the sheet scrolls. |
| D6 | Currency confirm | A sheet over D5, `Convert everything to {code}?`: `All your amounts will be converted at today's rate, 1 {from} = {rate} {to}.`, then `Your {salary before} salary becomes {salary after}.` when there is a Salary, and `This can't be undone.`; `ListRow`s `Salary` (with a Salary only) and `Spends`, each `{before} → {after}` in ledger's rounding, `Spends` every Spend of every month; `Convert` and `Cancel`. The answer replaces Home and Settings, and every stored Spend list and Recap is erased, since they hold the old Currency; the sheet closes on D5 with the new code checked. Held sideways it scrolls. |
| D6 | rate fetch failed | The same sheet: `InlineError` `We couldn't get today's rate.` with `Try again` in its action slot, `All your amounts will be converted once, at the rate we use. This can't be undone.`, the main button `Enter it myself` (→ D7) and `Cancel`. No rows. |
| D6 | nothing to convert | No Salary and no Spends: no sheet; the Currency changes on the tap and the row shows `Changed ✓` beside its check until the list next opens. |
| D6, D7 | refused | Past D57: `That's more than the app can store.` in the sheet, `Convert` disabled. Any other failed Convert: `We couldn't convert your amounts. Try again.`, and `Convert` tries again under the same key. |
| D7 | Manual rate | The same sheet, opened by `Enter it myself` or directly for a pair the provider does not cover. `TextField` `Rate: 1 {from} =` with the new code after it and the helper `We couldn't fetch this rate. Enter it yourself to continue.`; at most 11 digits before the separator and 12 after. Until a rate is typed, `All your amounts will be converted at the rate you enter. This can't be undone.`; a moment after typing stops the API previews it, and D6's salary sentence and rows replace that line. `Convert` stays disabled until the rate is above zero and that preview has answered. |
| D8 | Notifications | Headline `Notifications`; three `Toggle`s (51 by 31, violet when on), each over its helper line. Each change sends the whole preferences message at once; a failed save puts the switch back over `We couldn't save this. Try again.` Never asked, turning one on shows the system alert first, and nothing is saved unless it is allowed. Offline, the switches sit at 40% and do nothing. |
| D8 | permission denied | The info `Banner`, `Notifications are off for this app. Turn them on in iOS Settings.` with `Open Settings`, the app's notification settings in iOS. All three at 40%, and tapping one opens iOS Settings too. |
| D8 | reminder on | `Remind me at` under Daily reminder, its time, 20:00 by default, printed in the device locale; tapping it opens the system wheel, and each settled time is saved. |
| D9 | Export spends | PRD §6.16. Headline `Export spends`. One card, `Email me a CSV of every spend` / `It goes to {email}. Every spend you've ever logged, with its date, category, amount and note.`, then `Send it`. Offline the Settings row sits at 40%. Held sideways it scrolls. |
| D9 | sent | D67. The button itself becomes `Sent ✓` and dims to 40% for 60 seconds, then reads `Send it` again. |
| D9 | refused | The mail host refused the address outright: `InlineError` `We can't reach {email}.` with no link, since Profile cannot change the address (owner, 2026-09-29). Bounces reported later wait for the mail provider (#87). |
| D9 | send failed | Any other failure: `InlineError` `We couldn't send it. Try again.`, and `Send it` tries again under the same key. |
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
| X3 | **IDR proof · Home.** D45. B1 with Salary Rp150.000.000, Spent Rp94.000.000. Proves that `StatHero` shrinks to fit (80 → 56 → 40pt) and that separators follow the device language (D71). |
| X4 | **IDR proof · Recap cover.** C1 with the same ledger: the Saved and Spent chips wrap to two lines, so the cover illustration narrows from 310 to 260pt to keep the page dots on an 844pt screen. A picture test (IOS I9). |
| X5 | **IDR proof · Share image.** C6 with the same ledger, at 1080×1920: Saved and Spent stack one per row at 84px, since `Rp56.000.000` is wider than its column (PRD §6.15), and the month prints in the device language (`Agustus 2026`). A picture test (IOS I9). |

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
| Spend save failed | `We couldn't save this spend. Try again.` |
| Settings save failed (D2, D3, D8) | `We couldn't save this. Try again.` |
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
| No target prompt (Home, Recap 3) | `Set a saving target` |
| No target prompt, helper | `Pick a share of your salary to keep each month.` |
| Target preview, onboarding | `{amount} of {salary} each month` |
| Target blocked in onboarding | `Set a salary first` |
| Home, list label | `Past recaps` |
| Recap cover (C1) | `{Month} {Year}` / `{Display Name}'s recap` / `Your {Month} in money` / `Money that came your way` |
| Recap cover hint (C1, P1) | `Swipe to see where it all went.` |
| Recap 2 (C2) | `Kept` / `{Month}` / `You kept {Saved} of {Money In}. That's {kept%} of your target.` (no target: `You kept {Saved} of {Money In}.`) / `Kept against your target`, rows `Kept` and `Target` |
| Streak card | `Longest no-spend streak` / `{start} to {end}. {badge copy}` |
| Streak badge unit | `1 day` / `{n} days` |
| Recap 3 (C3) | `Target` / `Spending` / `{kept%} of your saving target` (no target: `Where {Spent} went`) / `Saving target` / `{Saved} of {target}` |
| Spending card | `Where {Spent} went` / `Mostly {category}` / `{names} · {n}%` |
| Spending card, no Spend | `Where it went` / `Nothing spent this month` / `You kept every penny that came in.` (only with Money In) |
| Splurge | `The splurge` / `{amount} on {note or category}. Your biggest single buy.` |
| Recap 4 (C4) | `Your money personality` / `This month, {Display Name}, you were` / the title / `Share your {Archetype}` |
| Archetype titles (ledger's) | `The Squirrel` / `The Hustler` / `Treat-Yourself` / `The Slow Burner` |
| Recap target card, screen 2 | `Saving target` / `Set a target to see how you did` / `Pick a share of your salary to keep each month, and next month's recap will score it.` |
| Recap target card, screen 4 | `Set a saving target to get an archetype` / `Your archetype comes from what you kept against what you meant to keep.` |
| Partial Recap cover (P1) | `{Month} so far` / `Your first {n} days` / `Your {Month} spending` / `What you've spent` / `{n} spends`, `1 spend`, `No spends` |
| Partial Recap streak (P2) | `Streak` / `{Month}` / `Your longest no-spend streak` |
| Partial Recap Payday card (P3) | `Saving target` / `Your salary lands on the {payday}.` |
| Partial Recap closing (P4) | `Your money personality` / ledger's closing line / `Archetypes` |
| Closing lines (ledger's) | `Your recap arrives on {1st of next month}.` / `Your first full recap arrives on {date}.` / `Set your monthly salary to get a full recap.` |
| Share image | `This month, {Display Name}, you were` / the title / `{Display Name} · {Month} {Year}` / `Saved` / `Spent` / `My Money Recap` |
| Restore account, the request | `You asked us to delete this account {day} at {time}. Nothing has gone yet.` ({day} is `today` or `yesterday`) |
| Restore account, what is lost | `Keep it and every spend, every recap and every setting stays exactly as it was. Continue and it all goes at {time} {day}, for good.` ({day} is `today` or `tomorrow`, 24 hours after the request) |
| Restore account, keep failed | `We couldn't keep your account. Try again.` |
| Notifications row (D1) | `{n} on` / `All off` / `Off in iOS Settings` |
| Notifications (D8) | `Notifications` / `Recap ready` / `09:00 on the 1st, when last month has data.` / `Daily reminder` / `Only if you haven't logged a spend that day.` / `Remind me at` / `Target warning` / `Once a month, the first time you pass your saving line.` |
| Notifications denied (D8) | `Notifications are off for this app. Turn them on in iOS Settings.` / `Open Settings` |
