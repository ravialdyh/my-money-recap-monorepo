# Running and testing the iOS app

How to run the app against the local Compose stack, on a real iPhone, and under test. What the app is and does
is in the [README](README.md).

## Run against local Compose

1. Start the stack with `--profile mail` ([compose README, "Run it"](../../infra/compose/README.md#run-it)).
2. Pick any iPhone simulator in Xcode and Run.
3. Tap **Continue with email**, type an address, and open the message in Mailpit at `http://localhost:8025`.
4. Hand the link's `lt` value to the app:

```bash
xcrun simctl openurl booted "mymoneyrecap://magic?lt=<value>"
```

Launch arguments: `--reset-session`, `--reset-snapshot`, `--magic-link <link>`, and in a debug build
`--authorization-code-flow`, which opens Keycloak's own login page. `MMR_ISSUER` and `MMR_BFF_URL` point the app
at another Keycloak and BFF.

## Run on a physical iPhone

A real iPhone cannot reach the Mac as `localhost`.

1. Set `MMR_HOST=<the Mac's LAN IP>` in `infra/compose/.env`, then run the `up` command again.
2. In Xcode, edit the scheme's Run action and add `MMR_ISSUER=http://<ip>:8080/realms/mmr` and
   `MMR_BFF_URL=http://<ip>:8082`.
3. Sign in by email. In the phone's Safari, type `mymoneyrecap://magic?lt=<value>`, with the `lt` value from the
   Mailpit message.
4. Set `MMR_HOST` back to `localhost` and run `up` again before using the simulator.

## Test it

From `apps/ios`, the same build and unit tests CI runs:

```bash
xcodebuild build-for-testing -project MyMoneyRecap.xcodeproj -scheme MyMoneyRecap -destination 'generic/platform=iOS Simulator' -derivedDataPath build -enableCodeCoverage YES -quiet
```

```bash
xcodebuild test-without-building -project MyMoneyRecap.xcodeproj -scheme MyMoneyRecap -destination "id=$(scripts/simulator-udid.sh)" -derivedDataPath build -only-testing:MyMoneyRecapTests -enableCodeCoverage YES -resultBundlePath build/MyMoneyRecap.xcresult
```

Delete `build/MyMoneyRecap.xcresult` before running it again.

The UI tests need the Compose stack. They sign in by email link and through Keycloak's page, then walk setup,
Home, Spends, the Recap, Settings, log out and delete account:

```bash
scripts/ui-test.sh
```

They run on the newest and the oldest installed iOS runtime; `scripts/ui-test.sh newest` or `oldest` runs one.
The app needs iOS 26.0 or later, the oldest runtime Xcode 27 installs:

```bash
xcodebuild -downloadPlatform iOS -buildVersion 26.0
```
