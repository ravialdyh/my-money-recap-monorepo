# Size and comment policy

`check_policy.py` is the custom hook behind the repo's two house rules. It runs on every commit
(`.pre-commit-config.yaml`) and again over every tracked file in CI.

```
python3 tools/policy/check_policy.py                 # every tracked file
python3 tools/policy/check_policy.py path/to/file.go # just these
```

`policy.json` holds the numbers and the exclusions. `scanner.py` is a tokeniser that knows string
literals, so `"http://localhost:8080"` is never mistaken for a comment, and knows four comment
syntaxes: `// /// /* */` for Go, Swift, TypeScript, JavaScript and protobuf, `#` for YAML, TOML,
shell, Python, Dockerfile, `.conf`, the ignore files and the other dotfile configs, `--` for SQL,
and `/* */` for CSS.

## The rules

**A file is at most 300 lines.** Every text file outside `exclude` (below), whether or not the
scanner knows its comment syntax.

**No em dash in anything a Member can read.** A `—` inside a string literal of a Swift, Go, TypeScript,
JavaScript or protobuf file, or anywhere in an HTML page, is rejected: it reads as unprofessional in the
app (owner, 2026-09-22). Comments may still use one; copy uses a full stop or a comma instead.

**A comment is either a directive or a docstring on a declaration. Nothing else survives.**

Concretely, all of these are rejected:

| Rejected | Instead |
| --- | --- |
| A comment attached to no declaration — a file header, a section banner, a stray note | Delete it. If it carries a decision, the decision belongs in a README or in `docs/`. |
| A trailing comment after code on the same line | Delete it, or rename the thing it was explaining. `const field = useRef() // the padded row` became `const paddedRow = useRef()`. |
| A comment inside a function body | Name the code. Extract the step into a function whose name says what the comment said — `lockMigrationLedger`, `daysRemainingInclusive`. |
| A Python module docstring | The file's name and its code say what it is. |
| A docstring over 255 characters | State the contract; put the reasoning in the nearest README. |
| `TODO`, `FIXME`, `XXX`, `HACK`, `WIP` | Open a GitHub issue. A tag in the source has no owner and no date. |
| Commented-out code | Delete it. Git remembers. |

Two things survive.

**Docstrings on a declaration**, at most 255 characters: Go package docs (`// Package bff …`), Go doc
comments on a `func`/`type`/`var`/`const`, Swift `///` on a type or member, TSDoc on an export, and
a comment on a field inside a struct or class body. This is where a `D*`, `A*`, `I*` or `§` citation
lives, which is the only machine-greppable link from code back to the decision that caused it
(`AGENTS.md`). A block counts as a docstring only when the next non-blank, non-comment line opens a
declaration; a block followed by a blank line, or by nothing, is a file header and is rejected.

**Directives**, which are instructions to a compiler or a tool rather than prose. The list is
deliberately short, and every entry is there because deleting it breaks something:

| Directive | What breaks without it |
| --- | --- |
| `#!/bin/sh` and the other shebangs | `execve` fails with `ENOEXEC`. pre-commit reads the shebang itself to build the command, so deleting the one in `tools/hooks/if-installed.sh` kills the gofmt, golangci-lint, swiftformat, swiftlint, hadolint and buf-lint hooks at once. Two pre-commit hooks also enforce it both ways. |
| `// swift-tools-version:` | SwiftPM cannot load `packages/contract/swift/Package.swift` at all. |
| `//go:embed` | It **is** the migration bundle. `internal/members` still compiles and still vets without it — the `embed.FS` is simply empty, `ReadDir` fails at run time, `members-migrate` exits 1, and the Compose stack never comes up. CI catches it, but only at the integration step. |
| `//go:build`, `// +build` | Build constraints; the file silently joins or leaves the build. |
| `# shellcheck disable=SC1091` | `infra/compose/a10-smoke.sh` sources `.env`, which is git-ignored and absent at lint time. The shellcheck hook fails without the waiver. |
| `//nolint`, `// swiftlint:`, `// swiftformat:`, `@ts-expect-error`, `eslint-disable`, `biome-ignore`, `prettier-ignore`, `c8/v8/istanbul ignore` | Linter waivers. `.golangci.yml` sets `nolintlint` to require both an explanation and a specific linter, so a bare `//nolint` is already rejected by golangci-lint. |
| `# syntax=`, `# escape=`, `# hadolint ignore=` | Docker parser directives; the first two change how the Dockerfile is parsed. |
| `# type:`, `# noqa`, `# pragma:`, `# mypy:`, `# ruff:` | Python tool waivers. |
| `# vN` after `uses: owner/action@<sha>` | Dependabot maintains these alongside the pin. Strip them and Dependabot's next PR puts them back, which the policy would then reject — a loop no commit can leave. |

A comment that merely restates the code is out of scope for the hook — no linter can see it. That one
is a review concern.

## What is skipped, and how

Three mechanisms, in this order. None of them repeats another.

**No known comment syntax** skips the comment rules, not the cap: an HTML page, a JSON fixture or a
`.tf` file is still held to 300 lines. `.html` has no syntax on purpose — a page mixes `<!-- -->`
with `//` and `/* */` inside `<script>` and `<style>`, which one syntax per extension cannot scan. A
file that is not UTF-8 text, a PNG, is skipped outright.

**The generated-file header.** A file with `Code generated by`, `@generated`, `DO NOT EDIT`,
`Do not edit by hand` or `Generated from` **inside a comment** in its first five lines is skipped:
everything buf writes, plus `Tokens.swift`, `css.generated.ts` and `illustrations.ts`. The marker has
to be in a comment, not merely somewhere in the first five lines — otherwise a string literal like
`const helpText = "Ask before you DO NOT EDIT the ledger"` would exempt its whole file, which is a
bypass the earlier substring match allowed. A file with no comment syntax cannot carry the marker,
so a generated one of those goes in `exclude`.

**`exclude`**, which holds three kinds of path:

- `my-money-recap/**` (the illustration pipeline — one-off tooling, not shipped code),
  `my-money-recap-ds/.design-sync/**` (owned by the design-sync tool) and `docs/**`.
- Files a tool rewrites and nobody can shorten by hand: `package-lock.json`, `*/go.sum`,
  `*/.terraform.lock.hcl`, and `*.xcodeproj/*` (the Xcode project, its scheme and
  `Package.resolved`). The lockfiles grow with every dependency bump.
- `*.md`: the cap is for code and configuration, and prose is not held to it.

`.pre-commit-config.yaml` carries a longer list, written once as a YAML anchor and shared by the seven
hooks that format or lint generated output. That is not a copy of this one: those hooks touch JSON, CSS
and Markdown, and generated files carry trailing whitespace exactly as their generator emits it. It is
attached per hook rather than globally so that `detect-private-key`, `gitleaks` and
`check-added-large-files` still see every file in the repo.
