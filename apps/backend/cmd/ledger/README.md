# ledger

The money service (`docs/ARCHITECTURE.md` §7.2): a Member's Salary, Saving Target and Spends and every
figure PRD §8 computes from them. Built by #99 (spec #92), with Spends since #128 and Recaps since #130: it serves `SaveSalary`,
`RelabelCurrency`, `SaveTarget`, `LogSpend`, `EditSpend` and `DeleteSpend` (#133), `GetSettings`, `GetMonthSummary`, `GetRecap`, `ListSpends` and
`ListMonthsWithData` (#131); `PreviewConversion` and `ApplyCurrencyConversion` (slice 7), `GetSpentToday` and
`ExportCSV` are planned. The BFF calls it since #103: Home's month,
onboarding's steps 3 and 4 (#104), a resumed step's saved answers and, since #129, every Spend screen. Its image, bearer check, start-up, migrations and variables are every service's (`apps/README.md`,
"Every service"), plus two of its own.

| Variable | Default | Meaning |
| --- | --- | --- |
| `LEDGER_MEMBERS_URL` | required | `members`, for the profile read that fills the replica: `http://members:8080` in Compose. |
| `LEDGER_AMQP_URL` | required | The broker its consumer reads `members`' events from and its outbox relay publishes the spend events to: the `amqp_url` secret in Compose. |

| Path | Holds |
| --- | --- |
| `main.go` | The binary: its name, its schema step, its handler, its consumer and its outbox relay, handed to `serve.Main`. |
| `../../internal/ledger/` | `ledger.go` (the replica and the reads), `writes.go` (the Salary and Target writes), `spends.go` (`LogSpend`, `EditSpend`, `DeleteSpend` and the events they stage), `month.go` (PRD §8's Home figures of a month), `recap.go` (`GetRecap` and Home's past-Recap list), `list.go` (`ListSpends`, `ListMonthsWithData`, the Category split and the months with data both lists share), `money.go` (a Currency's decimals and an amount inside a sentence), `consume.go` (the replica from events) and `migrations/*.sql` (embedded). |

## What it keeps

- **The replica** (`member_facts`): the Member's zone, sign-up date and Currency, which every figure
  needs. `members.created` and `members.updated` keep it fresh from the quorum queue `ledger.members`
  (#101, §4.3): each message id is recorded in `processed_events` in the transaction that applies it, so a
  second delivery changes nothing; an event older than the one applied (`facts_at`) is skipped, a Currency older
  than the last `RelabelCurrency` (`currency_at`) is too, and an empty fact never clears a stored one. A message that fails five times waits in `ledger.members.dead`
  for a person. On a miss, a fact no event has brought yet, it asks `members.GetProfile` once, under the
  caller's own token, which the service forwards (§6.2 step 3), and fills only what the row lacks, so a
  read racing a newer event cannot overwrite it. The row also keeps the subject
  whose token that read succeeded with, and a call under any other token is a miss, so `members` answers
  it: a token that does not own the Member gets `not_found`, however the row was filled (§4.2 point 3).
  A Member with no zone, sign-up date or Currency yet is `failed_precondition`.
- **The rows** (`salary_setting`, `saving_target_setting`), append-only (PRD §7). The row in force for a
  month is the one with the latest effective month not after it, and of two sharing that month the newer
  wins, which makes a step 3 save repeated after a crash safe.
- **The Spends** (`spend`): the amount and the amount as typed with its Currency, the Category, the local
  date, the note, when it was created and updated, and a deletion stamp; a deleted Spend counts nowhere.

## The rules

- A Salary or Saving Target sent with `setup`, which the BFF sets while onboarding is unfinished, takes effect
  in the sign-up month, even when onboarding ends in a later month: a first save is not an edit (spec #92), so
  D56 and D65 do not apply. Saved again, after a crash between its two writes or from step 4's back arrow, it
  replaces the setup row, a new Payday included (owner, 2026-09-23). Two rows sharing a month resolve newest first.
- The **first** Salary or Saving Target saved **after** onboarding, from Home's prompt, takes effect this
  month, so no month already recapped gains one (owner, 2026-09-23). Every later save is an edit.
- `RelabelCurrency` is onboarding's Currency change: every Salary row takes the new code, `original_currency`
  included, each amount and original rounded half up to the new decimals and never below the smallest unit, and
  the replica takes the code in the same transaction, so a save in it right after is not refused. No rate, no
  conversion row (spec #92).
- After it, a Salary follows D56 and §8.10: a new Payday starts next month, and so does an amount sent as
  next month's; one sent as this month's applies now; one sent with no answer applies now before this
  month's Payday and is `failed_precondition` from it on, which the phone answers with the D56 sheet.
- A later Saving Target starts next month (D65), and none can be set without a Salary (D44).
- An amount is in the Member's Currency, above zero, D57's 11 digits at most before the separator, and no
  more decimals than the Currency has. Every amount answered carries that Currency's CLDR decimals
  (`internal/cldr`, ARCHITECTURE §7.2), which the phone prints exactly: GBP `2400.00`, IDR `3000000`. A
  sentence names a currency that the older x/text symbols lack by its code: `VES 1.500`.
  The target amount rounds half up.
- A Spend's amount follows the Salary's rule; its Category is one of the eight; its date is in the last 60
  days including today, never ahead and never before the 1st of the sign-up month (D72), in the Member's zone (A3); its note is at most 80 characters, counted
  as NFC code points like a Display Name. `LogSpend` locks the Member's replica row and reads the Spends
  again inside its transaction, so `crossed_line` is true for exactly the log that first takes this month's
  Spent past Salary minus the target amount (PRD §8.8), never for a Spend in another month.
- `EditSpend` validates as `LogSpend` does, locks and reads again the same way and answers the month the Spend
  now sits in, `crossed_line` computed as for a log. A date outside the last 60 days (or before the sign-up
  month, D72) may stay unchanged, but no edit moves a Spend to or from one. `DeleteSpend` stamps `deleted_at`
  and answers the Spend as it was and the month it left, `crossed_line` always false. Another Member's Spend,
  or one already deleted, is `not_found`.
- A Recap is any month up to the current one (PRD §6.11 to §6.14, §8): the current month is the preview and a
  partial month (§8.11) is in its §6.17 situation, each with its closing line and no Archetype. `GetMonthSummary`
  also lists every earlier month with a Spend or Money In for Home, newest first, labelled by its Archetype
  name or `Partial`, which a full month without an Archetype reads too (§6.14).
- Every write needs an `idempotency_key`, claimed in `idempotency` inside the write's own transaction
  (`internal/idempotency`, A14), as `members` does.
- Each Spend write stages `ledger.spend.logged`, `ledger.spend.edited` or `ledger.spend.deleted` in `outbox`
  inside that same transaction (#135, ARCHITECTURE §4.3), carrying the answer's Spend, month and `crossed_line`;
  a replayed key stages nothing more. The relay beside the handler publishes it as `members`' does.

## Tests

`go test ./internal/ledger/` runs them. The money ones need the stack's `ledger` database, and the
Connect-boundary one and `TestMembersEventsReachTheReplica` also need `ledger`, `members` and a real Keycloak token (`infra/compose/README.md`,
"Running the Go tests against this stack"). `TestTheAnswerKey` runs every file in
`packages/contract/fixtures/ledger/` but the one it names as left out with its reason, the currency change, until slice 7, so 23 run. Each goes through the service and
its database, its rows seeded, the last Spend logged on its own date where the file asserts `crossedLine`,
its `saveSalary` made, the clock at its `asOf`, and every listed month, Recap, the settings and the Salary rows
after are compared whole.
