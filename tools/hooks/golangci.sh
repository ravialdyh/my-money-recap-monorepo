#!/bin/sh
# golangci-lint must run inside the module; the config lives at the repo root.
set -eu
root="$(git rev-parse --show-toplevel)"
if ! command -v golangci-lint >/dev/null 2>&1; then
  printf 'pre-commit: skipping golangci-lint, not installed locally (CI runs it)\n' >&2
  exit 0
fi
cd "$root/apps/backend"
exec golangci-lint run --fix --config "$root/.golangci.yml" ./...
