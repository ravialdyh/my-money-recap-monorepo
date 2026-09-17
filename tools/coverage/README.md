# Coverage ratchet

The target is 80% for every language. The repo is not there, so the gate holds a floor per language
and the floor only ever goes up.

```
python3 tools/coverage/coverage_gate.py go 42.1            # check
python3 tools/coverage/coverage_gate.py go 42.1 --update   # check and raise the floor
```

It fails when coverage drops more than 0.5 points below the floor. It passes when coverage rises,
and prints the command that records the new floor — run it and commit `baseline.json` in the same
pull request that added the tests. Once a language reaches 80% the floor becomes 80% and stays.

## The floors, and where they came from

All three come from a CI run, not a guess.

| Language | Floor | Where it comes from |
| --- | --- | --- |
| `go` | 52.9% | The backend job, with the Compose stack up, so `internal/members` and the end-to-end test really run. `internal/serve` is at 100%. |
| `ios` | 30.9% | The iOS job on a macOS runner, ten tests, `-enableCodeCoverage YES` on both halves of the build-then-test split. |
| `typescript` | 80% | `npm run test:coverage` — 134 tests over the components, hooks and Recap screens; measured 97.8%, and the floor sits at the target because the ratchet stops there. |

When a run beats its floor the gate passes and prints the command that records the new one:

```
go: 52.9%, up from 46.8%. Run python3 tools/coverage/coverage_gate.py go 52.9 --update and commit the floor.
```

Run it and commit `baseline.json` in the same pull request that added the tests.

## What is not counted

Go drops `apps/backend/gen/**` (buf's output) and `internal/testauth` (a test helper) from the profile
before measuring. TypeScript drops `src/assets/illustrations.ts` and `src/styles/css.generated.ts`.
Generated code is not code anyone writes tests for, and leaving it in the denominator makes the number
say nothing.

## Getting to 80%

TypeScript is there: 97.8% of statements and 90% of branches, over nine test files. The suite renders
components rather than only calling functions, so the fit-to-width measurement, the count-up animation
and every event handler are exercised. Go's remaining gap is `internal/bff`'s unexercised branches;
`internal/members` and `internal/serve` are covered. iOS at 30.9% is ten tests over the view models,
the Keychain and money formatting, with the views untested.

Nothing forces that into one change. Add tests with the slice that touches the code, raise the floor
in the same pull request, and the ratchet keeps what you gained.
