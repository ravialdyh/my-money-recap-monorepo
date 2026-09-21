#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
env_file=../../infra/compose/.env
from_env() {
  value="$([ -f "$env_file" ] && sed -n "s/^$1=//p" "$env_file" | tail -n 1 || true)"
  echo "${value:-$2}"
}
project="$(from_env COMPOSE_PROJECT_NAME mmr)"
: "${MMR_ISSUER:=http://localhost:$(from_env KEYCLOAK_HOST_PORT 8080)/realms/mmr}"
: "${MMR_BFF_URL:=http://localhost:$(from_env BFF_HOST_PORT 8082)}"
: "${MMR_MAILPIT_URL:=http://localhost:$(from_env MAILPIT_HOST_PORT 8025)}"
: "${MMR_TEST_USER_PASSWORD:=$(docker inspect "$project-realm-apply-1" --format '{{range .Config.Env}}{{println .}}{{end}}' | sed -n 's/^TF_VAR_test_user_password=//p')}"
[ -n "$MMR_TEST_USER_PASSWORD" ] || { echo "MMR_TEST_USER_PASSWORD is empty and no $project-realm-apply-1 container exists" >&2; exit 1; }
echo "stack $project: issuer $MMR_ISSUER, BFF $MMR_BFF_URL, Mailpit $MMR_MAILPIT_URL"

xcodebuild build-for-testing -project MyMoneyRecap.xcodeproj -scheme MyMoneyRecap \
  -destination 'generic/platform=iOS Simulator' -derivedDataPath build -quiet

seen=""
for which in newest oldest; do
  udid="$(scripts/simulator-udid.sh "$which")"
  case " $seen " in *" $udid "*) echo "only one iOS runtime is installed; $which is the same simulator"; continue;; esac
  seen="$seen $udid"
  xcrun simctl boot "$udid" 2>/dev/null || true
  xcrun simctl bootstatus "$udid" -b >/dev/null
  echo "== UI test on $(xcrun simctl list devices -j | python3 -c "import json,sys; print(next(r.rsplit('.',1)[-1] for r,ds in json.load(sys.stdin)['devices'].items() for d in ds if d['udid']=='$udid'))") ($which)"
  rm -rf "build/ui-$which.xcresult" "build/ui-$which-screenshots"
  TEST_RUNNER_MMR_TEST_USER_PASSWORD="$MMR_TEST_USER_PASSWORD" TEST_RUNNER_MMR_ISSUER="$MMR_ISSUER" \
    TEST_RUNNER_MMR_BFF_URL="$MMR_BFF_URL" TEST_RUNNER_MMR_MAILPIT_URL="$MMR_MAILPIT_URL" \
    xcodebuild test-without-building \
    -project MyMoneyRecap.xcodeproj -scheme MyMoneyRecap -destination "id=$udid" -derivedDataPath build \
    -only-testing:MyMoneyRecapUITests -resultBundlePath "build/ui-$which.xcresult" -quiet
  xcrun xcresulttool export attachments --path "build/ui-$which.xcresult" --output-path "build/ui-$which-screenshots" >/dev/null
  echo "screenshots: build/ui-$which-screenshots"
done
