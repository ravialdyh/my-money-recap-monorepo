# Coverage gate

Test coverage may not fall. The target is 80% for Go and for iOS, and each language has a floor in
`baseline.json` that only rises. Both floors are at the target now.

## How it works

- **Fails** when coverage drops more than 0.5 points below the floor.
- **Passes** when coverage rises, and prints the command that records the new floor.
- **Stops rising** once a language reaches the target.
- **Go** leaves out the generated code in `apps/backend/gen/` and the test helpers `internal/testauth` and
  `internal/testdb`.
- **Go is measured** by CI's `backend` job, with the Compose stack up. **iOS is measured** by CI's `ios` job, on
  a simulator with no stored sign-in.

## Run it

Check a measured number:

```bash
python3 tools/coverage/coverage_gate.py go 84.4
```

Check it and raise the floor:

```bash
python3 tools/coverage/coverage_gate.py go 84.4 --update
```

When a floor is still below the target, commit the new `baseline.json` in the same pull request as the tests that
raised it.
