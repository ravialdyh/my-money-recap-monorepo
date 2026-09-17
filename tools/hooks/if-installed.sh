#!/bin/sh
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
