#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $(basename "$0") keys <example.yaml>" >&2
  echo "       $(basename "$0") value <example.yaml> <key>" >&2
  exit 2
}

[ $# -ge 2 ] || usage
file=$2

case $1 in
  keys)
    [ $# -eq 2 ] || usage
    sed -n 's/^\([a-z0-9_]*\):.*/\1/p' "$file"
    ;;
  value)
    [ $# -eq 3 ] || usage
    sed -n "s/^$3: //p" "$file"
    ;;
  *) usage ;;
esac
