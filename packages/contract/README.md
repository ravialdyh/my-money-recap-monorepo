# packages/contract

The one contract every client is generated from (`docs/ARCHITECTURE.md` §4.1, A1).
A Buf module, package `money_recap.v1`, under `money_recap/v1/*.proto`:

| File | Holds |
| --- | --- |
| `money.proto` | `Money`: a decimal-string amount plus an ISO 4217 code, never a float or minor units (PRD §7, D71) |
| `members.proto` | `MembersService`: `GetOrCreateBySubject`, `GetProfile`; the `Profile` message |
| `bff.proto` | `BffService`: `GetHome`; the `HomeView` message with every field Home draws (PRD §6.8) |
| `shared.proto` | What every slice 2 file reuses: the `Category`, `Archetype`, `SalaryAppliesFrom` and `RateSource` enums, `Spend`, `EventEnvelope` (A4); the idempotency convention is the doc comment on its `package` line (A14) |

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
