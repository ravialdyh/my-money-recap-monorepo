# Ledger fixtures

The answer key for PRD §8. Every worked example in §8 and every numeric acceptance criterion in
§12 is one JSON file here, typed by `LedgerFixture` (`money_recap/v1/ledger_fixture.proto`), so
the arithmetic that only `ledger` implements (ARCHITECTURE A2, §7.2) is checked against numbers
the PRD wrote down, not numbers the implementation produced.

Written in #42, #48 and #49. `apps/backend/internal/ledger/fixtures_test.go` decodes every file, and
`ledger`'s `TestTheAnswerKey` (#99, #128, #130, #152) runs them through the service: it seeds the file's rows, logs
the last Spend with `LogSpend` on its own date when the file asserts `crossedLine`, makes its `saveSalary`,
sets the clock to `asOf` in `timezone` and compares what `GetMonthSummary`, `GetRecap`, `GetSettings` and the
Salary rows answer to the expected fields. It names the one file it leaves out and why: the currency change,
`prd-8.9-maya-gbp-to-idr`, until slice 7.

## What a fixture is

One ledger and one "as of" date, in protobuf JSON (`lowerCamelCase` names, enums as their names).
What each field holds is its doc comment in `ledger_fixture.proto`; `spends` holds nothing dated
after `asOf`. The rules below are the ones no field doc states.

How `ledger`'s tests compare an answer to an expected message: a top-level expected field that is left
out (`recaps` in a Home-only file) asserts nothing; inside a listed message every field is
compared, so an `optional` the PRD says is absent (Saved before Payday, D41; the Archetype of a
Partial Recap, §8.6) is left out of the file and must be absent in the answer. Three things are
never compared: `Spend.id`, which `ledger` assigns (`maya-01` is a label that pairs an input row
with `splurge` or `spendsAfter`), and `createdAt` and `updatedAt`, which `ledger` stamps.

Amounts follow `Money`: a decimal string with at most 11 digits before the separator (D57) and an
ISO 4217 code; GBP carries two decimals (`"2400.00"`), IDR none (`"15000000"`), the decimals CLDR
reports for the currency (§8.4). A sentence that embeds an amount is written as the PRD prints it
(`£160`, `Rp2.600.000`).

Every `SalarySetting` and every `Spend`, wherever it sits, carries `originalAmount`: the amount as
typed, with its Currency (PRD §7, §8.9 step 4). Every file so far starts from a ledger never
converted, so its input rows `salaries` and `spends`, and every row of a file without `convert`,
have `originalAmount` equal to `amount`. That is a limit of today's files, not a rule of the
ledger: converting again (§8.9 step 6) starts from Rp amounts with £ originals, and a file starting
there, the only shape that catches a conversion from `originalAmount` instead of `amount`, widens
this check in `fixtures_test.go` first. After `convert`, every other row (`settings.salary`,
`settings.nextSalary`, `salariesAfter`, `spendsAfter`, a Recap's `splurge`) keeps its input row's
`originalAmount`, a `SalarySetting` paired by `effectiveFrom` and a `Spend` by `id`; a row with no
input row fails. `fixtures_test.go` checks all three and names the row that drifts.

Every file that asserts `settings` names `newSavingTargetFrom`, the month a Saving Target saved on `asOf` would
start in (D65, #172): `asOf`'s own month in a file with no `savingTargets`, the month after in every other file, since
each of their rows counts as saved after setup.

`streakDay` on a month is the month's Streak so far (§8.5, the glossary's Streak): the longest
spend-free run in the window up to `asOf`. With no Spends it is the elapsed days, which the empty Spend list prints
as "Streak day {n}" (SCREEN-STATES B6: 23 on 26 September for a 4 September sign-up). `streak.days`
on a Recap is the same rule over the whole window, and `streak.runStart` and `streak.runEnd` name that run's first
and last day, the earlier of two equally long runs, both left out when the Streak is 0; every file with an Archetype
asserts its `archetypeTitle`. A scalar
left out of a file is its zero value to protobuf JSON, so `partial: false` and an absent `partial`
are the same assertion; only a message or `optional` field can be "absent".

## Adding one

1. Name the file after what it proves: `prd-<section>-<what it proves>.json`, for example
   `prd-8.5-signup-4-september-first-run.json` or `prd-12.6-squirrel-kept-116.json`.
2. Work every expected number from the cited paragraph and say how in `doc`.
3. Add the name to `manifest` in `apps/backend/internal/ledger/fixtures_test.go`, keeping it
   alphabetical; the test fails on a file without a line and a line without a file.
4. `cd apps/backend && go test ./internal/ledger/`.

## The set

`manifest` in `fixtures_test.go` is the list, and each file's `doc` says what it proves.

The past months of the Payday 31 and `Last day` files assert `daysRemaining` 1, the §8.7 clamp applied
literally to a month already over, and `streakDay` equal to the month's length, the no-Spends rule
over a full window. `ledger` answers both that way (#99), so the reading stands.
