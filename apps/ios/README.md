# apps/ios

The SwiftUI app (`docs/IOS.md`), iOS 17+, iPhone and iPad. It covers sign-in (email link, Apple, Google),
onboarding, Home, the Spend list and sheets, the four Recap pages and the share poster, Settings with
Profile, Salary and payday, Saving target, Currency and Notifications, and the push permission. What
each screen shows is `docs/SCREEN-STATES.md`; the rules are `docs/PRD.md` §6; the code is under
`Screens/` (see Layout below).

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
   disabled and the dashed preview card. `Log a spend` opens its sheet: an amount, a Category, today or a day in the
   last 60, a note, and Save, after which Home's numbers move to the answer's, and a Spend that first passes the warning line puts the coral Target Warning above the target card, whose × hides it until the next launch. With the Mac's network off, the grey `No connection. You can look, but not log.` sits under the chips and `Log a spend` dims to 40%, as do the empty Spend list's plus button and an opened Spend's `Save changes` and `Delete this spend`; all of it goes when the network is back, with no relaunch. Relaunching skips Sign in and draws Home from the stored
   `HomeSnapshot` before the network answers; stop Compose and it still draws the last snapshot. A link
   the realm refuses returns to Enter email with the A3 copy for its cause. Sign in with Apple and
   Continue with Google open the web sheet on `appleid.apple.com` and `accounts.google.com`, which
   answer with their own `invalid_client` page until the realm holds real client ids; cancel the
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
simulator device or OS version: the build
uses the generic simulator destination and `scripts/simulator-udid.sh` picks a booted or available
iPhone simulator through `simctl`.

## The UI tests against Compose

`UITests/SignInUITests.swift` signs in by magic link (I9): it types a fresh address,
`mmr-test-<uuid>@example.test`, because a fixed one would meet PRD §11's five links per address per hour
on the second run within the hour; watches Check your inbox count down; reads the message from Mailpit
over HTTP from inside the test process; and relaunches the app with the link as `--magic-link`, the
handler the custom scheme and a universal link reach. The new Member meets step 1; the test force-quits
and relaunches to prove step 1 reopens, types a name with a digit the field refuses, picks an Avatar and
taps Next, then goes back from step 2 to find both answers re-selected. Part-way through setup it force-quits
on step 2 to prove step 2 reopens with step 1's answers restored from the server, and relaunches with
`MMR_BFF_URL` pointing at a closed port to prove it meets the no-connection state, whose Try again keeps it
there. Back online, at step 2 it searches for a word no currency matches, clears the search, finds the rupiah by
name and taps Next. At step 3 it goes back to step 2 and forward again, types a twelfth digit the field drops
beside `That's more than the app can store.`, types an amount, picks a Payday and taps Next. It then asserts
Home at zero greeting the Member in rupiah with the Salary prompt gone, logs one Spend from Home (Save off
until an amount and a Category, the date picker opened, the sheet held sideways) and sees the hero show it,
opens the Spend list from the hero, the Category detail and back, and the switcher from the month chip, then
edits the Spend, cancels the delete alert once and deletes it (`UITests/EditSpendSteps.swift`), opens the month so far
from the dashed card under `Past recaps`, finds P1 with no spends, turns the phone reaching Close, swipes to P2 and
turns it reaching Close, swipes to P3 (nothing
spent, the Payday card only before Payday), turns it reaching Close, swipes to P4 (its closing line, the pills and no
Share), turns it reaching Close, taps the first page dot back to P1 and closes back to Home (`UITests/RecapSteps.swift`),
opens Settings from its chip, reads each row's value, turns it reaching Log out, renames the Member in Profile, turned
too, and goes back to Home greeting the new name, raises the Saving Target from next month, then converts the
rupiah to pounds: the stack has no rates App ID, so the fetch fails, Try again fails too, and it takes Enter it
myself, types a rate, sees the Salary previewed, converts and finds D1 and Home in pounds, then opens the month so far
again so it is stored in pounds, then opens Notifications from a row reading `2 on`, turns the Daily reminder on,
sees `Remind me at`, turns D8 reaching Target warning, and finds the row reading `3 on`, then opens Export spends,
sends it and sees `Sent ✓`, dimmed, turned too (`UITests/SettingsSteps.swift`; each later Settings step joins it, the sign-out steps
last), relaunches to prove Home renders from the snapshot, and relaunches with the closed port again to prove a Member who finished setup keeps that Home with no
signal, in pounds, and that the dashed card opens the stored P1. The first-load skeleton is a picture test (`HomeViewTests`). `UITests/AuthorizationCodeUITests.swift` keeps the flow every Apple and Google Member will
use under test: with `--authorization-code-flow` it cancels the web sheet, then signs in as the test Member
through Keycloak's login page, which covers the PKCE request, the sheet and the state built from its
authorization response, answers step 1 with the name the token's `given_name` filled in, and at step 2
finds the simulator region's own currency already chosen, so it just taps Next, then taps Skip for now at
step 3. The first skips itself unless the runner names Mailpit (`MMR_MAILPIT_URL`), the
second unless it carries the test Member's password; CI's `ios` job passes
`-only-testing:MyMoneyRecapTests`, so it never runs them. No UI test opens Apple's or Google's page:
with placeholder client ids each answers with its own error page, which the repository neither controls
nor may assert as a round-trip, so the hint the app sends is a unit test
(`Tests/AuthConfigTests.swift`) and the sheet on the provider's host was checked by hand.

The same section asks for two simulator runtimes, both chosen when the script runs and never written
down: the newest installed iOS runtime and the oldest. The oldest is iOS 26.0, not the deployment target's
iOS 17: Xcode 27, which builds this app, offers no iOS 17 or 18 runtime for download, so 26.0 is the
oldest that can be installed, and nothing here is proved on iOS 17 or 18.
`scripts/ui-test.sh` builds once, then runs both UI tests on each against the stack
`infra/compose/.env` describes: its `COMPOSE_PROJECT_NAME` and its Keycloak, BFF and Mailpit host ports,
each falling back to the default, and each overridden by `MMR_ISSUER`, `MMR_BFF_URL` or
`MMR_MAILPIT_URL` when set. Before each runtime it deletes the realm test Member's `members` row, so the
Keycloak test always meets step 1, and uninstalls the app, so the magic-link test meets the notification
permission alert after its first Spend (B5) and taps Allow, which springboard draws. It takes the password from `MMR_TEST_USER_PASSWORD` or, failing that, from
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
| `App/` | Entry point, the SwiftData container, `Push` (`PushSystem`, the seam over iOS's permission and APNs that the unit tests fake, `PushRegistration` and the `AppDelegate` that hears the token), root switch between Sign in and the launch (sign-out deletes the cache), `Launch` and `LaunchView` (the zone, the restore sheet for a pending deletion, then Home or the first unanswered of steps 1 to 4), `Info.plist` (URL scheme, the ATS local-networking allowance). |
| `Auth/` | `AuthConfig` (issuer, client id, redirect, the login-token grant; no secret exists in the target), `AuthSession` (AppAuth discovery, PKCE authorization, `OIDAuthState`, the incoming magic link, fresh tokens, and a sign-out only when the realm refuses a refresh), `MagicLink` (the login token from any of the three sources, the direct grant, the state synthesized around the authorization request, and the failure codes behind the A3 copies and A1's error copy), `Keychain`. |
| `Networking/` | `BffClient` builds the generated `MoneyRecap_V1_BffServiceClient` (proto codec, `MMR_BFF_URL` or its default) and calls once more after the interceptor refreshes a refused token (`retried`); `BearerInterceptor` is the one interceptor of I3, and sends `RequestMagicLink` with no token, keyed by its path in both halves. `SaveAttempt` holds one save attempt's `Idempotency-Key`, the same across its retries and new when the values change, which a write passes as a per-call header. `Connection` follows the platform's network path (`NWPathMonitor`): with none, Home shows the grey offline `Banner` and disables `Log a spend`, as the empty Spend list does its plus button and the Edit spend sheet its Save changes and Delete (D58). No request or response type is written here. |
| `Snapshots/` | The SwiftData models of I4, each the answer's bytes plus `updatedAt`, replaced whole on every success: `HomeSnapshot` (the `HomeView`), `SpendListSnapshot(month)` (the Spend list view of one month), `RecapSnapshot(month)` (the `RecapView` of one month) and `SettingsSnapshot` (the `SettingsView`). |
| `Screens/` | One folder per SCREEN-STATES group, each screen a view plus an `@Observable` view model (I6): `SignIn/` holds A1 Sign in, A2 Enter email and A3 Check your inbox on one navigation stack; `Onboarding/` holds A4 Name and avatar, with `DisplayName`, the rule `members` applies (NFC, 1 to 30 code points of letters, spaces, hyphens and apostrophes), which the First name field enforces as it is typed, A5 Currency, the list built from `Locale.commonISOCurrencyCodes`, the platform's common currencies, named and sorted in British English, A6 Salary and Payday, whose three helper dates are `Calendar` arithmetic (PRD §8.2), and `NoConnectionView`, setup's one offline state; `Home/` holds B1 Home, B2 to B4 Log a spend and B7 Edit spend (one sheet, `LogSpendView`) and B6 and B6b, the Spend list and a Category's Spends; `Recap/` holds the Recap pager (`RecapPagerView`), its pages, each drawn by its case on its own ground: the cover, C1 and P1 (`RecapCoverPage`), kept and streak, C2 and P2 (`RecapKeptPage`, beside page 2's words), target and spending, C3 and P3 (`RecapGoalPage`), and the Archetype, C4 and P4 (`RecapArchetypePage`), beside `SharePoster`, the C6 share image drawn from its five facts, and `ShareSheet`, the system share sheet it opens in; `Account/` holds D11b Restore account; `Settings/` holds D1 Settings, D2 Profile, D3 Salary and payday with its D56 sheet, D4 Saving target and D5 Currency with its D6 and D7 sheet, and D8 Notifications. |
| `DesignSystem/` | Generated `Tokens.swift` (do not edit; `npm run build:swift-tokens`), `TextStyle+Font` (applies a type-ramp entry to a view, its line height measured against the font's own), `Money` (the platform currency formatter at the amount's own decimals, I7) and the SwiftUI components mirrored one file per name from `my-money-recap-ds/src/components`: `Screen` (with its `ground`, ink type on the yellow one), `Headline` (`l`, `m`, `compact` and `display`, the white `l` and `display` drawn in a label at their tokens' line heights, which are under the font's own; + `SubCopy`, `Eyebrow` with its coral tone), `Illustration` (hanging into the card below by `hang`, filling the free height above it unless `grows` is false, a tall one capped by `maxHeight`, never shrunk to fit), `Avatar`, `AvatarGrid` (the six keys `Profile.avatar_key` lists, each named for VoiceOver), `Card` (+ `CardTitle`; `split` for an illustration at its edge, `accent` the violet card on the yellow ground), `Chip` (with `asLabel`, the step label pill, `tone`, a Category's selection chip, `highlight`, the yellow pill naming the Member on the Recap cover, and `inverse`, the ink pill), `ProgressBar` (violet, green and coral, with the `piggy` marker drawn from `glyphs.tsx`'s shapes), `CategoryBar`, `Icon` (the eight Category glyphs drawn from `glyphs.tsx`'s shapes), `StatHero` (80, 56 or 40pt, measured), `StatChip` (earned, saved and spent), `StatChipRow` (chips that wrap to a new line), `PromptCard`, `ListRow` (a chevron, the check a chosen row carries, or nothing; its label coral when destructive; disabled, the whole row at 40%), `Sheet` (the `full` variant: the white surface filling the ground under a back chevron and headline, D1 to D5 and D8), `MMRButton`, `MMRTextField` (with its optional suffix, D7's code), `AmountField` (the Currency's symbol and decimals, the D57 cap, grouped as typed inside a `UITextField` delegate so a fast keystroke is never lost), `DayPicker` (1 to 31 and Last day; its dates are the caller's), `InlineError` (with its optional action, D6's `Try again`), `Banner` (the coral Target Warning, the violet info tone with an optional link, D8's `Open Settings`, and the grey no-connection tone, with an optional ×), `MMRToggle` (`Toggle.tsx`'s 51 by 31 switch, violet when on, drawn rather than the system's, whose size differs between iOS versions, and heard by VoiceOver as a switch), `TextLink`, the helper-text link, `BackRow`, the canvases' 44pt back chevron beside a label pill, `StreakBadge`, `StreakDots` (one per day of the window, the run lit), `CompareBars` (two or more bars scaled to their total or their largest, revealed by `Reveal`), `PageDots` (each dot a control in its own 44pt-tall slice, ink `onLight`), `ArchetypePill` and `ArchetypeRow` (the four pills, the month's picked out) and `CloseControl`, the Recap pager's 44pt close control, ink `onLight`, which the canvases draw and the React design system does not. Colours, radii and type come only from the tokens. The React `ChipRow` is a plain `HStack` here, and `PromptCard` has a `dashed` variant (the D47 outline card) the React one draws by hand. |
| `Resources/` | Asset catalog with the cover, envelope, jar and vending illustrations, the four Archetype mascots and the six Avatars at 3x, named by their contract keys (PNGs from `my-money-recap/assets/illustrations/`, `ill-cover.webp`, `ill-envelope.webp`, `ill-jar.webp`, `ill-vending.webp`, `ill-squirrel.webp`, `ill-hustler.webp`, `ill-treat.webp`, `ill-slowburner.webp` and `ill-avatar-<key>.webp`); `Fonts/` with Fredoka and Nunito and their licences. |
| `Tests/`, `UITests/` | XCTest unit tests on the view models, the magic-link handler and session (against a stub realm), the Keychain and the token refresh, with `Tests/Fixtures/home-zero.json` the recorded `GetHome` answer for the test Member after steps 1 and 2, in rupiah, `Tests/Fixtures/home-recaps.json` one recorded with two past months, a Hustler and a Partial, `log-spend.json` a recorded `LogSpend` answer, `list-spends.json` a recorded `ListSpends` answer, `settings-filled.json` and `settings-not-set.json` recorded `GetSettings` answers (a Salary on the 25th, a 20% Target and all three notifications on, then neither and PRD §10's defaults), `currency-*.json` recorded `PreviewCurrencyChange` and `ChangeCurrency` answers (nothing to convert and the direct change, then a Rp150.000.000 Salary and a Rp25.000 Spend previewed and converted to pounds at a typed 0.000046) and `recap-*.json` `GetRecap` answers (a full Squirrel month, the month so far, a sign-up month, a month with no Money In, a one-day run, no run (a Hustler), an overspent month and a Slow Burner, with `recap-maya-september.json` ledger's answer key for §8.6 rather than a recording, since the stack cannot close a September before 1 October; for screen 3, Maya's months in pounds: target met, missed, no Spend, no Saving Target, and the month so far before and after Payday); steps 1 and 2, Home's first-load skeleton, its rupiah proof (X3), its Target Warning banner before and after Payday and dismissed, its offline Banner with and without the warning, the Log a spend and Edit spend sheets, the Spend list with its empty months and Category detail, the Recap cover, C1, P1, its loading state, held sideways and in rupiah (X4), page 2, C2 and its variants, P2 and held sideways, screen 3, C3 in each state and P3, held sideways, and screen 4, C4 for each Archetype and without one and P4 in each situation, held sideways, Settings filled, with nothing set, offline, with nothing stored and with push denied, D4 with and without its line and with no Salary, and D1, D2 and D4 held sideways, D6 in rupiah, its failed fetch, D7 before and after its preview, D5 with `Changed ✓`, D5 to D7 held sideways, and D8 with the reminder on and denied, held sideways too, drawn in a window (`NameAvatarViewTests`, `SettingsViewTests`, `CurrencyEditViewTests`, `CurrencyViewTests`, `HomeViewTests`, `LogSpendViewTests`, `SpendListViewTests`, `RecapViewTests`, `RecapKeptTests`, `RecapGoalViewTests`, `RecapArchetypeTests`, all through `Picture.swift`); the share image, one poster per Archetype at 1080 by 1920 and its rupiah proof (X5), with the share sheet presented and dismissed (`SharePosterTests`); every listed currency checked against the backend's CLDR file (`CurrencyViewModelTests`, which reads it from the repository); the two UI tests above. |
| `scripts/` | `simulator-udid.sh`, `ui-test.sh`. |

The Fredoka and Nunito variable fonts sit in `Resources/Fonts` with their SIL Open Font License text, from
github.com/google/fonts, and `Info.plist` registers them (`UIAppFonts`), so the type ramp and the amount field's
measuring both use them. Every screen works held sideways on an iPhone: a screen with no scroll view of its
own passes `Screen(scrolls: true)`, and steps 1 and 3 scroll their headline with their cards; the magic-link UI
test turns the phone at each screen and reaches its main control. Each Recap page scrolls the same way
while swiping still turns the page; the iPad column widths are left to the hardening slice.
