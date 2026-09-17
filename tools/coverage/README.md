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
go: 52.9%, up from 46.8%. Run python3 tools/coverage/coverage_gate.py go 52.9 --update and commit the floor.
```

Run it and commit `baseline.json` in the same pull request that added the tests.

## The floors, and where they came from

All three come from a CI run, not a guess.

| Language | Floor | Measured | Where it comes from |
| --- | --- | --- | --- |
| `go` | 52.9% | 52.9% | The backend job, with the Compose stack up, so `internal/members` and the end-to-end test really run. `internal/serve` is at 100%. |
| `ios` | 30.9% | 30.9% | The iOS job on a macOS runner, ten tests, `-enableCodeCoverage YES` on both halves of the build-then-test split. |
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

Go's gap is `internal/bff`'s unexercised branches; `internal/members` and `internal/serve` are covered.
iOS at 30.9% is ten tests over the view models, the Keychain and money formatting, with the views
untested.

Nothing forces those into one change. Add tests with the slice that touches the code, raise the floor
in the same pull request, and the ratchet keeps what you gained.
