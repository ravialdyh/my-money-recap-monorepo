#!/bin/sh
set -eu
cd "$(git rev-parse --show-toplevel)"

fail=0
check() {
  stale=$(git status --porcelain -- "$@")
  if [ -n "$stale" ]; then
    printf 'generated files are stale, commit the regenerated output:\n%s\n' "$stale" >&2
    fail=1
  fi
}

if command -v npm >/dev/null 2>&1 && [ -d node_modules ]; then
  npm run --silent build:tokens >/dev/null
  npm run --silent build:swift-tokens >/dev/null
  npm run --silent build:assets >/dev/null
  check my-money-recap-ds/src/tokens my-money-recap-ds/src/styles my-money-recap-ds/src/assets apps/ios/DesignSystem
else
  printf 'check-generated: skipping the design-system generators, run npm ci first\n' >&2
fi

if command -v buf >/dev/null 2>&1; then
  (cd packages/contract && buf generate)
  check apps/backend/gen packages/contract/swift/Sources
else
  printf 'check-generated: skipping buf generate, buf not installed (CI runs it)\n' >&2
fi

exit "$fail"
