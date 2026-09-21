import XCTest

/// The authorization-code flow Apple and Google Members use, through Continue with email under the debug
/// launch argument: the PKCE request, the web sheet, the state from its authorization response (spec #61
/// seam 6). Needs the test user's password.
@MainActor
final class AuthorizationCodeUITests: XCTestCase {
    func testCancelThenSignInThroughKeycloak() throws {
        guard let password = ProcessInfo.processInfo.environment["MMR_TEST_USER_PASSWORD"] else {
            throw XCTSkip(
                "set MMR_TEST_USER_PASSWORD (infra/compose/.env TEST_USER_PASSWORD) to run against local Compose"
            )
        }
        let app = XCUIApplication()
        app.launchEnvironment = ProcessInfo.processInfo.environment
            .filter { ["MMR_BFF_URL", "MMR_ISSUER"].contains($0.key) }
        app.launchArguments = ["--reset-session", "--reset-snapshot", "--authorization-code-flow"]
        app.launch()
        XCTAssert(app.staticTexts["Your month in money."].waitForExistence(timeout: 10))

        app.buttons["Continue with email"].tap()
        try systemButton("Cancel", in: app).tap()
        XCTAssert(app.staticTexts["We couldn't sign you in. Try again."].waitForExistence(timeout: 5))
        XCTAssert(app.buttons["Continue with email"].isEnabled, "cancelling the web sheet re-enables the buttons")
        snapshot("01-cancelled-inline-error")

        app.buttons["Continue with email"].tap()
        try systemButton("Continue", in: app).tap()
        XCTAssertFalse(
            app.staticTexts["We couldn't sign you in. Try again."].exists,
            "the next tap clears the inline error"
        )
        guard let username = first(hosts.map(\.webViews.textFields.firstMatch), attempts: 60) else {
            snapshot("02-keycloak-login-page-missing")
            return XCTFail("Keycloak login page did not load; is Compose up?")
        }
        snapshot("02-keycloak-login-page")
        username.tap()
        username.typeText("test")
        let sheet = try XCTUnwrap(hosts.first { $0.webViews.firstMatch.exists })
        sheet.webViews.secureTextFields.firstMatch.tap()
        sheet.webViews.secureTextFields.firstMatch.typeText(password)
        sheet.webViews.buttons["Sign In"].tap()

        assertHomeAtZero(app)
        snapshot("03-home-at-zero-through-the-web-sheet")
    }

    /// The web sheet and its "wants to use localhost to sign in" alert are drawn by system processes
    /// (SpringBoard, SafariViewService), not the app, so element queries look there first.
    private let hosts = ["com.apple.springboard", "com.apple.SafariViewService"]
        .map(XCUIApplication.init(bundleIdentifier:)) + [XCUIApplication()]

    private func systemButton(_ title: String, in _: XCUIApplication) throws -> XCUIElement {
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
