# Ledger fixtures

The answer key for PRD §8. Every worked example in §8 and every numeric acceptance criterion in
§12 is one JSON file here, typed by `LedgerFixture` (`money_recap/v1/ledger_fixture.proto`), so
the arithmetic that only `ledger` implements (ARCHITECTURE A2, §7.2) is checked against numbers
the PRD wrote down, not numbers the implementation produced.

Written in slice 2 (#42, #48, #49). Run by `ledger` in slice 5, which loads each file's rows, sets
its clock to `asOf` in `timezone`, applies `saveSalary` or `convert` if present, and compares what
its RPCs answer to the expected fields. Until then, `apps/backend/internal/ledger/fixtures_test.go`
decodes every file in the backend CI job.

## What a fixture is

One ledger and one "as of" date, in protobuf JSON (`lowerCamelCase` names, enums as their names):

| Field | Role |
| --- | --- |
| `doc` | The PRD paragraph every number is worked from, with the working. A failing fixture names what to reread. |
| `timezone`, `signedUpOn`, `currencyCode` | The `member_facts` replica (§7.2). `signedUpOn` bounds the Streak window (D43) and names the sign-up month (D52). |
| `salaries`, `savingTargets`, `spends` | The rows as stored. `spends` holds nothing dated after `asOf`; its `id` is a label and `createdAt`/`updatedAt` are left for `ledger` to stamp. |
| `asOf` | Today in the Member's zone. |
| `saveSalary`, `convert` | A write made on `asOf` before the expected values are read (§8.10, §8.9); `memberId` and `idempotencyKey` stay empty. |
| `months`, `recaps`, `settings`, `conversionPreview` | What `GetMonthSummary`, `GetRecap`, `GetSettings` and `PreviewConversion` answer. List only what the example proves. |
| `crossedLine` | What logging the last entry of `spends` answered (§8.8). |
| `salariesAfter`, `spendsAfter` | The rows after `convert` or `saveSalary`, originals untouched (§8.9 step 4). |

How slice 5 compares an answer to an expected message: a top-level expected field that is left
out (`recaps` in a Home-only file) asserts nothing; inside a listed message every field is
compared, so an `optional` the PRD says is absent (Saved before Payday, D41; the Archetype of a
Partial Recap, §8.6) is left out of the file and must be absent in the answer. Three things are
never compared: `Spend.id`, which `ledger` assigns (`maya-01` is a label that pairs an input row
with `splurge` or `spendsAfter`), and `createdAt` and `updatedAt`, which `ledger` stamps.

Amounts follow `Money`: a decimal string with at most 11 digits before the separator (D57) and an
ISO 4217 code; GBP carries two decimals (`"2400.00"`), IDR none (`"15000000"`), the decimals CLDR
reports for the currency (§8.4). A sentence that embeds an amount is written as the PRD prints it
(`£160`, `Rp 2.600.000`).

`streakDay` on a month is the month's Streak so far (§8.5, the glossary's Streak): the longest
spend-free run in the window up to `asOf`. With no Spends it is the elapsed days, which Home prints
as "Streak day {n}" (SCREEN-STATES B6: 23 on 26 September for a 4 September sign-up). `streak.days`
on a Recap is the same rule over the whole window. A scalar
left out of a file is its zero value to protobuf JSON, so `partial: false` and an absent `partial`
are the same assertion; only a message or `optional` field can be "absent".

## Adding one

1. Name the file after the rule it proves: `prd-<section>-<who>-<when>.json`, for example
   `prd-8.5-signup-4-september.json` or `prd-12.6-squirrel-kept-116.json`.
2. Work every expected number from the cited paragraph and say how in `doc`.
3. Add the name to `manifest` in `apps/backend/internal/ledger/fixtures_test.go`, keeping it
   alphabetical; the test fails on a file without a line and a line without a file.
4. `cd apps/backend && go test ./internal/ledger/`.

## The set

| File | Proves |
| --- | --- |
| `prd-8.3-maya-20-september` | Maya before Payday: Spent £900, Money In £0, no Saved (§8.2, §8.3); target £480 (§8.4); Streak day 9 (§8.5); the before-Payday sentence and 11 days remaining (§8.7); the line not crossed (§8.8). |
| `prd-8.3-maya-30-september` | Maya after Payday: Spent £1,760, Money In £2,400, Saved £640 (§8.3); Streak 9 still (§8.5); the £160-a-day sentence (§8.7); the D56 sheet flag set. |
| `prd-8.6-maya-september-recap` | Maya's September Recap on 1 October: Squirrel (§8.6), kept 133% (§8.4), Streak 9 with its window and dots (§8.5), Splurge £420 (§8.3), the Category shares (§6.13), "£160 to spare" (§8.7). |
| `prd-8.4-budi-idr` | Budi's IDR ledger: target Rp 3.000.000 (§8.4, D45), Saved Rp 5.600.000, Streak 6 (§8.5), the §8.7 a-day sentence in rupiah. |
| `prd-8.2-payday-31-across-months` | Payday 31 read on 29 February 2028: Money In lands on 30 September, 31 October, 28 February 2027 and 29 February 2028 (§8.2, acceptance 13); no Saving Target. |
| `prd-8.2-payday-last-across-months` | `Last day` read on 28 February 2027: 30 September, 31 October, 28 February 2027 (§8.2, D51). |
| `prd-8.5-signup-4-september-first-run` | Sign-up on 4 September, read that day: "Streak day 1", not 4 (§8.5, D43, acceptance 11); Payday 25 still ahead, so not partial (§8.11). |
| `prd-8.5-signup-4-september-recap` | The same Member's September Recap: window from the 4th, Streak 6 not 9 (§8.5, acceptance 11); kept 292, Hustler (§8.6), £920 to spare (§8.7). |
| `prd-8.7-maya-1-october` | 1 October, nothing spent: the before-Payday sentence per the rule; the paragraph's £61 a day contradicts it and is left in `doc` for the owner. |
| `prd-8.7-maya-31-october` | 31 October: days remaining 31 − 31 + 1 = 1, never 0; £1,420 a day (§8.7); Streak 26 (§8.5). |
| `prd-8.7-maya-300-over` | £300 over the line on 30 September: allowance max(0, −300) = £0, so the over-the-line sentence; the after-Payday warning copy and the line crossed (§8.7, §8.8). |
| `prd-8.8-maya-20-september-before-payday` | Passing the £1,920 line on 20 September: "You've spent £2,000 of the £1,920 you meant to spend this month." (§8.8, D53). |
| `prd-8.8-maya-27-october-after-payday` | Passing it on 27 October: "You've passed your saving line. £170 over." (§8.8, D53). |
| `prd-8.9-maya-gbp-to-idr` | GBP → IDR at 20,150: £2,400 → Rp 48.360.000, £1,760 → Rp 35.464.000 row by row, originals untouched, 20% untouched, the preview (§8.9, acceptance 8). |
| `prd-8.10-maya-27-september-this-month` | £2,400 → £2,800 on 27 September, "The £2,800 I was paid on 25 September": September's Money In £2,800, target £560 (§8.10, D56). |
| `prd-8.10-maya-27-september-next-month` | The other answer, "My new salary, from 25 October": September stays at £2,400, the October row is `nextSalary` (§8.10, D56, acceptance 14). |
| `prd-8.10-maya-20-september-no-sheet` | The edit on 20 September: `paydayPassedThisMonth` false, no sheet, so `appliesFrom` is left unspecified and the row is effective from September (§8.10, D56). |
| `prd-8.11-signup-27-september` | Sign-up on 27 September with Payday 25: September is partial, "Your first 4 days" over £200, no Saved, no Archetype, the 1 November closing line (§8.11, D52, acceptance 12). |
| `prd-12.4-idr-salary-10th` | Salary Rp 3.000.000, Payday 25, target 20% set on the 10th: Money In Rp 0, no Saved, target Rp 600.000, 15 days to Payday (acceptance 4). |
| `prd-12.4-idr-salary-25th` | The same on the 25th: Money In and Saved Rp 3.000.000, Rp 400.000 a day (acceptance 4). |
| `prd-12.5-idr-warning-fires` | Spends reaching Rp 2.400.001 on the 20th cross the Rp 2.400.000 line with the before-Payday copy (acceptance 5, §8.8). |
| `prd-12.5-idr-warning-fires-once` | A further Spend on the 22nd does not cross again; the banner still names Spent (acceptance 5). |
| `prd-12.6-squirrel-kept-116` | Money In £36, Spent £11 across Snacks 5, Games 4, Toys 2, target 60%, Streak 9: Squirrel, Splurge £4 on Games, kept 116 (acceptance 6). |

Two files carry an assumption slice 5 may revisit: the past months of the Payday 31 and `Last day`
files assert `daysRemaining` 1, the §8.7 clamp applied literally to a month already over, and
`streakDay` equal to the month's length, the no-Spends rule over a full window.
