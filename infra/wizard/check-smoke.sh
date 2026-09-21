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

encrypt() { sops encrypt --age "$recipient" --input-type yaml --output-type yaml "$1" >"$work/prod.yaml"; }
count() { grep -cE "$1" <<<"$out" || true; }
run() {
  status=0
  out=$(WIZARD_SECRETS_DIR=$work SOPS_AGE_KEY_FILE=${1:-$work/key.txt} ./setup.sh --check 2>&1) || status=$?
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

echo "wizard check mode OK on $(sops --version --disable-version-check) and $(age-keygen --version)"
