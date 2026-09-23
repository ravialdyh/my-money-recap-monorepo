# My Money Recap — iOS app

| | |
|---|---|
| Status | Draft v1, 2026-09-14. Decisions I1–I9 in §2 are the decision register for the app; its Status column is the only place acceptance is recorded. |
| Scope | What is inside the iPhone and iPad app: sign-in, session, networking, the on-device cache, design tokens, formatting, sharing, notifications and tests. The screens themselves are `SCREEN-STATES.md`; the rules are `PRD.md`; the backend is `ARCHITECTURE.md`. |
| Platform | iOS 17+, SwiftUI, iPhone and iPad (PRD D11, D68). |

## 1. Summary

The app is a display for numbers the server computes (`ARCHITECTURE.md` A2). It signs in through Keycloak, calls the BFF with a generated Connect client, stores the last answer for each screen in SwiftData so Home and every Recap render offline (PRD D61), and formats money and dates with Foundation. It contains no business arithmetic: no Spent, no Streak, no Archetype. The only calculations on the phone are platform calls, currency formatting and calendar date maths.

## 2. Decisions

| # | Decision | Why | Rejected | Status |
|---|---|---|---|---|
| I1 | **Sign-in with AppAuth-iOS.** For Apple and Google the OpenID-certified library drives `ASWebAuthenticationSession` with PKCE, discovers Keycloak's endpoints from `.well-known/openid-configuration`, and handles refresh. For email the app opens no browser: an incoming link hands it a login token (the universal link once a domain exists, the custom scheme until then, a launch argument in the UI test, all through one handler) and it sends one `OIDTokenRequest` with the direct grant (`ARCHITECTURE.md` §7.5, A10); the resulting tokens go into the same `OIDAuthState`. That state is built from an authorization response synthesized around the same authorization request the Apple and Google path builds, with the token response applied second: AppAuth builds every refresh out of the last authorization response, and a state built from a token response alone refreshes with a nil client id, an uncatchable exception. | PRD §6.1 names `ASWebAuthenticationSession` and PKCE; AppAuth is the reference implementation of exactly that and removes a hand-written token flow from a money app. | Hand-rolled PKCE: less code to depend on, more code to get wrong. | in use: the authorization-code flow since slice 1, handed straight to Apple or Google by `kc_idp_hint` since #71, with Keycloak's own login page still reached through a debug-only launch argument for the UI test; the login token's direct grant since #70 (Enter email, Check your inbox and the incoming-link handler) |
| I2 | **Session in the Keychain behind `LAContext`.** As PRD §9 "Session on the device" specifies; the iOS-only detail is that the whole AppAuth `OIDAuthState` is the Keychain item. | Straight from PRD §9. | iCloud Keychain sync: would move a session between devices, which D61's per-device cache does not expect. | Keychain item in use since slice 1; since 2026-09-22 every sign-in asks for `offline_access`, which the realm keeps 30 days from last use (ARCHITECTURE §7.5 "Session lifetime"); the `LAContext` gate and the app's own 30-day erase arrive with slice 8. Until then a Member away for 30 days meets the realm's refusal, A1 "sign-in refused"; slice 8 makes the app's clock and the realm's agree so the calm A1 "session expired" line comes first |
| I3 | **Networking is the generated client.** `connect-swift` code generated from `packages/contract` by `buf generate`; one interceptor adds the bearer token and asks AppAuth to refresh on 401; every write sends an `Idempotency-Key` header, one key per save attempt kept across its retries and replaced when the values change (`SaveAttempt`, 2026-09-23, #96: the generated client takes per-call headers, so no second interceptor is needed); every write RPC exists in the contract since slice 2. No hand-written request or response types. | One contract for every client (`ARCHITECTURE.md` A1). | URLSession by hand with Codable structs: a second copy of every message type. | in use since slice 1 (`BearerInterceptor`) |
| I4 | **The cache is one SwiftData model per screen answer.** `HomeSnapshot`, `RecapSnapshot(month)`, `SpendListSnapshot(month)`, `SettingsSnapshot`, each holding the BFF response bytes plus `updatedAt`. A successful response replaces its snapshot whole; the app never edits a snapshot; log out and deletion delete the store (PRD §7 "on device only"). | The cache is read-only by definition, so a normalised schema would be structure with no reader. | Core Data or GRDB with Spend rows: implies the phone recomputes, which A2 forbids. | in use since slice 1 (`HomeSnapshot`) |
| I5 | **Design tokens are generated into Swift.** A `build:swift-tokens` script in `my-money-recap-ds` emits `Tokens.swift` from `tokens.json`; the output is checked in and CI fails if stale, the same rule the CSS outputs already follow. The illustrations are placed in the asset catalog by hand (the cover, and the envelope since #70) until a script exists. | `tokens.json` stays the single source (README "Single source of truth"). | Typing the hex values into Swift: the drift this repo has spent two weeks removing. | in use since slice 1 |
| I6 | **SwiftUI with `@Observable` view models, no framework.** One view model per screen, holding the snapshot and the loading, offline and error states named in `SCREEN-STATES.md`. Navigation is `NavigationStack` plus `.sheet`. | The app has about twenty screens and no shared mutable state beyond the session. | TCA or similar: a second vocabulary for a small app. | in use since slice 1 |
| I7 | **Formatting and dates are Foundation.** Money through `Decimal.FormatStyle.Currency` with the Member's ISO code, the device locale and exactly the decimals the amount carries (PRD D71); `StatHero` sizes through `.minimumScaleFactor` over the three steps (§11), and `AmountField`, a text field that cannot scale, by measuring its line with UIKit's font metrics and taking the largest step that fits; the payday helper's three dates through `Calendar`. | These are platform calls, not rules; the amount's own decimals because the platform's count for a currency differs between iOS versions. `money.ts` in the design system does the same with `Intl`, keeping the platform's decimals for its sample amounts. | A shared Swift package of rules: A2 puts them on the server. | in use since slice 1 (`Money`, `ISODate`); the amount's own decimals since #100 (2026-09-23); `AmountField`'s measured steps and the payday helper since #103 (2026-09-23); `StatHero`'s three steps arrive with the IDR proof |
| I8 | **Share and notifications use the system APIs.** `ImageRenderer` draws screen 4 at 1080 × 1920 for `UIActivityViewController` (PRD §6.15); `UNUserNotificationCenter` asks permission after the first Spend and the device token goes to `RegisterDevice` (§10). | Nothing here needs a library. | — | proposed (slices 6 and 7) |
| I9 | **Tests are unit tests plus snapshot proofs.** XCTest on view models with recorded BFF responses as fixtures; `swift-snapshot-testing` images of Home, the Recap cover and the share image in IDR, which is the proof PRD D45 demands before new work lands; one UI test per way into Home: the magic link, and the authorization-code flow Apple and Google use, driven by a debug-only launch argument (spec #61 seam 6). | The arithmetic is tested on the server; the phone's risk is layout at long numbers and accessibility sizes (D66). | Full UI test suites: slow on Mac runners, and they retest the server's numbers. | in use since slice 1 (view-model tests, one UI test; the magic-link and authorization-code UI tests since #70); the snapshot proofs start with X3 |

## 3. App structure

```
apps/ios/
  MyMoneyRecap.xcodeproj
  App/            entry, the SwiftData container, root switch between Sign in and Home, Info.plist
  Screens/        one folder per SCREEN-STATES.md group, each a view plus its view model (I6): SignIn, Home
  Snapshots/      SwiftData models (I4)
  Networking/     the generated Connect client and the bearer interceptor (I3)
  Auth/           AppAuth configuration and session, the magic-link handler, Keychain (I1, I2)
  DesignSystem/   generated Tokens.swift and the SwiftUI components rebuilt from @my-money-recap/ui (I5)
  Resources/      asset catalog (the cover and envelope illustrations)
  Tests/, UITests/, scripts/   XCTest targets and the simulator scripts (apps/ios/README.md)
```

Later slices add the Onboarding, Spend, Recap and Settings folders under `Screens/`, the Face ID gate in
`App/`, and the app icon and avatars in `Resources/`. The SwiftUI components mirror the names in
`my-money-recap-ds/src/components`, one Swift file per component, so a screen in `SCREEN-STATES.md` reads
the same in both codebases.

## 4. Open items

- Dependency versions are pinned in the Xcode project (`apps/ios/README.md`, "Open"); `swift-snapshot-testing` is pinned when the first snapshot proof lands (X3, the IDR Home).
- Simulator runtimes for I9 are chosen when the tests run, newest and oldest installed (`apps/ios/README.md`, "The UI tests against Compose"). The oldest is iOS 26.0: Xcode 27 offers no iOS 17 or 18 runtime for download, so nothing is proved on either (2026-09-23, #95).
- Mac CI runner: GitHub-hosted macOS runners are the default; the cost is measured after the first month.
