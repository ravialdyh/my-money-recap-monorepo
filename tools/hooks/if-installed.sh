#!/bin/sh
# Runs a hook only where its tool exists, optionally from another directory (-C <dir>, relative to the
# repo root). SwiftLint, SwiftFormat, golangci-lint, hadolint and buf are per-language toolchains; a
# contributor touching only the design system has no reason to install all of them. CI installs every
# one and runs the same commands as hard gates (.github/workflows/ci.yml), so a skip here is a local
# convenience, never a way for unlinted code to reach main.
set -eu
cd "$(git rev-parse --show-toplevel)"
if [ "$1" = "-C" ]; then
  workdir="$2"
  shift 2
else
  workdir="."
fi
tool="$1"
shift
if ! command -v "$tool" >/dev/null 2>&1; then
  printf 'pre-commit: skipping %s, not installed locally (CI runs it)\n' "$tool" >&2
  exit 0
fi
cd "$workdir"
exec "$tool" "$@"
