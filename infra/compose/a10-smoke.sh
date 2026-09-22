#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
set -a
# shellcheck disable=SC1091 # .env is written from .env.example at run time, never committed.
. ./.env
set +a

KC=http://localhost:${KEYCLOAK_HOST_PORT:-8080}
REALM=$KC/realms/mmr
TOKEN=$REALM/protocol/openid-connect/token
EMAIL=test@example.com
fail() { echo "FAIL: $*" >&2; exit 1; }
claim() { jq -r ".access_token | split(\".\")[1] | @base64d | fromjson | .$1"; }

image=$(docker compose config --images | grep phasetwo-keycloak)
echo "image: $image"

sa=$(curl -sfS -X POST "$TOKEN" -d grant_type=client_credentials -d client_id=bff \
  -d "client_secret=$BFF_CLIENT_SECRET" | jq -r .access_token)
expected=$(curl -sfS "$KC/admin/realms/mmr/users?email=$EMAIL&exact=true" -H "Authorization: Bearer $sa" | jq -r '.[0].id')
[[ $expected != null ]] || fail "test Member $EMAIL not found"

lt=$(curl -sfS -X POST "$REALM/login-token" -H "Authorization: Bearer $sa" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"client_id\":\"ios\",\"expiration_seconds\":900,\"reusable\":false,\"force_create\":true,\"set_email_verified\":true}" \
  | jq -r .login_hint)
[[ $lt == lt:* ]] || fail "login-token endpoint returned '$lt'"
echo "issued: ${lt:0:11}…"

redeem() { curl -s -o "$1" -w '%{http_code}' -X POST "$TOKEN" -d grant_type=password -d client_id=ios --data-urlencode 'scope=login-token offline_access' -d "login_token=$lt"; }
body=$(mktemp)
status=$(redeem "$body")
[[ $status == 200 ]] || fail "first redemption: HTTP $status $(cat "$body")"
sub=$(claim sub <"$body")
[[ $sub == "$expected" ]] || fail "subject $sub is not the test Member $expected"
echo "redeemed: sub=$sub username=$(claim preferred_username <"$body") aud=$(claim aud <"$body") email_verified=$(claim email_verified <"$body")"
kind=$(jq -r '.refresh_token | split(".")[1] | @base64d | fromjson | .typ' <"$body")
[[ $kind == Offline ]] || fail "the redeemed sign-in is a $kind one, not the offline one PRD D62's 30 days needs"
echo "sign-in kind: $kind (30 days from last use)"
offline=$(jq -r .refresh_token <"$body")

status=$(redeem "$body")
[[ $status == 401 && $(jq -r .error "$body") == invalid_grant ]] || fail "second redemption: HTTP $status $(jq -r '.error // "new tokens issued"' "$body")"
echo "second redemption refused: HTTP $status $(cat "$body")"

status=$(curl -s -o "$body" -w '%{http_code}' -X POST "$TOKEN" -d grant_type=password -d client_id=ios \
  -d username=test -d "password=$TEST_USER_PASSWORD")
[[ $status == 200 && $(claim sub <"$body") == "$expected" ]] || fail "password grant on ios: HTTP $status $(jq -r '.error // "a token for another subject"' "$body")"
echo "password grant on ios still works"
first=$(curl -sfS -X POST "$TOKEN" -d grant_type=password -d client_id=ios -d username=test \
  -d "password=$TEST_USER_PASSWORD" --data-urlencode 'scope=openid offline_access' | jq -r .refresh_token)
curl -sfS -o /dev/null -X POST "$TOKEN" -d grant_type=refresh_token -d client_id=ios --data-urlencode "refresh_token=$first"
status=$(curl -s -o "$body" -w '%{http_code}' -X POST "$TOKEN" -d grant_type=refresh_token -d client_id=ios --data-urlencode "refresh_token=$first")
[[ $status == 400 && $(jq -r .error "$body") == invalid_grant ]] || fail "a used refresh token still renews: HTTP $status $(jq -r '.error // "new tokens issued"' "$body")"
echo "a used refresh token is refused: $(jq -r .error_description "$body")"
curl -sfS -o /dev/null -X POST "$REALM/protocol/openid-connect/logout" -d client_id=ios --data-urlencode "refresh_token=$offline"
status=$(curl -s -o "$body" -w '%{http_code}' -X POST "$TOKEN" -d grant_type=refresh_token -d client_id=ios --data-urlencode "refresh_token=$offline")
[[ $status == 400 && $(jq -r .error "$body") == invalid_grant ]] || fail "logout left the offline sign-in alive: HTTP $status $(jq -r '.error // "new tokens issued"' "$body")"
echo "logout ended the offline sign-in"
rm -f "$body"
echo "A10 OK on $image"
