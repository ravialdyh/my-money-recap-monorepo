#!/bin/sh
set -eu
cd "$(git rev-parse --show-toplevel)"

fail=0
check() {
  stale=$(git status --porcelain -- "$@")
  if [ -n "$stale" ]; then
    printf 'generated files are stale, commit the regenerated output:\n%s\n' "$stale" >&2
    git --no-pager diff -- "$@" >&2
    fail=1
  fi
}

ds() {
  npm run --silent build:tokens >/dev/null
  npm run --silent build:swift-tokens >/dev/null
  npm run --silent build:assets >/dev/null
  check my-money-recap-ds/src/tokens/tokens.css my-money-recap-ds/src/styles/css.generated.ts \
    my-money-recap-ds/src/assets/illustrations.ts apps/ios/DesignSystem/Tokens.swift
}

contract() {
  (cd packages/contract && buf generate)
  check apps/backend/gen packages/contract/swift/Sources
}

case ${1:-} in
  ds) ds ;;
  contract) contract ;;
  "")
    if command -v npm >/dev/null 2>&1 && [ -d node_modules ]; then
      ds
    else
      printf 'check-generated: skipping the design-system generators, run npm ci first\n' >&2
    fi
    if command -v buf >/dev/null 2>&1; then
      contract
    else
      printf 'check-generated: skipping buf generate, buf not installed (CI runs it)\n' >&2
    fi
    ;;
  *)
    printf 'usage: %s [ds|contract]\n' "$0" >&2
    exit 2
    ;;
esac

exit "$fail"
