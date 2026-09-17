# packages/contract

The one contract every client is generated from (`docs/ARCHITECTURE.md` §4.1, A1).
A Buf module, package `money_recap.v1`, under `money_recap/v1/*.proto`:

| File | Holds |
| --- | --- |
| `money.proto` | `Money`: a decimal-string amount plus an ISO 4217 code, never a float or minor units (PRD §7, D71) |
| `members.proto` | `MembersService`: `GetOrCreateBySubject`, `GetProfile`, `SaveProfile`, `SaveOnboardingStep`, `SetTimezone`, `RequestDeletion`, `CancelDeletion` (ARCHITECTURE §7.1); the `Profile` message with onboarding step, `signed_up_on` and `deletion_requested_at` (D60) |
| `members_events.proto` | `MemberFacts` (the replica ledger and notify keep) and the five `members.*` events, each embedding `EventEnvelope`: `MembersCreated`, `MembersUpdated`, `MembersDeletionRequested`, `MembersDeletionCancelled`, `MembersDeleted` |
| `bff.proto` | `BffService`: `GetHome`; the `HomeView` message with every field Home draws (PRD §6.8) |
| `ledger.proto` | `LedgerService`: the fourteen RPCs of ARCHITECTURE §7.2 with their read requests and responses (`GetSpentToday` over a batch of Member ids, `ExportCSV` as bytes) |
| `ledger_views.proto` | What ledger computes and the BFF stores whole (A2): `MonthSummary` (every Home figure plus days remaining and the partial flag), `Recap` for the full and the Partial Recap (PRD §6.11 to §6.14, §6.17), `LedgerSettings`, `ConversionPreview` |
| `ledger_writes.proto` | The six ledger write requests and responses; each request opens with `idempotency_key` (A14), each response carries the recomputed month, spend writes carry `crossed_line` (PRD §8.8) |
| `ledger_events.proto` | The seven `ledger.*` events, each embedding `EventEnvelope`; spend events carry the Spend, the month and `crossed_line` (§7.2 "Publishes") |
| `shared.proto` | What every slice 2 file reuses: the `Category`, `Archetype`, `SalaryAppliesFrom` and `RateSource` enums, `Spend`, `EventEnvelope` (A4); the idempotency convention is the doc comment on its `package` line (A14) |
| `notify.proto` | `NotifyService` (§7.3): `GetWarningState`, `EvaluateWarning`, `GetPrefs`, `SavePrefs`, `RegisterDevice`, `EmailMagicLink`, `EmailExport`; `NotificationPrefs`; the `EmailBounced` error detail behind the PRD §6.16 inline error |
| `notify_events.proto` | `NotifyMemberPurged`, routing key `notify.member.purged`, embedding `EventEnvelope` |
| `currency.proto` | `CurrencyService` (§7.4): `Quote` returning rate, `RateSource` and as-of date; the `UnsupportedPair` error detail (D35, PRD §8.9 step 1) |
| `currency_events.proto` | `CurrencyMemberPurged`, routing key `currency.member.purged`, embedding `EventEnvelope` |

Slice 1 defines only what the walking skeleton calls (#10). Slice 2 adds the other RPCs and the event messages (#42); `shared.proto` is its first piece (#50).

## Generated outputs, checked in

| Output | Where | Consumer |
| --- | --- | --- |
| Go messages and Connect stubs/clients | `apps/backend/gen/money_recap/v1/` | the BFF and the services (one Go module at `apps/backend`) |
| Swift messages and `connect-swift` client | `swift/Sources/MoneyRecapContract/` (local Swift package `MoneyRecapContract`) | the iOS app (`docs/IOS.md` I3) |

Nothing under those paths is edited by hand. Plugin versions are pinned in `buf.gen.yaml`
so the output is reproducible; the Swift package pins the matching runtime versions.

## Regenerate

From the repo root, after editing a `.proto`:

```bash
cd packages/contract && buf lint && buf generate
```

Then `cd apps/backend && go build ./...` to confirm the Go side compiles, and
`cd packages/contract/swift && swift build` for the Swift side (needs Xcode).

To run the breaking check locally before pushing (the `git rev-parse` form also works from a worktree):

```bash
cd packages/contract && buf breaking --against "$(git rev-parse --git-common-dir)#ref=origin/main,subdir=packages/contract"
```

## Rules CI enforces on every pull request (`.github/workflows/ci.yml`, job `contract`)

- `buf lint` (the `STANDARD` rule set).
- `buf breaking --against` the `main` branch: a renamed, removed or retyped field fails the build.
  Add fields; never reuse a number or rename one.
- `buf generate` then `git diff --exit-code`: stale generated code fails the build,
  the same freshness rule the design tokens follow. The `backend` job compiles the Go output.
