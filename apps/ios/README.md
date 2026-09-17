# apps/ios

The SwiftUI app (`docs/IOS.md`), iOS 17+, iPhone and iPad. So far: Sign in (PRD §6.1) through
Keycloak's login page with AppAuth-iOS (I1), the session in the Keychain (I2), the generated Connect
client calling `GetHome` (I3), the `HomeSnapshot` cache in SwiftData (I4) and Home drawn at zero in the
B1 "no salary set" layout (PRD §6.8, D54, D69). Nothing on Home navigates yet; the sheets and screens
behind the chips, cards and buttons arrive with their own tickets.

## Open

```bash
open apps/ios/MyMoneyRecap.xcodeproj
```

Xcode resolves AppAuth-iOS (pinned to 3.0.0), the local contract package `MoneyRecapContract`
(`packages/contract/swift`, the `connect-swift` client `buf generate` writes) and, at the same exact
versions that package pins, `connect-swift` 1.2.3 and `swift-protobuf` 1.38.1, which the app links
directly for the interceptor and the snapshot bytes. `Package.resolved` is committed. No signing team
is needed for the simulator, but the build must stay signed ("Sign to Run Locally"): with
`CODE_SIGNING_ALLOWED=NO` the app has no application-identifier entitlement and every Keychain call
fails with `errSecMissingEntitlement`, so a relaunch shows Sign in again.

## Run against local Compose

1. Bring the stack up (`infra/compose/README.md`): Keycloak on `http://localhost:8080` with the `mmr`
   realm, its public `ios` client (PKCE S256, redirect `mymoneyrecap://auth/callback`) and the test
   user `test`, password `TEST_USER_PASSWORD` from `infra/compose/.env`; the BFF on
   `http://localhost:8082` (`BFF_HOST_PORT`), which is what the app calls.
2. Pick any iPhone simulator in Xcode and Run. The simulator shares the Mac's `localhost`, and the
   app allows plain HTTP on the local network only (`NSAllowsLocalNetworking` in `App/Info.plist`).
3. Tap **Continue with email**, allow the system web sheet, sign in as the test user. Home loads:
   the first time the §6.19 skeleton (two cards at their filled heights, no spinner) until `GetHome`
   answers, then "Spent so far" at zero, `Money in` at zero and no Saved chip (D41), the two
   `PromptCard`s, `Log a spend`, `See last month` disabled and the dashed preview card. Relaunching
   skips Sign in and draws Home from the stored `HomeSnapshot` before the network answers; stop
   Compose and it still draws the last snapshot. Cancel the sheet to see the `InlineError` inside the
   card. Sign in with Apple and Continue with Google are drawn only until slice 3.

Seams, all read at launch: argument `--reset-session` erases the stored session, `--reset-snapshot`
empties the SwiftData store, and the environment variables `MMR_ISSUER` (default
`http://localhost:8080/realms/mmr`) and `MMR_BFF_URL` (default `http://localhost:8082`) point the app at
another Keycloak and BFF. The UI test uses the last one with an unreachable port to prove Home still
renders offline.

## Run on a physical iPhone

The phone reaches the Mac by its LAN address, and the token issuer must be that same address. The three
steps (the `MMR_HOST` variable in `.env`, the `MMR_ISSUER` and `MMR_BFF_URL` Run-action variables, and
setting it back) are in `infra/compose/README.md`, "A phone on the LAN". Sign in as the test user; Home at
zero appears, and a relaunch draws it from the snapshot. App Transport Security in `App/Info.plist` permits
local-network HTTP for this development stack only, until slice 3 terminates TLS in Nginx; nothing needs to
be edited for a device build.

## Build and test from the command line

The same two commands CI runs. Nothing names a simulator device or OS version (the "Device and OS
coverage" section of issues #16 and #18): the build uses the generic simulator destination and
`scripts/simulator-udid.sh` picks a booted or available iPhone simulator through `simctl`.

```bash
cd apps/ios
xcodebuild build-for-testing -project MyMoneyRecap.xcodeproj -scheme MyMoneyRecap -destination 'generic/platform=iOS Simulator' -derivedDataPath build
xcodebuild test-without-building -project MyMoneyRecap.xcodeproj -scheme MyMoneyRecap -destination "id=$(scripts/simulator-udid.sh)" -derivedDataPath build -only-testing:MyMoneyRecapTests
```

## The UI test against Compose

`UITests/SignInUITests.swift` is the one UI test of I9: it cancels the web sheet, signs in as the test
user through Keycloak's login page, asserts Home at zero, relaunches to prove Home renders from the
snapshot, relaunches with `MMR_BFF_URL` pointing at a closed port to prove the snapshot survives an
unreachable BFF, and relaunches with `--reset-snapshot` to prove the first load is the skeleton. It
skips itself unless the test user's password is in the runner's environment, so CI's `ios` job runs
only the unit tests.

The same section asks for two simulator runtimes, both chosen when the script runs and never written
down: the newest installed iOS runtime and the oldest (the deployment target's floor, iOS 17, when Xcode
offers that runtime; on a Mac whose Xcode no longer lists it, the oldest it does; Xcode 27 no longer
offers iOS 17 or 18 for download, so on such a Mac the floor is iOS 26.0).
`scripts/ui-test.sh` builds once, then runs the UI test on each, taking the password from
`MMR_TEST_USER_PASSWORD` or, failing that, from the running `mmr-keycloak-1` container:

```bash
scripts/ui-test.sh
```

Result bundles land in `build/ui-newest.xcresult` and `build/ui-oldest.xcresult`, with the screenshots
exported next to them. `scripts/simulator-udid.sh newest|oldest` is the selection on its own (it
creates an iPhone on a runtime that has none), and with no argument it keeps the booted-or-first
behaviour the unit-test command uses. To run one destination by hand:

```bash
TEST_RUNNER_MMR_TEST_USER_PASSWORD="$(docker inspect mmr-keycloak-1 --format '{{range .Config.Env}}{{println .}}{{end}}' | sed -n 's/^TEST_USER_PASSWORD=//p')" xcodebuild test-without-building -project MyMoneyRecap.xcodeproj -scheme MyMoneyRecap -destination "id=$(scripts/simulator-udid.sh oldest)" -derivedDataPath build -only-testing:MyMoneyRecapUITests -resultBundlePath build/ui.xcresult
xcrun xcresulttool export attachments --path build/ui.xcresult --output-path build/ui-screenshots
```

Install an older runtime through Xcode > Settings > Components (or `xcodebuild -downloadPlatform iOS
-buildVersion 17.5`); the script picks it up on the next run without any change.

## Layout

| Folder | What lives here |
| --- | --- |
| `App/` | Entry point, the SwiftData container, root switch between Sign in and Home (sign-out deletes the cache), `Info.plist` (URL scheme, the ATS local-networking allowance). |
| `Auth/` | `AuthConfig` (issuer, client id, redirect; no secret exists in the target), `AuthSession` (AppAuth discovery, PKCE authorization, `OIDAuthState`, fresh tokens and the refresh-or-sign-out on `unauthenticated`), `Keychain`. |
| `Networking/` | `BffClient` builds the generated `MoneyRecap_V1_BffServiceClient` (proto codec, `MMR_BFF_URL` or its default); `BearerInterceptor` is the one interceptor of I3 so far. No request or response type is written here. |
| `Snapshots/` | `HomeSnapshot`, the one SwiftData model of I4: the `HomeView` bytes plus `updatedAt`, replaced whole on every success. |
| `Screens/` | One folder per SCREEN-STATES group: `SignIn/` and `Home/`, each a view plus an `@Observable` view model (I6). |
| `DesignSystem/` | Generated `Tokens.swift` (do not edit; `npm run build:swift-tokens`), `Money` (the platform currency formatter, I7) and the SwiftUI components mirrored one file per name from `my-money-recap-ds/src/components`: `Screen`, `Headline` (+ `SubCopy`, `Eyebrow`, `HelperText`), `Illustration`, `Avatar`, `Card` (+ `CardTitle`), `Chip`, `StatHero`, `StatChip`, `PromptCard`, `ListRow`, `MMRButton`, `InlineError`. Colours, radii and type come only from the tokens. The React row wrappers (`ChipRow`, `StatChipRow`) are a plain `HStack` here, and `PromptCard` has a `dashed` variant (the D47 outline card) the React one draws by hand. |
| `Resources/` | Asset catalog with the cover illustration (PNG from `my-money-recap/assets/illustrations/ill-cover.webp`). |
| `Tests/`, `UITests/` | XCTest unit tests on the view models, with `Tests/Fixtures/home-zero.json` the recorded `GetHome` answer for the test user; the one UI test above. |
| `scripts/` | `simulator-udid.sh`, `ui-test.sh`. |

Not yet here: the Fredoka and Nunito font files are not bundled, so the type ramp renders in the system
font until they are; the Avatar illustrations are not in the asset catalog, so `Avatar` draws its violet
well alone; `StatHero` shrinks with `minimumScaleFactor` rather than the three discrete steps; the
`swift-snapshot-testing` images of I9 start with the IDR proof (X3).
