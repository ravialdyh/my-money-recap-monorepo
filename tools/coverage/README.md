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

## The floors, and how each is measured

The floors are `baseline.json`, which only `--update` writes. Each comes from a measured run, not a
guess: CI, or the CI commands run locally when GitHub refuses to start the jobs. The history of the
measurements is in the pull requests that raised them.

| Language | Measured by | Latest |
| --- | --- | --- |
| `go` | The `backend` job's commands, with the Compose stack up and Mailpit started, so `internal/members`, `internal/notify` and the three end-to-end tests really run. The floor sits at the target, not at the measurement, because the ratchet stops there. | 82.4% |
| `ios` | The `ios` job's commands, `-enableCodeCoverage YES` on both halves of the build-then-test split, on a simulator with no stored session, as CI's is: the test host is the app, and a signed-in one draws Home and counts far higher. When two clean runs differ, the floor takes the lower. | 61.4% |
| `typescript` | `npm run test:coverage` in `my-money-recap-ds`. The floor sits at the target, as Go's does. | 97.8% |

## What is not counted

Go drops `apps/backend/gen/**` (buf's output) and `internal/testauth` (a test helper) from the profile
before measuring. TypeScript drops `src/assets/illustrations.ts` and `src/styles/css.generated.ts`;
those exclusions live in `my-money-recap-ds/vitest.config.ts`. Generated code is not code anyone
writes tests for, and leaving it in the denominator makes the number say nothing.

## Where the gap is

TypeScript has none left worth chasing: the suite renders components rather than only calling
functions, so the fit-to-width measurement, the count-up animation and every event handler are
exercised.

Go's gap is `internal/members`, whose handlers the boundary test reaches in the Compose container,
outside the profile, and `cmd/bff`, `cmd/members` and `cmd/notify`: their `main`, `run` and handler
wiring, which only the Compose stack runs. `serve.Main`'s serve path, which needs a listener and a
signal rather than a return value, is `internal/serve`'s gap. `internal/migrate`'s uncovered lines are
store errors a live Postgres does not raise, and `internal/notify`'s are the STARTTLS upgrade no local
mail host offers and write errors a socket under a live deadline does not reach. `internal/ledger`
holds only the fixture decoding test until `ledger` is built.

iOS's gap is the views: the unit tests cover the view models, sign-in and the session, the Keychain
and money formatting (`apps/ios/README.md`, "Layout"), and no unit test renders a view.

Nothing forces those into one change. Add tests with the slice that touches the code, raise the floor
in the same pull request, and the ratchet keeps what you gained.
