# packages/contract

The one contract every client is generated from (`docs/ARCHITECTURE.md` §4.1, A1).
A Buf module, package `money_recap.v1`, under `money_recap/v1/*.proto`:

| File | Holds |
| --- | --- |
| `money.proto` | `Money`: a decimal-string amount plus an ISO 4217 code, never a float or minor units (PRD §7, D71) |
| `shared.proto` | What every file reuses: the enums, `Spend`, `EventEnvelope` (A4) and the view pieces both the BFF and ledger carry; the idempotency convention is the doc comment on its `package` line (A14) |
| `bff.proto` | `BffService`, the one service the phone talks to (ARCHITECTURE §6) |
| `bff_views.proto` | The screen answers the phone stores whole (I4) |
| `bff_reads.proto`, `bff_writes.proto`, `bff_account.proto` | `BffService`'s request and response pairs: the reads, the writes that answer a fresh `HomeView` (§6.2 step 5), and the account writes |
| `members.proto`, `members_events.proto` | `MembersService` and `Profile` (ARCHITECTURE §7.1); `MemberFacts` and the `members.*` events |
| `ledger.proto`, `ledger_views.proto`, `ledger_writes.proto`, `ledger_events.proto` | `LedgerService` (§7.2): its reads, what it computes, its writes (each opening with `idempotency_key`, A14) and the `ledger.*` events |
| `ledger_fixture.proto` | `LedgerFixture`, test-only (A2): the shape of every file in `fixtures/ledger/` (see its README) |
| `notify.proto`, `notify_events.proto` | `NotifyService` (§7.3), `NotificationPrefs`, the `EmailBounced` error detail; `notify.member.purged` |
| `currency.proto` | `CurrencyService` (§7.4) and the `UnsupportedPair` error detail (D35); `currency` publishes no event |

The `.proto` files are the inventory: each service block lists its RPCs, and each message says what
it is for. The phone-facing request and response messages carry the `BffService` prefix the
`STANDARD` lint rules require, because the service RPCs of ARCHITECTURE §7 already own the bare names
in the same package.

## Generated outputs, checked in

| Output | Where | Consumer |
| --- | --- | --- |
| Go messages and Connect stubs/clients | `apps/backend/gen/money_recap/v1/` | the BFF and the services (one Go module at `apps/backend`) |
| Swift: the `BffService` client and the messages it reaches, plus `EmailBounced` (`types:` in `buf.gen.yaml`); the service-to-service RPCs, events and fixture schema are Go only | `swift/Sources/MoneyRecapContract/` (local Swift package `MoneyRecapContract`) | the iOS app (`docs/IOS.md` I3), which talks to the BFF alone (§6) |

Nothing under those paths is edited by hand. Plugin versions are pinned in `buf.gen.yaml`
so the output is reproducible. A plugin bump moves its runtime pin too: `apps/backend/go.mod` for Go;
`swift/Package.swift` and the Xcode project's `connect-swift` and `swift-protobuf` references
(`apps/ios/README.md`) for Swift.

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

## Rules CI enforces on every contract change (`.github/workflows/ci.yml`, job `contract`)

- `buf lint` (the `STANDARD` rule set).
- `buf breaking --against` the `main` branch, `PACKAGE` rules: a renamed, removed or retyped field fails the build.
  Add fields; never reuse a number or rename one. A removal needs the owner's acceptance, recorded on its pull request
  with the `buf breaking` report (PR #60, PR #161), and a removed field's number and name are `reserved`. A message may
  move to another file of the package, since Go and Swift see one package either way; the cross-file cycle #54 hit is
  why it is `PACKAGE`, not `FILE`.
- `buf generate` then a clean tree under the two output paths, new files included: stale
  generated code fails the build, the same freshness rule the design tokens follow. The `backend` job compiles the Go output.
