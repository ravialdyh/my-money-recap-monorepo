#!/bin/sh
set -eu
cd "$(git rev-parse --show-toplevel)"
if ! command -v tofu >/dev/null 2>&1; then
  printf 'pre-commit: skipping tofu %s, not installed locally (CI runs it)\n' "$1" >&2
  exit 0
fi
for root in infra/tofu/*/; do
  (cd "$root" && tofu init -backend=false -input=false >/dev/null && tofu "$@")
done
