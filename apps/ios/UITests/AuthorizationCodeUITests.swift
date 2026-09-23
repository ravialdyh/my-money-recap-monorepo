import XCTest

/// The authorization-code flow Apple and Google Members use, through Continue with email under the debug
/// launch argument: the PKCE request, the web sheet, the state from its authorization response (spec #61
/// seam 6). Needs the test Member's password.
@MainActor
final class AuthorizationCodeUITests: XCTestCase {
    override func setUp() {
        super.setUp()
        continueAfterFailure = false
    }

    func testCancelThenSignInThroughKeycloak() throws {
        guard let password = ProcessInfo.processInfo.environment["MMR_TEST_USER_PASSWORD"] else {
            throw XCTSkip(
                "set MMR_TEST_USER_PASSWORD (infra/compose/.env TEST_USER_PASSWORD) to run against local Compose"
            )
        }
        let app = XCUIApplication()
        app.launchEnvironment = Self.appEnvironment
        app.launchArguments = ["--reset-session", "--reset-snapshot", "--authorization-code-flow"]
        app.launch()
        XCTAssert(app.staticTexts["Your month in money."].waitForExistence(timeout: 10))

        app.buttons["Continue with email"].tap()
        try systemButton("Cancel").tap()
        XCTAssert(app.staticTexts["We couldn't sign you in. Try again."].waitForExistence(timeout: 5))
        XCTAssert(app.buttons["Continue with email"].isEnabled, "cancelling the web sheet re-enables the buttons")
        snapshot("01-cancelled-inline-error")

        app.buttons["Continue with email"].tap()
        try systemButton("Continue").tap()
        XCTAssertFalse(
            app.staticTexts["We couldn't sign you in. Try again."].exists,
            "the next tap clears the inline error"
        )
        guard let username = first(hosts.map(\.webViews.textFields.firstMatch), attempts: 60) else {
            snapshot("02-keycloak-login-page-missing")
            return XCTFail("Keycloak login page did not load; is Compose up?")
        }
        snapshot("02-keycloak-login-page")
        let sheet = try XCTUnwrap(hosts.first { $0.webViews.firstMatch.exists })
        let page = sheet.webViews.firstMatch
        if !isReauthenticating(sheet) {
            type("test", into: username, on: page)
        }
        let secret = sheet.webViews.secureTextFields.firstMatch
        type(password, into: secret, on: page)
        submit(secret, holding: password)

        let field = app.textFields["First name"]
        XCTAssert(field.waitForExistence(timeout: 30), "the runner cleared the test Member's row, so step 1 opens")
        XCTAssertEqual(field.value as? String, "Test", "the token's given_name fills the field (PRD §6.4)")
        answerStepOne(app, avatar: "Fox")
        let own = try XCTUnwrap(Locale.current.currency?.identifier, "the simulator's region names a currency")
        let name = Locale(identifier: "en_GB").localizedString(forCurrencyCode: own) ?? own
        XCTAssert(app.staticTexts["Which currency?"].waitForExistence(timeout: 30), "step 2 (PRD §6.5)")
        XCTAssert(app.buttons["\(name), \(own)"].isSelected, "the region's own currency starts chosen (PRD §6.5)")
        XCTAssert(app.buttons["Next"].isEnabled, "so the Member just taps Next")
        snapshot("03-step-2-own-currency")
        app.buttons["Next"].tap()
        XCTAssert(app.staticTexts["Your monthly salary"].waitForExistence(timeout: 30), "step 3 (PRD §6.6)")
        app.buttons["Skip for now"].tap()
        assertHomeAtZero(app, greeting: "Hi, Test", currency: own)
        snapshot("04-home-at-zero-through-the-web-sheet")
    }

    /// The web sheet and its "wants to use localhost to sign in" alert are drawn by system processes
    /// (SpringBoard, SafariViewService), not the app, so element queries look there first.
    private let hosts = ["com.apple.springboard", "com.apple.SafariViewService"]
        .map(XCUIApplication.init(bundleIdentifier:)) + [XCUIApplication()]

    /// Keycloak's SSO cookie outlives a run, and under `prompt=login` the next run inside its idle window gets the
    /// re-authentication form: the username fixed to the test Member, the focus already in the password field.
    private func isReauthenticating(_ sheet: XCUIApplication) -> Bool {
        sheet.webViews.buttons["Restart login"].waitForExistence(timeout: 2)
    }

    /// The web sheet drops keys typed while its software keyboard slides away, silently. Each attempt clears the field
    /// with ⌘A and ⌫, which also spend that slide, then types and waits for the field to read the text back.
    private func type(_ text: String, into field: XCUIElement, on page: XCUIElement) {
        focus(field, on: page)
        for attempt in 1 ... 3 {
            field.typeKey("a", modifierFlags: .command)
            field.typeKey(XCUIKeyboardKey.delete.rawValue, modifierFlags: [])
            field.typeText(text)
            if field.settles(on: text) {
                return
            }
            let typed = (field.value as? String ?? "").count
            XCTContext.runActivity(named: "attempt \(attempt) read back \(typed) of \(text.count)") { _ in }
        }
        XCTFail("\(field) never read back what was typed")
    }

    /// On iOS 26.0 the keyboard's toolbar covers the password field once the username is typed, so a tap there never
    /// focuses it. A tap that leaves the field unfocused is followed by one on the top of the page, putting the
    /// keyboard away.
    private func focus(_ field: XCUIElement, on page: XCUIElement) {
        let focused = NSPredicate(format: "hasKeyboardFocus == true")
        for attempt in 1 ... 3 {
            field.tap()
            let settled = XCTNSPredicateExpectation(predicate: focused, object: field)
            if XCTWaiter().wait(for: [settled], timeout: 5) == .completed {
                return
            }
            XCTContext.runActivity(named: "tap \(attempt) left the field unfocused; keyboard put away") { _ in }
            page.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.15)).tap()
        }
        XCTFail("\(field) never took the keyboard focus")
    }

    /// A lone Return is dropped the same way, and the Sign In button can sit under the keyboard's toolbar. So Return
    /// is sent again while the form, still holding the whole text, has not gone after ten seconds.
    private func submit(_ field: XCUIElement, holding text: String) {
        for attempt in 1 ... 3 {
            field.typeText("\n")
            let gone = XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: field)
            if XCTWaiter().wait(for: [gone], timeout: 10) == .completed {
                return
            }
            guard (field.value as? String ?? "").count == text.count else {
                return
            }
            XCTContext.runActivity(named: "Return \(attempt) left the form in place") { _ in }
        }
    }

    private func systemButton(_ title: String) throws -> XCUIElement {
        try XCTUnwrap(first(hosts.map { $0.buttons[title] }, attempts: 6), "no system button titled \(title)")
    }

    /// Polls each candidate for about a second per attempt and returns the first that exists.
    private func first(_ candidates: [XCUIElement], attempts: Int) -> XCUIElement? {
        for _ in 0 ..< attempts {
            for candidate in candidates where candidate.waitForExistence(timeout: 1) {
                return candidate
            }
        }
        return nil
    }
}

private extension XCUIElement {
    /// WebKit applies keys after `typeText` returns, so this polls for up to ten seconds. A password field reads back
    /// as one bullet per character, so only its length can be checked.
    func settles(on text: String) -> Bool {
        let predicate = elementType == .secureTextField
            ? NSPredicate(format: "value.length == %d", text.count)
            : NSPredicate(format: "value == %@", text)
        let settled = XCTNSPredicateExpectation(predicate: predicate, object: self)
        return XCTWaiter().wait(for: [settled], timeout: 10) == .completed
    }
}
