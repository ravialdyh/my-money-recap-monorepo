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

tokens() {
  node design-tokens/build-swift-tokens.mjs >/dev/null
  check apps/ios/DesignSystem/Tokens.swift
}

contract() {
  (cd packages/contract && buf generate)
  check apps/backend/gen packages/contract/swift/Sources
}

case ${1:-} in
  tokens) tokens ;;
  contract) contract ;;
  "")
    if command -v node >/dev/null 2>&1; then
      tokens
    else
      printf 'check-generated: skipping the token generator, node not installed (CI runs it)\n' >&2
    fi
    if command -v buf >/dev/null 2>&1; then
      contract
    else
      printf 'check-generated: skipping buf generate, buf not installed (CI runs it)\n' >&2
    fi
    ;;
  *)
    printf 'usage: %s [tokens|contract]\n' "$0" >&2
    exit 2
    ;;
esac

exit "$fail"
