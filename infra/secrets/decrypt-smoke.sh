#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
fail() { echo "FAIL: $*" >&2; exit 1; }
mode() { stat -c '%a' "$1" 2>/dev/null || stat -f '%Lp' "$1"; }
expected() { ./keyset.sh value example.yaml "$1"; }

age-keygen -o "$work/key.txt" 2>/dev/null
recipient=$(age-keygen -y "$work/key.txt")
echo "throwaway recipient: $recipient"

sops encrypt --age "$recipient" --input-type yaml --output-type yaml example.yaml >"$work/prod.yaml"
grep -q '^sops:' "$work/prod.yaml" || fail "sops wrote no metadata"
! grep -q 'replace-me' "$work/prod.yaml" || fail "a placeholder survived encryption as plain text"
echo "encrypted: $(wc -l <"$work/prod.yaml" | tr -d ' ') lines of ciphertext and metadata"

SOPS_AGE_KEY_FILE=$work/key.txt ./decrypt.sh "$work/prod.yaml" "$work/out"

keys=$(./keyset.sh keys example.yaml)
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

env=$work/machine.env
../compose/machine-env.sh "$work/out" >"$env"
stack=$(./keyset.sh stack)
for key in $stack; do
  grep -qxF "$(tr '[:lower:]' '[:upper:]' <<<"$key")='$(expected "$key")'" "$env" || fail "the machine's .env lost $key"
done
ids=$(grep -oE '\$\{[A-Z_]+:-replace-me\}' ../compose/compose.yaml | sed -E 's/^\$\{([A-Z_]+):.*/\1/' | sort -u)
[ -n "$ids" ] || fail "no replace-me default found in compose.yaml"
for name in $ids; do
  grep -qxF "$name='$(expected "$(tr '[:upper:]' '[:lower:]' <<<"$name")")'" "$env" || fail "the machine's .env leaves $name at replace-me"
done
grep -qxF "OPEN_EXCHANGE_RATES_APP_ID='$(expected open_exchange_rates_app_id)'" "$env" || fail "the machine's .env lost the rates App ID"
grep -qxF "REALM_SSL_REQUIRED=external" "$env" || fail "the machine's .env leaves ssl_required at the laptop's none"
! grep -qE '^(TEST_USER_PASSWORD|DEPLOY_SSH_PRIVATE_KEY)=' "$env" || fail "the machine's .env carries a test password or the deploy key"
cp -R "$work/out" "$work/quoted" && chmod u+w "$work/quoted/bff_client_secret"
printf "it's" >"$work/quoted/bff_client_secret"
! ../compose/machine-env.sh "$work/quoted" >/dev/null 2>&1 || fail "a value with a quote reached the machine's .env"
echo "machine .env: $(wc -l <"$env" | tr -d ' ') lines, all $(wc -w <<<"$stack" | tr -d ' ') stack values and $(wc -w <<<"$ids" | tr -d ' ') provider ids intact, no test Member password, a quote refused"
echo "secrets OK on $(sops --version --disable-version-check) and $(age-keygen --version)"
