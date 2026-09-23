# ledger

The money service (`docs/ARCHITECTURE.md` §7.2): a Member's Salary and Saving Target and every figure
PRD §8 computes from them. Built in slice 4 (#99) for a Member who has no Spends yet, since nothing can
log one before slice 5: it serves `SaveSalary`, `RelabelCurrency`, `SaveTarget`, `GetSettings` and
`GetMonthSummary`, and the rest of `LedgerService` is planned. The BFF calls it since #103: Home's month,
onboarding's step 3 and a resumed step 3's saved Salary; step 4 follows (#104). Its image, bearer check, start-up, migrations and variables are every service's (`apps/README.md`,
"Every service"), plus two of its own.

| Variable | Default | Meaning |
| --- | --- | --- |
| `LEDGER_MEMBERS_URL` | required | `members`, for the profile read that fills the replica: `http://members:8080` in Compose. |
| `LEDGER_AMQP_URL` | required | The broker its consumer reads `members`' events from: the `amqp_url` secret in Compose. |

| Path | Holds |
| --- | --- |
| `main.go` | The binary: its name, its schema step, its handler and its consumer, handed to `serve.Main`. |
| `../../internal/ledger/` | `ledger.go` (the replica and the reads), `writes.go` (the two writes), `month.go` (PRD §8 for a month without Spends), `money.go` (a Currency's decimals and an amount inside a sentence), `consume.go` (the replica from events) and `migrations/*.sql` (embedded). |

## What it keeps

- **The replica** (`member_facts`): the Member's zone, sign-up date and Currency, which every figure
  needs. `members.created` and `members.updated` keep it fresh from the quorum queue `ledger.members`
  (#101, §4.3): each message id is recorded in `processed_events` in the transaction that applies it, so a
  second delivery changes nothing; an event older than the one applied (`facts_at`) is skipped, and an
  empty fact never clears a stored one. A message that fails five times waits in `ledger.members.dead`
  for a person. On a miss, a fact no event has brought yet, it asks `members.GetProfile` once, under the
  caller's own token, which the service forwards (§6.2 step 3), and fills only what the row lacks, so a
  read racing a newer event cannot overwrite it. The row also keeps the subject
  whose token that read succeeded with, and a call under any other token is a miss, so `members` answers
  it: a token that does not own the Member gets `not_found`, however the row was filled (§4.2 point 3).
  A Member with no zone, sign-up date or Currency yet is `failed_precondition`.
- **The rows** (`salary_setting`, `saving_target_setting`), append-only (PRD §7). The row in force for a
  month is the one with the latest effective month not after it, and of two sharing that month the newer
  wins, which makes a step 3 save repeated after a crash safe.

## The rules

- The **first** Salary and the **first** Saving Target take effect in the sign-up month: a first save is not
  an edit (spec #92), so D56 and D65 do not apply to it, even when onboarding ends in a later month.
- So does every Salary sent with `setup`, which the BFF sets while onboarding is unfinished: step 3 saved
  again, after a crash between its two writes or from step 4's back arrow, replaces the setup Salary, a new
  Payday included (owner, 2026-09-23). Two rows sharing a start month resolve newest first.
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
- Every write needs an `idempotency_key`, claimed in `idempotency` inside the write's own transaction
  (`internal/idempotency`, A14), as `members` does.

## Tests

`go test ./internal/ledger/` runs them. The money ones need the stack's `ledger` database, and the
Connect-boundary one and `TestMembersEventsReachTheReplica` also need `ledger`, `members` and a real Keycloak token (`infra/compose/README.md`,
"Running the Go tests against this stack"). `TestTheAnswerKeyWithoutSpends` runs every file in
`packages/contract/fixtures/ledger/` that has no Spends, SaveSalary, conversion or Recap: seven today.
Each goes through the service and its database, its rows seeded and the clock at its `asOf`, and every
listed month and the settings are compared whole.
