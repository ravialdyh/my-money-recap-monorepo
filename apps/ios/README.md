# apps/ios

The SwiftUI app (`docs/IOS.md`), iOS 17+, iPhone and iPad. Built so far:

- **Sign in** (PRD §6.1 to §6.3): by email, through Enter email and Check your inbox to a link that opens the app
  signed in with no browser (I1); Sign in with Apple and Continue with Google, which hand the system web sheet
  straight to their provider (#71); and the authorization-code flow through Keycloak's login page, behind a
  debug-only launch argument. The session lives in the Keychain (I2).
- **Launch**: the generated Connect client (I3) sends the device's zone with `SetTimezone` on every launch, under
  one `Idempotency-Key` (A3, A14), and the launch then routes as IOS I4 orders it: D11b Restore account, a stored
  Home, or the first unanswered onboarding step. On the restore sheet (PRD §6.18, D60, #105), "Keep my account"
  clears the deletion and continues, and "Continue with deletion" signs out on the phone; nothing in the app asks
  for a deletion yet.
- **Onboarding** (spec #92): step 1, Name and Avatar (PRD §6.4, #98), step 2, Currency (§6.5, #100), step 3,
  Salary and Payday (§6.6, #103), and step 4, Saving Target (§6.7, #104), each with the answers already saved
  filled in. Step 4 without a Salary shows the prompt back to step 3 and one Skip for now (D44). Each back chevron
  returns to the step before without losing a later answer (#102), and a Currency saved again during setup
  relabels step 3's Salary. With no connection, setup says `No connection. We need one to save your answers.`: on
  its own with Try again at launch, and in the step's card otherwise, where Next tries again.
- **Home** (PRD §6.8, D54, D69): drawn at zero, from the `HomeSnapshot` cache (I4) when one is stored, greeting the
  Member by name with their Avatar and printing every figure in their Currency, with the Saving Target's card
  filled in once one is set. A `failed_precondition` from `GetHome`, step 1 or 2 unanswered, deletes the stored
  Home and reopens that step. The two prompt cards reopen step 3 and step 4, whose back chevron then returns to
  Home. `Log a spend` opens its sheet (§6.9, #129), and a save moves Home's numbers to the answer's; a Spend that first
  passes the warning line puts the coral Target Warning above the target card (§8.8, #132). With no network path,
  the grey offline `Banner` shows and `Log a spend`, the empty list's plus button and an opened Spend's two actions
  dim to 40% (D58, #134). The hero opens the Spend list (§6.10, #131), and the month chip opens it at its month switcher (D59): one
  row per Category, largest first, each opening its Spends, and a month already opened reads offline from its
  `SpendListSnapshot`. A Spend in a Category opens the Edit spend sheet (§6.10, #133), the log sheet prefilled with
  Save changes and Delete this spend; after either write Home and every month it changed are asked again. Under a `Past recaps` label,
  each past-Recap row opens its month's Recap, and before the first Recap the dashed preview card opens the month so
  far; `See last month` is enabled only when the month before Home's is in that list, and opens it (spec #151, #153).
  A Recap opens full screen on its first page, the cover (C1 for a full month, P1 for the month so far and both
  partial situations), with its page dots and the close control that returns to Home. A month already opened draws
  its stored `RecapSnapshot` at once and reads offline; with nothing stored, until the answer it draws the cover's
  ground and illustration only, and the numbers fade in when it arrives. A swipe turns to page 2 on the yellow
  ground, where the close control and the dots turn ink: C2, what was kept against the Saving Target as two bars (the
  target card, with no action until slice 7, when there is none) and the longest no-spend streak, its dates from
  ledger and its dots with the run lit; P2, the streak alone, its dots stopping at today in the current month (#154).
  Swiping on reaches the target and spending page (C3, P3, #155), then the last page on the green ground (#156): C4
  names the month's Archetype with ledger's title and description, its mascot, Saved against Spent as two bars and
  `Share your {Archetype}`, which draws the 1080 by 1920 poster (C6) and opens the system share sheet (#157), over the
  four Archetype pills with the month's picked out and no page dots; with no Saving
  Target the target card, with no action until slice 7, stands in the mascot's place, the pills sit at 40% and there
  is no Share. P4 closes with ledger's closing line over the pills at 40%. Bars reveal as their page comes on screen.
  The Settings chip does nothing until slice 7.

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
Share), turns it reaching Close, taps the first page dot back to P1 and closes back to Home (`UITests/RecapSteps.swift`), relaunches to prove Home renders from the snapshot, and relaunches with the closed port again to prove a Member who finished setup keeps that Home with no
signal and that the dashed card opens the stored P1. The first-load skeleton is a picture test (`HomeViewTests`). `UITests/AuthorizationCodeUITests.swift` keeps the flow every Apple and Google Member will
use under test: with `--authorization-code-flow` it cancels the web sheet, then signs in as the test Member
through Keycloak's login page, which covers the PKCE request, the sheet and the state built from its
authorization response, answers step 1 with the name the token's `given_name` filled in, and at step 2
finds the simulator region's own currency already chosen, so it just taps Next, then taps Skip for now at
step 3. The first skips itself unless the runner names Mailpit (`MMR_MAILPIT_URL`), the
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
`MMR_MAILPIT_URL` when set. Before each runtime it deletes the realm test Member's `members` row, so the
Keycloak test always meets step 1. It takes the password from `MMR_TEST_USER_PASSWORD` or, failing that, from
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
| `App/` | Entry point, the SwiftData container, root switch between Sign in and the launch (sign-out deletes the cache), `Launch` and `LaunchView` (the zone, the restore sheet for a pending deletion, then Home or the first unanswered of steps 1 to 4), `Info.plist` (URL scheme, the ATS local-networking allowance). |
| `Auth/` | `AuthConfig` (issuer, client id, redirect, the login-token grant; no secret exists in the target), `AuthSession` (AppAuth discovery, PKCE authorization, `OIDAuthState`, the incoming magic link, fresh tokens, and a sign-out only when the realm refuses a refresh), `MagicLink` (the login token from any of the three sources, the direct grant, the state synthesized around the authorization request, and the failure codes behind the A3 copies and A1's error copy), `Keychain`. |
| `Networking/` | `BffClient` builds the generated `MoneyRecap_V1_BffServiceClient` (proto codec, `MMR_BFF_URL` or its default) and calls once more after the interceptor refreshes a refused token (`retried`); `BearerInterceptor` is the one interceptor of I3, and sends `RequestMagicLink` with no token, keyed by its path in both halves. `SaveAttempt` holds one save attempt's `Idempotency-Key`, the same across its retries and new when the values change, which a write passes as a per-call header. `Connection` follows the platform's network path (`NWPathMonitor`): with none, Home shows the grey offline `Banner` and disables `Log a spend`, as the empty Spend list does its plus button and the Edit spend sheet its Save changes and Delete (D58). No request or response type is written here. |
| `Snapshots/` | The SwiftData models of I4, each the answer's bytes plus `updatedAt`, replaced whole on every success: `HomeSnapshot` (the `HomeView`), `SpendListSnapshot(month)` (the Spend list view of one month) and `RecapSnapshot(month)` (the `RecapView` of one month). |
| `Screens/` | One folder per SCREEN-STATES group, each screen a view plus an `@Observable` view model (I6): `SignIn/` holds A1 Sign in, A2 Enter email and A3 Check your inbox on one navigation stack; `Onboarding/` holds A4 Name and avatar, with `DisplayName`, the rule `members` applies (NFC, 1 to 30 code points of letters, spaces, hyphens and apostrophes), which the First name field enforces as it is typed, A5 Currency, the list built from `Locale.commonISOCurrencyCodes`, the platform's common currencies, named and sorted in British English, A6 Salary and Payday, whose three helper dates are `Calendar` arithmetic (PRD §8.2), and `NoConnectionView`, setup's one offline state; `Home/` holds B1 Home, B2 to B4 Log a spend and B7 Edit spend (one sheet, `LogSpendView`) and B6 and B6b, the Spend list and a Category's Spends; `Recap/` holds the Recap pager (`RecapPagerView`), its pages, each drawn by its case on its own ground: the cover, C1 and P1 (`RecapCoverPage`), kept and streak, C2 and P2 (`RecapKeptPage`, beside page 2's words), target and spending, C3 and P3 (`RecapGoalPage`), and the Archetype, C4 and P4 (`RecapArchetypePage`); `Account/` holds D11b Restore account. |
| `DesignSystem/` | Generated `Tokens.swift` (do not edit; `npm run build:swift-tokens`), `TextStyle+Font` (applies a type-ramp entry to a view, its line height measured against the font's own), `Money` (the platform currency formatter at the amount's own decimals, I7) and the SwiftUI components mirrored one file per name from `my-money-recap-ds/src/components`: `Screen` (with its `ground`, ink type on the yellow one), `Headline` (`l`, `m`, `compact` and `display`, the white `l` and `display` drawn in a label at their tokens' line heights, which are under the font's own; + `SubCopy`, `Eyebrow` with its coral tone), `Illustration` (hanging into the card below by `hang`, filling the free height above it unless `grows` is false, a tall one capped by `maxHeight`, never shrunk to fit), `Avatar`, `AvatarGrid` (the six keys `Profile.avatar_key` lists, each named for VoiceOver), `Card` (+ `CardTitle`; `split` for an illustration at its edge, `accent` the violet card on the yellow ground), `Chip` (with `asLabel`, the step label pill, `tone`, a Category's selection chip, `highlight`, the yellow pill naming the Member on the Recap cover, and `inverse`, the ink pill), `ProgressBar` (violet, green and coral, with the `piggy` marker drawn from `glyphs.tsx`'s shapes), `CategoryBar`, `Icon` (the eight Category glyphs drawn from `glyphs.tsx`'s shapes), `StatHero` (80, 56 or 40pt, measured), `StatChip` (earned, saved and spent), `StatChipRow` (chips that wrap to a new line), `PromptCard`, `ListRow` (a chevron, the check a chosen row carries, or nothing; its label coral when destructive), `MMRButton`, `MMRTextField`, `AmountField` (the Currency's symbol and decimals, the D57 cap, grouped as typed inside a `UITextField` delegate so a fast keystroke is never lost), `DayPicker` (1 to 31 and Last day; its dates are the caller's), `InlineError`, `Banner` (the coral Target Warning and the grey no-connection tone, with an optional ×), `TextLink`, the helper-text link, `BackRow`, the canvases' 44pt back chevron beside a label pill, `StreakBadge`, `StreakDots` (one per day of the window, the run lit), `CompareBars` (two or more bars scaled to their total or their largest, revealed by `Reveal`), `PageDots` (each dot a control in its own 44pt-tall slice, ink `onLight`), `ArchetypePill` and `ArchetypeRow` (the four pills, the month's picked out) and `CloseControl`, the Recap pager's 44pt close control, ink `onLight`, which the canvases draw and the React design system does not. Colours, radii and type come only from the tokens. The React `ChipRow` is a plain `HStack` here, and `PromptCard` has a `dashed` variant (the D47 outline card) the React one draws by hand. |
| `Resources/` | Asset catalog with the cover, envelope, jar and vending illustrations, the four Archetype mascots and the six Avatars at 3x, named by their contract keys (PNGs from `my-money-recap/assets/illustrations/`, `ill-cover.webp`, `ill-envelope.webp`, `ill-jar.webp`, `ill-vending.webp`, `ill-squirrel.webp`, `ill-hustler.webp`, `ill-treat.webp`, `ill-slowburner.webp` and `ill-avatar-<key>.webp`); `Fonts/` with Fredoka and Nunito and their licences. |
| `Tests/`, `UITests/` | XCTest unit tests on the view models, the magic-link handler and session (against a stub realm), the Keychain and the token refresh, with `Tests/Fixtures/home-zero.json` the recorded `GetHome` answer for the test Member after steps 1 and 2, in rupiah, `Tests/Fixtures/home-recaps.json` one recorded with two past months, a Hustler and a Partial (#130), `log-spend.json` a recorded `LogSpend` answer, `list-spends.json` a recorded `ListSpends` answer and `recap-*.json` `GetRecap` answers (a full Squirrel month, the month so far, a sign-up month, a month with no Money In, a one-day run, no run (a Hustler), an overspent month and a Slow Burner (#156), with `recap-maya-september.json` ledger's answer key for §8.6 rather than a recording, since the stack cannot close a September before 1 October; for screen 3, Maya's months in pounds: target met, missed, no Spend, no Saving Target, and the month so far before and after Payday); steps 1 and 2, Home's first-load skeleton, its rupiah proof (X3), its Target Warning banner before and after Payday and dismissed, its offline Banner with and without the warning, the Log a spend and Edit spend sheets, the Spend list with its empty months and Category detail, the Recap cover, C1, P1, its loading state, held sideways and in rupiah (X4), page 2, C2 and its variants, P2 and held sideways, screen 3, C3 in each state and P3, held sideways, and screen 4, C4 for each Archetype and without one and P4 in each situation, held sideways, drawn in a window (`NameAvatarViewTests`, `CurrencyViewTests`, `HomeViewTests`, `LogSpendViewTests`, `SpendListViewTests`, `RecapViewTests`, `RecapKeptTests`, `RecapGoalViewTests`, `RecapArchetypeTests`, all through `Picture.swift`); every listed currency checked against the backend's CLDR file (`CurrencyViewModelTests`, which reads it from the repository); the two UI tests above. |
| `scripts/` | `simulator-udid.sh`, `ui-test.sh`. |

The Fredoka and Nunito variable fonts sit in `Resources/Fonts` with their SIL Open Font License text, from
github.com/google/fonts, and `Info.plist` registers them (`UIAppFonts`), so the type ramp and the amount field's
measuring both use them (#106). Every screen works held sideways on an iPhone: a screen with no scroll view of its
own passes `Screen(scrolls: true)`, and steps 1 and 3 scroll their headline with their cards; the magic-link UI
test turns the phone at each screen and reaches its main control (#106). Each Recap page scrolls the same way
while swiping still turns the page (spec #151); the iPad column widths are left to the hardening slice.
