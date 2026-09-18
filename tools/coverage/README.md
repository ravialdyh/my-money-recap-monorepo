# Coverage ratchet

The target is 80% for every language. The gate holds a floor per language, the floor only ever rises,
and it stops once a language reaches the target.

```
python3 tools/coverage/coverage_gate.py go 42.1            # check
python3 tools/coverage/coverage_gate.py go 42.1 --update   # check and raise the floor
```

It fails when coverage drops more than 0.5 points below the floor, and passes when coverage rises,
printing the command that records the new one:

```
go: 53.6%, up from 52.9%. Run python3 tools/coverage/coverage_gate.py go 53.6 --update and commit the floor.
```

Run it and commit `baseline.json` in the same pull request that added the tests.

## The floors, and where they came from

All three come from a measured run, not a guess: CI, or the CI commands run locally when GitHub refuses to start the jobs.

| Language | Floor | Measured | Where it comes from |
| --- | --- | --- | --- |
| `go` | 65.6% | 65.6% | The backend job's commands, with the Compose stack up, so `internal/members` and the end-to-end test really run. 52.9% came from a CI run and 53.6% from the same commands run locally on PR #56 while GitHub refused to start the jobs (AGENTS.md exception); 65.6% was measured locally the same way on `fix/backend-and-ios`. `internal/serve` is at 100%, `internal/bff` at 96.7%, `internal/auth` at 85.1%. |
| `ios` | 37.7% | 37.7% | The iOS job's commands, thirteen tests, `-enableCodeCoverage YES` on both halves of the build-then-test split. 30.9% came from a macOS runner; 37.7% was measured locally the same way on `fix/backend-and-ios`. |
| `typescript` | 80% | 97.8% | `npm run test:coverage` — 134 tests across nine files, over the components, the hooks and the four Recap screens. The floor sits at the target, not at the measurement, because the ratchet stops there. |

## What is not counted

Go drops `apps/backend/gen/**` (buf's output) and `internal/testauth` (a test helper) from the profile
before measuring. TypeScript drops `src/assets/illustrations.ts` and `src/styles/css.generated.ts`;
those exclusions live in `my-money-recap-ds/vitest.config.ts`. Generated code is not code anyone
writes tests for, and leaving it in the denominator makes the number say nothing.

## Where the gap is

TypeScript has none left worth chasing: 97.8% of statements and 90% of branches. The suite renders
components rather than only calling functions, so the fit-to-width measurement, the count-up animation
and every event handler are exercised.

Go's gap is `internal/members` at 70.6%, whose handlers the boundary test reaches in the Compose
container, outside the profile, and the two `cmd/` binaries at 0%; `internal/bff` is at 96.7%,
`internal/auth` at 85.1% and `internal/serve` at 100%. `internal/ledger` holds only the fixture decoding test until slice 5.
iOS at 37.7% is thirteen tests over the view models, the Keychain, money formatting and the token
refresh (`AuthSession`, `BearerInterceptor`), with the views untested.

Nothing forces those into one change. Add tests with the slice that touches the code, raise the floor
in the same pull request, and the ratchet keeps what you gained.
