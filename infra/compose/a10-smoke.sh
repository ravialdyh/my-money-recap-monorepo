#!/usr/bin/env bash
# A10 smoke test (issue #14): on the pinned Phase Two image, the bff service account issues a Login Token
# for the test user, the ios client redeems it at the token endpoint without a browser, and a second
# redemption is refused with invalid_grant. Also checks the username-password grant still works on ios.
# Needs the Compose stack up (see README.md) and the same .env; prints the token's subject and the image tag.
set -euo pipefail
cd "$(dirname "$0")"
set -a; . ./.env; set +a

KC=${KEYCLOAK_URL:-http://localhost:${KEYCLOAK_HOST_PORT:-8080}}
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
[[ $expected != null ]] || fail "test user $EMAIL not found"

lt=$(curl -sfS -X POST "$REALM/login-token" -H "Authorization: Bearer $sa" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"client_id\":\"ios\",\"expiration_seconds\":900,\"reusable\":false,\"force_create\":true,\"set_email_verified\":true}" \
  | jq -r .login_hint)
[[ $lt == lt:* ]] || fail "login-token endpoint returned '$lt'"
echo "issued: ${lt:0:11}…"

redeem() { curl -s -o "$1" -w '%{http_code}' -X POST "$TOKEN" -d grant_type=password -d client_id=ios -d scope=login-token -d "login_token=$lt"; }
body=$(mktemp)
status=$(redeem "$body")
[[ $status == 200 ]] || fail "first redemption: HTTP $status $(cat "$body")"
sub=$(claim sub <"$body")
[[ $sub == "$expected" ]] || fail "subject $sub is not the test user $expected"
echo "redeemed: sub=$sub username=$(claim preferred_username <"$body") aud=$(claim aud <"$body") email_verified=$(claim email_verified <"$body")"

status=$(redeem "$body")
[[ $status == 401 && $(jq -r .error "$body") == invalid_grant ]] || fail "second redemption: HTTP $status $(cat "$body")"
echo "second redemption refused: HTTP $status $(cat "$body")"

status=$(curl -s -o "$body" -w '%{http_code}' -X POST "$TOKEN" -d grant_type=password -d client_id=ios \
  -d username=test -d "password=$TEST_USER_PASSWORD")
[[ $status == 200 && $(claim sub <"$body") == "$expected" ]] || fail "password grant on ios: HTTP $status $(cat "$body")"
echo "password grant on ios still works"
rm -f "$body"
echo "A10 OK on $image"
