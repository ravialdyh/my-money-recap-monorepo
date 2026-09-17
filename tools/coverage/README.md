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

| Language | Floor | Measured |
| --- | --- | --- |
| `go` | 36.0% | `go test ./...` with no Docker, so the members and end-to-end suites skipped. |
| `typescript` | 8.6% | `npm run test:coverage` — one test file, `money.test.ts`, against 33 modules. |
| `ios` | 0.0% | Not measured: no Xcode outside the macOS runner. |

**The Go and iOS floors are placeholders and the first CI run will beat them.** The backend job brings
up the Compose stack, so `internal/members` and the end-to-end test actually execute there; the iOS job
now passes `-enableCodeCoverage YES`. Both will print `up from …, run … --update and commit the floor`.
Do that once and the real numbers are locked in.

## What is not counted

Go drops `apps/backend/gen/**` (buf's output) and `internal/testauth` (a test helper) from the profile
before measuring. TypeScript drops `src/assets/illustrations.ts` and `src/styles/css.generated.ts`.
Generated code is not code anyone writes tests for, and leaving it in the denominator makes the number
say nothing.

## Getting to 80%

The gap is almost entirely the design system: 26 components with no tests at all, and the `internal/serve`
and `internal/members` Go packages whose tests only run with Compose up. Nothing here forces that work into
one change — add tests with the slice that touches the code, raise the floor, and the ratchet keeps it.
