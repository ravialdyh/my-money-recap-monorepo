#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
: "${MMR_TEST_USER_PASSWORD:=$(docker inspect mmr-keycloak-1 --format '{{range .Config.Env}}{{println .}}{{end}}' | sed -n 's/^TEST_USER_PASSWORD=//p')}"
[ -n "$MMR_TEST_USER_PASSWORD" ] || { echo "MMR_TEST_USER_PASSWORD is empty and no mmr-keycloak-1 container is running" >&2; exit 1; }

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
  TEST_RUNNER_MMR_TEST_USER_PASSWORD="$MMR_TEST_USER_PASSWORD" xcodebuild test-without-building \
    -project MyMoneyRecap.xcodeproj -scheme MyMoneyRecap -destination "id=$udid" -derivedDataPath build \
    -only-testing:MyMoneyRecapUITests -resultBundlePath "build/ui-$which.xcresult" -quiet
  xcrun xcresulttool export attachments --path "build/ui-$which.xcresult" --output-path "build/ui-$which-screenshots" >/dev/null
  echo "screenshots: build/ui-$which-screenshots"
done
