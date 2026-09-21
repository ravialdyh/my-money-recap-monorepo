#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
fail() { echo "FAIL: $*" >&2; exit 1; }

age-keygen -o "$work/key.txt" 2>/dev/null
recipient=$(age-keygen -y "$work/key.txt")
echo "throwaway recipient: $recipient"

example=$work/example.yaml
cp ../secrets/example.yaml "$example"
sed 's/replace-me/filled-by-the-smoke/g' "$example" >"$work/filled.yaml"

dir=$work
encrypt() { sops encrypt --age "$recipient" --input-type yaml --output-type yaml "$1" >"$dir/prod.yaml"; }
count() { grep -cE "$1" <<<"$out" || true; }
run() {
  status=0
  out=$(WIZARD_SECRETS_DIR=$dir SOPS_AGE_KEY_FILE=${1:-$work/key.txt} ./setup.sh --check 2>&1) || status=$?
}
walk() {
  status=0
  out=$(WIZARD_SECRETS_DIR=$1 WIZARD_RULES_FILE=$1/.sops.yaml TMPDIR=$1 \
    SOPS_AGE_KEY_FILE=$2 ./setup.sh <"$3" 2>&1) || status=$?
}

run
[ "$status" -ne 0 ] || fail "check mode passed with no secrets file"
[ "$(count '^ *[0-9]+  missing  ')" = 8 ] || fail "an absent secrets file left $(count '  missing  ') of 8 missing"
grep -qE '^ 2  by hand  Apple Developer Program' <<<"$out" || fail "the Apple step is not reported by hand"
[ "$(count '^ *[0-9]+  (done|missing|by hand)  ')" = 9 ] || fail "check mode reported other than nine steps"
echo "absent secrets file: all eight checkable steps missing, Apple by hand, exit $status"

encrypt "$work/filled.yaml"
run /dev/null
[ "$status" -ne 0 ] || fail "check mode passed without the private half of the age key"
grep -qE '^ 1  missing  ' <<<"$out" || fail "the age step passed without the private half"
echo "no private half: the decrypt attempt fails and step 1 is missing, exit $status"

run
[ "$status" -eq 0 ] || fail "check mode failed with every value filled:"$'\n'"$out"
grep -qE '^ 1  done  ' <<<"$out" || fail "the age step is not done with the private half"
[ "$(count '^ *[0-9]+  done  ')" = 8 ] || fail "every value filled left a step missing"
echo "private half and every value filled: eight done, exit $status"

unfilled=$(../secrets/keyset.sh value "$example" notify_database_url)
sed "s|^notify_database_url: .*|notify_database_url: $unfilled|" "$work/filled.yaml" >"$work/stack.yaml"
encrypt "$work/stack.yaml"
run
[ "$status" -ne 0 ] || fail "check mode passed with a stack value still at its placeholder"
grep -qE '^ -  missing  .* notify_database_url' <<<"$out" || fail "the unfilled stack value was not named"
echo "stack value at its placeholder: named and caught, exit $status"

grep -v '^smtp_token:' "$work/filled.yaml" >"$work/short.yaml"
encrypt "$work/short.yaml"
run
[ "$status" -ne 0 ] || fail "a key the example file declares and the secrets file lacks passed"
grep -q "smtp_token absent from $work/prod.yaml" <<<"$out" || fail "the key absent from the secrets file was not named"
echo "key in the example file only: named and caught, exit $status"

encrypt "$work/filled.yaml"
echo 'mailpit_token: replace-me' >>"$example"
run
[ "$status" -ne 0 ] || fail "a key no step produces passed"
grep -q 'drift: .* declares mailpit_token' <<<"$out" || fail "the key no step produces was not named"
echo "key no step produces: named and caught, exit $status"

held=$work/held
mkdir -p "$held"
cp ../secrets/example.yaml "$held/example.yaml"
walk "$held" "$work/key.txt" /dev/null
[ "$status" -eq 0 ] || fail "the walk failed with an age key already on disk:"$'\n'"$out"
grep -q "$recipient" "$held/.sops.yaml" || fail "step 1 wrote no creation rule for the key already on disk"
[ -f "$held/prod.yaml" ] || fail "step 1 created no secrets file for the key already on disk"
! grep -q 'does not exist yet' <<<"$out" || fail "a later step still dead-ended on a missing secrets file"
[ ! -e ../../.sops.yaml ] && [ ! -e ../secrets/prod.yaml ] || fail "the walk wrote into the repository tree"
dir=$held
run
grep -qE '^ 1  done  ' <<<"$out" || fail "step 1 is not done after a walk that reused the key on disk"
grep -qE '^ 3  missing  .* app_store_connect_api_key' <<<"$out" || fail "step 3 is not reported unfilled"
echo "age key already on disk: step 1 still writes both files, and check mode reads them"

bare=$work/bare
mkdir -p "$bare"
cp ../secrets/example.yaml "$bare/example.yaml"
printf 'n\ny\ny\ny\ny\ny\ny\ny\ny\n' >"$work/answers"
walk "$bare" "$bare/no-key.txt" "$work/answers"
[ ! -e "$bare/mmr-deploy-key" ] || fail "step 9 wrote a private key with no secrets file to paste it into"
grep -q "$bare/prod.yaml does not exist yet" <<<"$out" || fail "step 9 did not report the missing secrets file"
echo "no secrets file: step 9 leaves no private key on disk"

! grep -nE "sed -n .s/\^" setup.sh ../secrets/decrypt-smoke.sh ||
  fail "example.yaml is parsed outside ../secrets/keyset.sh"
echo "one home: neither script parses example.yaml on its own"

echo "wizard check mode OK on $(sops --version --disable-version-check) and $(age-keygen --version)"
