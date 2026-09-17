#!/bin/sh
# Runs a hook only where its tool exists. SwiftLint, SwiftFormat and golangci-lint are per-language
# toolchains; a contributor touching only the design system has no reason to install all three.
# CI installs every one of them and runs the same commands as hard gates (.github/workflows/ci.yml),
# so a skip here is a local convenience, never a way for unlinted code to reach main.
set -eu
tool="$1"
shift
if ! command -v "$tool" >/dev/null 2>&1; then
  printf 'pre-commit: skipping %s, not installed locally (CI runs it)\n' "$tool" >&2
  exit 0
fi
exec "$tool" "$@"
