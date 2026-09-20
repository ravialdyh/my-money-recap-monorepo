#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

status=0
docker compose run --rm realm-apply plan -input=false -detailed-exitcode || status=$?
case $status in
  0) echo "no-op: the next up re-creates nothing" ;;
  2) echo "FAIL: the next up would change the realm" >&2 ;;
  *) echo "FAIL: tofu plan exited $status" >&2 ;;
esac
[ "$status" = 0 ]
