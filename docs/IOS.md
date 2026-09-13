# My Money Recap — iOS app

| | |
|---|---|
| Status | Draft v1, 2026-09-14. Decisions I1–I9 are proposed; each becomes an ADR under `adr/` once accepted. |
| Scope | What is inside the iPhone and iPad app: sign-in, session, networking, the on-device cache, design tokens, formatting, sharing, notifications and tests. The screens themselves are `SCREEN-STATES.md`; the rules are `PRD.md`; the backend is `ARCHITECTURE.md`. |
| Platform | iOS 17+, SwiftUI, iPhone and iPad (PRD D11, D68). |

## 1. Summary

The app is a display for numbers the server computes (`ARCHITECTURE.md` A2). It signs in through Keycloak, calls the BFF with a generated Connect client, stores the last answer for each screen in SwiftData so Home and every Recap render offline (PRD D61), and formats money and dates with Foundation. It contains no business arithmetic: no Spent, no Streak, no Archetype. The only calculations on the phone are platform calls, currency formatting and calendar date maths.

## 2. Decisions

| # | Decision | Why | Rejected |
|---|---|---|---|
| I1 | **Sign-in with AppAuth-iOS.** The OpenID-certified library drives `ASWebAuthenticationSession` with PKCE, discovers Keycloak's endpoints from `.well-known/openid-configuration`, and handles refresh. | PRD §6.1 names `ASWebAuthenticationSession` and PKCE; AppAuth is the reference implementation of exactly that and removes a hand-written token flow from a money app. | Hand-rolled PKCE: less code to depend on, more code to get wrong. |
| I2 | **Session in the Keychain behind `LAContext`.** Refresh token stored with `kSecAttrAccessibleWhenUnlockedThisDeviceOnly`, released only after Face ID, Touch ID or passcode; a `lastUsedAt` stamp enforces the 30-day rule (PRD D62); log out erases the item and the cache. | Straight from PRD §9. | iCloud Keychain sync: would move a session between devices, which D61's per-device cache does not expect. |
| I3 | **Networking is the generated client.** `connect-swift` code generated from `packages/contract` by `buf generate`; one interceptor adds the bearer token and asks AppAuth to refresh on 401; one interceptor adds the `Idempotency-Key` on writes. No hand-written request or response types. | One contract for every client (`ARCHITECTURE.md` A1). | URLSession by hand with Codable structs: a second copy of every message type. |
| I4 | **The cache is one SwiftData model per screen answer.** `HomeSnapshot`, `RecapSnapshot(month)`, `SpendListSnapshot(month)`, `SettingsSnapshot`, each holding the BFF response bytes plus `updatedAt`. A successful response replaces its snapshot whole; the app never edits a snapshot; log out and deletion delete the store (PRD §7 "on device only"). | The cache is read-only by definition, so a normalised schema would be structure with no reader. | Core Data or GRDB with Spend rows: implies the phone recomputes, which A2 forbids. |
| I5 | **Design tokens are generated into Swift.** A `build:swift-tokens` script in `my-money-recap-ds` emits `Tokens.swift` (colours, radii, type ramp, motion curves) from `tokens.json`, and `build:assets` also writes the WebP illustrations into an asset catalog. Both outputs are checked in and CI fails if stale, the same rule the CSS outputs already follow. | `tokens.json` stays the single source (README "Single source of truth"). | Typing the hex values into Swift: the drift this repo has spent two weeks removing. |
| I6 | **SwiftUI with `@Observable` view models, no framework.** One view model per screen, holding the snapshot and the loading, offline and error states named in `SCREEN-STATES.md`. Navigation is `NavigationStack` plus `.sheet`. | The app has about twenty screens and no shared mutable state beyond the session. | TCA or similar: a second vocabulary for a small app. |
| I7 | **Formatting and dates are Foundation.** Money through `Decimal.FormatStyle.Currency` with the Member's ISO code and the device locale (PRD D71); `StatHero` and `AmountField` sizes through `.minimumScaleFactor` over the three steps (§11); the payday helper's three dates through `Calendar`. | These are platform calls, not rules; `money.ts` in the design system does the same with `Intl`. | A shared Swift package of rules: A2 puts them on the server. |
| I8 | **Share and notifications use the system APIs.** `ImageRenderer` draws screen 4 at 1080 × 1920 for `UIActivityViewController` (PRD §6.15); `UNUserNotificationCenter` asks permission after the first Spend and the device token goes to `RegisterDevice` (§10). | Nothing here needs a library. | — |
| I9 | **Tests are unit tests plus snapshot proofs.** XCTest on view models with recorded BFF responses as fixtures; `swift-snapshot-testing` images of Home, the Recap cover and the share image in IDR, which is the proof PRD D45 demands before new work lands; one UI test for the sign-in-to-Home path. | The arithmetic is tested on the server; the phone's risk is layout at long numbers and accessibility sizes (D66). | Full UI test suites: slow on Mac runners, and they retest the server's numbers. |

## 3. App structure

```
apps/ios/
  MyMoneyRecap.xcodeproj
  App/            entry, session gate (Face ID), root navigation
  Screens/        one folder per SCREEN-STATES.md group: SignIn, Onboarding, Home, Spend, Recap, Settings
  Snapshots/      SwiftData models (I4)
  Networking/     generated Connect client + the two interceptors (I3)
  Auth/           AppAuth configuration, Keychain, LAContext (I1, I2)
  DesignSystem/   generated Tokens.swift and the SwiftUI components rebuilt from @my-money-recap/ui (I5)
  Resources/      asset catalog with the generated illustrations and app icon
```

The SwiftUI components mirror the names in `my-money-recap-ds/src/components`, one Swift file per component, so a screen in `SCREEN-STATES.md` reads the same in both codebases.

## 4. Open items

- The exact `swift-snapshot-testing` and AppAuth-iOS versions are pinned when the Xcode project is created, not here.
- Mac CI runner: GitHub-hosted macOS runners are the default; the cost is measured after the first month.
