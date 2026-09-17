# apps/ios

The SwiftUI app (`docs/IOS.md`), iOS 17+, iPhone and iPad. This slice is Sign in (PRD §6.1) through
Keycloak's login page with AppAuth-iOS (I1), the session in the Keychain (I2) and a signed-in
placeholder; Home replaces the placeholder in its own ticket.

## Open

```bash
open apps/ios/MyMoneyRecap.xcodeproj
```

Xcode resolves AppAuth-iOS (pinned to 3.0.0 in the project; `Package.resolved` is committed). No
signing team is needed for the simulator, but the build must stay signed ("Sign to Run Locally"):
with `CODE_SIGNING_ALLOWED=NO` the app has no application-identifier entitlement and every Keychain
call fails with `errSecMissingEntitlement`, so a relaunch shows Sign in again.

## Run against local Compose

1. Bring the stack up (`infra/compose/README.md`): Keycloak on `http://localhost:8080` with the `mmr`
   realm, its public `ios` client (PKCE S256, redirect `mymoneyrecap://auth/callback`) and the test
   user `test`, password `TEST_USER_PASSWORD` from `infra/compose/.env`.
2. Pick any iPhone simulator in Xcode and Run. The simulator shares the Mac's `localhost`, and the
   app allows plain HTTP to `localhost` only (`App/Info.plist`).
3. Tap **Continue with email**, allow the system web sheet, sign in as the test user. Relaunching
   skips Sign in: the `OIDAuthState` is stored `kSecAttrAccessibleWhenUnlockedThisDeviceOnly`. Cancel
   the sheet to see the `InlineError` inside the card. Sign in with Apple and Continue with Google are
   drawn only until slice 3.

Launch argument `--reset-session` erases the stored session, which is how the UI test starts clean.

## Build and test from the command line

The same two commands CI runs. Nothing names a simulator device or OS version (issue #16, "Device
and OS coverage"): the build uses the generic simulator destination and `scripts/simulator-udid.sh`
picks a booted or available iPhone simulator through `simctl`.

```bash
cd apps/ios
xcodebuild build-for-testing -project MyMoneyRecap.xcodeproj -scheme MyMoneyRecap -destination 'generic/platform=iOS Simulator' -derivedDataPath build
xcodebuild test-without-building -project MyMoneyRecap.xcodeproj -scheme MyMoneyRecap -destination "id=$(scripts/simulator-udid.sh)" -derivedDataPath build -only-testing:MyMoneyRecapTests
```

The UI test (`UITests/`, the sign-in-to-Home path of I9) drives the real web sheet against local
Compose and skips itself unless the test user's password is in the runner's environment:

```bash
TEST_RUNNER_MMR_TEST_USER_PASSWORD="$(grep '^TEST_USER_PASSWORD=' ../../infra/compose/.env | cut -d= -f2-)" xcodebuild test-without-building -project MyMoneyRecap.xcodeproj -scheme MyMoneyRecap -destination "id=$(scripts/simulator-udid.sh)" -derivedDataPath build -only-testing:MyMoneyRecapUITests -resultBundlePath build/ui.xcresult
xcrun xcresulttool export attachments --path build/ui.xcresult --output-path build/ui-screenshots
```

## Layout

| Folder | What lives here |
| --- | --- |
| `App/` | Entry point, root switch between Sign in and the signed-in placeholder, `Info.plist` (URL scheme, ATS exception for `localhost`). |
| `Auth/` | `AuthConfig` (issuer, client id, redirect; no secret exists in the target), `AuthSession` (AppAuth discovery, PKCE authorization, `OIDAuthState`), `Keychain`. |
| `Screens/` | One folder per SCREEN-STATES group: `SignIn/` (view + `@Observable` view model), `Home/` placeholder. |
| `DesignSystem/` | Generated `Tokens.swift` (do not edit; `npm run build:swift-tokens`) and the SwiftUI components mirrored from `my-money-recap-ds/src/components`: `Screen`, `Headline`, `SubCopy`, `Illustration`, `Card`, `MMRButton`, `InlineError`. Colours, radii and type come only from the tokens. |
| `Resources/` | Asset catalog with the cover illustration (PNG from `my-money-recap/assets/illustrations/ill-cover.webp`). |
| `Tests/`, `UITests/` | XCTest unit tests on the view model; the one UI test above. |
| `scripts/` | `simulator-udid.sh`. |

Not yet here: `Networking/` and `Snapshots/` (I3, I4) arrive with the BFF client; the Fredoka and
Nunito font files are not bundled, so the type ramp renders in the system font until they are.
