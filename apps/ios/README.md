# apps/ios

The SwiftUI app (`docs/IOS.md`), iOS 17+, iPhone and iPad. So far: Sign in (PRD §6.1) by email, through
Enter email and Check your inbox (§6.2, §6.3) to a link that opens the app signed in with no browser
(I1), Sign in with Apple and Continue with Google handing the system web sheet straight to their
provider (#71), and the authorization-code flow through Keycloak's login page behind a debug-only launch
argument;
the session in the Keychain (I2), the generated Connect client calling `GetHome` (I3), the
`HomeSnapshot` cache in SwiftData (I4) and Home drawn at zero in the B1 "no salary set" layout (PRD §6.8,
D54, D69). Nothing on Home navigates yet; the sheets and screens
behind the chips, cards and buttons arrive with their own tickets.

## Open

```bash
open apps/ios/MyMoneyRecap.xcodeproj
```

Xcode resolves AppAuth-iOS, the local contract package `MoneyRecapContract` (`packages/contract/swift`,
the `connect-swift` client `buf generate` writes) and `connect-swift` and `swift-protobuf`, which the app
also links directly for the interceptor, `Message.with` and the snapshot bytes, at the exact versions that
package pins. The pins live in the Xcode project and `packages/contract/swift/Package.swift` (a plugin
bump moves both, `packages/contract/README.md`), and `Package.resolved` is committed. No signing team
is needed for the simulator, but the build must stay signed ("Sign to Run Locally"): with
`CODE_SIGNING_ALLOWED=NO` the app has no application-identifier entitlement and every Keychain call
fails with `errSecMissingEntitlement`, so a relaunch shows Sign in again.

## Run against local Compose

1. Bring the stack up with `--profile mail` (`infra/compose/README.md`): Keycloak on
   `http://localhost:8080` with the `mmr` realm, its public `ios` client (PKCE S256, redirect
   `mymoneyrecap://auth/callback`) and the test Member `test`, password `TEST_USER_PASSWORD` from
   `infra/compose/.env`; the BFF on `http://localhost:8082` (`BFF_HOST_PORT`), which is what the app
   calls; Mailpit on `http://localhost:8025` (`MAILPIT_HOST_PORT`), where the links arrive.
2. Pick any iPhone simulator in Xcode and Run. The simulator shares the Mac's `localhost`, and the
   app allows plain HTTP on the local network only (`NSAllowsLocalNetworking` in `App/Info.plist`).
3. Tap **Continue with email**, type an address and tap **Send me a link**; Check your inbox names it
   and counts down to Resend. Open the message in Mailpit and hand its `lt` value to the app through
   the custom scheme, `xcrun simctl openurl booted "mymoneyrecap://magic?lt=<value>"`, the handler a
   universal link will reach once a domain exists. Home loads: the first time the §6.19 skeleton (two
   cards at their filled heights, no spinner) until `GetHome` answers, then "Spent so far" at zero,
   `Money in` at zero and no Saved chip (D41), the two `PromptCard`s, `Log a spend`, `See last month`
   disabled and the dashed preview card. Relaunching skips Sign in and draws Home from the stored
   `HomeSnapshot` before the network answers; stop Compose and it still draws the last snapshot. A link
   the realm refuses returns to Enter email with the A3 copy for its cause. Sign in with Apple and
   Continue with Google open the web sheet on `appleid.apple.com` and `accounts.google.com`, which
   answer with their own `invalid_client` page until the realm holds real client ids (#71); cancel the
   sheet to see the `InlineError` inside the card.

Seams, all read at launch: argument `--reset-session` erases the stored session, `--reset-snapshot`
empties the SwiftData store, `--magic-link <link>` hands the app a link as the custom scheme would, and
in a debug build `--authorization-code-flow` makes Continue with email open Keycloak's login page in the
system web sheet, the flow Apple and Google use (sign in as the test Member; cancel it to see the
`InlineError` inside the card). The environment variables `MMR_ISSUER` (default
`http://localhost:8080/realms/mmr`) and `MMR_BFF_URL` (default `http://localhost:8082`) point the app at
another Keycloak and BFF. The magic-link UI test (`UITests/SignInUITests.swift`) points `MMR_BFF_URL` at
an unreachable port to prove Home still renders offline.

## Run on a physical iPhone

The phone reaches the Mac by its LAN address, and the token issuer must be that same address. The three
steps (the `MMR_HOST` variable in `.env`, the `MMR_ISSUER` and `MMR_BFF_URL` Run-action variables, and
setting it back) are in `infra/compose/README.md`, "A phone on the LAN". Sign in by email, opening
`mymoneyrecap://magic?lt=<value>` from Mailpit's message in the phone's Safari, or as the test Member with
`--authorization-code-flow` among the Run action's arguments; Home at zero appears, and a relaunch draws
it from the snapshot. `App/Info.plist` already allows local-network HTTP, so nothing needs to be edited
for a device build.

## Build and test from the command line

Run the build and unit-test steps of the `ios` job in `.github/workflows/ci.yml` from `apps/ios`; they
carry the coverage flags `tools/coverage/README.md` measures with. Delete `build/MyMoneyRecap.xcresult`
before a rerun: xcodebuild refuses an existing result bundle and runs no test. Nothing names a
simulator device or OS version (the "Device and OS coverage" section of issues #16 and #18): the build
uses the generic simulator destination and `scripts/simulator-udid.sh` picks a booted or available
iPhone simulator through `simctl`.

## The UI tests against Compose

`UITests/SignInUITests.swift` signs in by magic link (I9): it types a fresh address,
`mmr-test-<uuid>@example.test`, because a fixed one would meet PRD §11's five links per address per hour
on the second run within the hour; watches Check your inbox count down; reads the message from Mailpit
over HTTP from inside the test process; and relaunches the app with the link as `--magic-link`, the
handler the custom scheme and a universal link reach. It then asserts Home at zero, relaunches to prove
Home renders from the snapshot, relaunches with `MMR_BFF_URL` pointing at a closed port to prove the
snapshot survives an unreachable BFF, and relaunches with `--reset-snapshot` to prove the first load is
the skeleton. `UITests/AuthorizationCodeUITests.swift` keeps the flow every Apple and Google Member will
use under test: with `--authorization-code-flow` it cancels the web sheet, then signs in as the test Member
through Keycloak's login page, which covers the PKCE request, the sheet and the state built from its
authorization response. The first skips itself unless the runner names Mailpit (`MMR_MAILPIT_URL`), the
second unless it carries the test Member's password; CI's `ios` job passes
`-only-testing:MyMoneyRecapTests`, so it never runs them. No UI test opens Apple's or Google's page:
with placeholder client ids each answers with its own error page, which the repository neither controls
nor may assert as a round-trip (#71), so the hint the app sends is a unit test
(`Tests/AuthConfigTests.swift`) and the sheet on the provider's host was checked by hand.

The same section asks for two simulator runtimes, both chosen when the script runs and never written
down: the newest installed iOS runtime and the oldest. The oldest is iOS 26.0, not the deployment target's
iOS 17: Xcode 27, which builds this app, offers no iOS 17 or 18 runtime for download, so 26.0 is the
oldest that can be installed, and nothing here is proved on iOS 17 or 18 (#95).
`scripts/ui-test.sh` builds once, then runs both UI tests on each against the stack
`infra/compose/.env` describes: its `COMPOSE_PROJECT_NAME` and its Keycloak, BFF and Mailpit host ports,
each falling back to the default, and each overridden by `MMR_ISSUER`, `MMR_BFF_URL` or
`MMR_MAILPIT_URL` when set. It takes the password from `MMR_TEST_USER_PASSWORD` or, failing that, from
the `<project>-realm-apply-1` container, which carries it as `TF_VAR_test_user_password`. That container
re-applies the realm on every `up`, so a changed `.env` value is live after the next `up`
(`infra/compose/README.md`, "The `mmr` realm"):

```bash
scripts/ui-test.sh
```

Result bundles land in `build/ui-newest.xcresult` and `build/ui-oldest.xcresult`, with the screenshots
exported next to them; `scripts/ui-test.sh oldest` (or `newest`) runs one destination.
`scripts/simulator-udid.sh newest|oldest` is the selection on its own (it creates an iPhone on a runtime
that has none), and with no argument it keeps the booted-or-first behaviour of the `ios` job's
unit-test step.

Install iOS 26.0 with `xcodebuild -downloadPlatform iOS -buildVersion 26.0` (about 7.5 GB); the script
picks it up on the next run and names both simulators as it runs. With one runtime installed it says so
and runs once.

## Layout

| Folder | What lives here |
| --- | --- |
| `App/` | Entry point, the SwiftData container, root switch between Sign in and Home (sign-out deletes the cache), `Info.plist` (URL scheme, the ATS local-networking allowance). |
| `Auth/` | `AuthConfig` (issuer, client id, redirect, the login-token grant; no secret exists in the target), `AuthSession` (AppAuth discovery, PKCE authorization, `OIDAuthState`, the incoming magic link, fresh tokens, and a sign-out only when the realm refuses a refresh), `MagicLink` (the login token from any of the three sources, the direct grant, the state synthesized around the authorization request, and the failure codes behind the A3 copies and A1's error copy), `Keychain`. |
| `Networking/` | `BffClient` builds the generated `MoneyRecap_V1_BffServiceClient` (proto codec, `MMR_BFF_URL` or its default); `BearerInterceptor` is the one interceptor of I3 so far, and sends `RequestMagicLink` with no token, keyed by its path in both halves. No request or response type is written here. |
| `Snapshots/` | `HomeSnapshot`, the one SwiftData model of I4: the `HomeView` bytes plus `updatedAt`, replaced whole on every success. |
| `Screens/` | One folder per SCREEN-STATES group, each screen a view plus an `@Observable` view model (I6): `SignIn/` holds A1 Sign in, A2 Enter email and A3 Check your inbox on one navigation stack; `Home/`. |
| `DesignSystem/` | Generated `Tokens.swift` (do not edit; `npm run build:swift-tokens`), `TextStyle+Font` (applies a type-ramp entry to a view), `Money` (the platform currency formatter, I7) and the SwiftUI components mirrored one file per name from `my-money-recap-ds/src/components`: `Screen`, `Headline` (+ `SubCopy`, `Eyebrow`), `Illustration`, `Avatar`, `Card` (+ `CardTitle`), `Chip`, `StatHero`, `StatChip`, `PromptCard`, `ListRow`, `MMRButton`, `MMRTextField`, `InlineError`. Colours, radii and type come only from the tokens. The React row wrappers (`ChipRow`, `StatChipRow`) are a plain `HStack` here, and `PromptCard` has a `dashed` variant (the D47 outline card) the React one draws by hand. |
| `Resources/` | Asset catalog with the cover and envelope illustrations (PNGs from `my-money-recap/assets/illustrations/`, `ill-cover.webp` and `ill-envelope.webp`). |
| `Tests/`, `UITests/` | XCTest unit tests on the view models, the magic-link handler and session (against a stub realm), the Keychain and the token refresh, with `Tests/Fixtures/home-zero.json` the recorded `GetHome` answer for the test Member; the two UI tests above. |
| `scripts/` | `simulator-udid.sh`, `ui-test.sh`. |

Not yet here: the Fredoka and Nunito font files are not bundled, so the type ramp renders in the system
font until they are; the Avatar illustrations are not in the asset catalog, so `Avatar` draws its violet
well alone; `StatHero` shrinks with `minimumScaleFactor` rather than the three discrete steps; the
`swift-snapshot-testing` images of I9 start with the IDR proof (X3).
