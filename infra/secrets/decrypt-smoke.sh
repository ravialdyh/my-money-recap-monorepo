#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
fail() { echo "FAIL: $*" >&2; exit 1; }
mode() { stat -c '%a' "$1" 2>/dev/null || stat -f '%Lp' "$1"; }
expected() { sed -n "s/^$1: //p" example.yaml; }

age-keygen -o "$work/key.txt" 2>/dev/null
recipient=$(age-keygen -y "$work/key.txt")
echo "throwaway recipient: $recipient"

sops encrypt --age "$recipient" --input-type yaml --output-type yaml example.yaml >"$work/prod.yaml"
grep -q '^sops:' "$work/prod.yaml" || fail "sops wrote no metadata"
! grep -q 'replace-me' "$work/prod.yaml" || fail "a placeholder survived encryption as plain text"
echo "encrypted: $(wc -l <"$work/prod.yaml" | tr -d ' ') lines of ciphertext and metadata"

SOPS_AGE_KEY_FILE=$work/key.txt ./decrypt.sh "$work/prod.yaml" "$work/out"

keys=$(sed -n 's/^\([a-z0-9_]*\):.*/\1/p' example.yaml)
for key in $keys; do
  file=$work/out/$key
  [ -f "$file" ] || fail "$key has no file"
  [ "$(mode "$file")" = 400 ] || fail "$key is mode $(mode "$file"), want 400"
  [ "$(cat "$file")" = "$(expected "$key")" ] || fail "$key holds $(cat "$file")"
  [ "$(tail -c1 "$file" | wc -l)" -eq 0 ] || fail "$key ends in a newline the value never had"
done

want=$(wc -w <<<"$keys" | tr -d " ")
got=$(find "$work/out" -type f | wc -l | tr -d " ")
[ "$want" -eq "$got" ] || fail "$want keys made $got files"
echo "$want values, each its own file, mode 0400, contents intact"
echo "secrets OK on $(sops --version --disable-version-check) and $(age-keygen --version)"
