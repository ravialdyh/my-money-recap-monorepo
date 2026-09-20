#!/usr/bin/env bash
set -euo pipefail

[ $# -eq 2 ] || { echo "usage: $(basename "$0") <encrypted.yaml> <directory>" >&2; exit 2; }
encrypted=$1
dir=$2

plain=$(sops decrypt --output-type json "$encrypted")
umask 077
mkdir -p "$dir"

while IFS= read -r key; do
  case $key in
    [a-z][a-z0-9_]*) ;;
    *) echo "refusing '$key': a key names a file, so it is lower-case, digits and underscores" >&2; exit 1 ;;
  esac
  jq -ej --arg k "$key" '.[$k] | select(type == "string")' <<<"$plain" >"$dir/$key" ||
    { echo "$key is not a string; the secrets file is one flat map of scalars" >&2; exit 1; }
  chmod 0400 "$dir/$key"
done < <(jq -r 'keys_unsorted[]' <<<"$plain")

echo "decrypted $(jq -r 'keys | length' <<<"$plain") values into $dir"
