import XCTest

/// The sign-in-to-Home path against local Compose (IOS I9). Runs only when the realm test user's
/// password is passed in: `TEST_RUNNER_MMR_TEST_USER_PASSWORD=… xcodebuild test …`. CI skips it.
final class SignInUITests: XCTestCase {
    func testCancelThenSignInThroughKeycloakAndRelaunch() throws {
        guard let password = ProcessInfo.processInfo.environment["MMR_TEST_USER_PASSWORD"] else {
            throw XCTSkip("set MMR_TEST_USER_PASSWORD (infra/compose/.env TEST_USER_PASSWORD) to run against local Compose")
        }
        let app = XCUIApplication()
        app.launchArguments = ["--reset-session"]
        app.launch()
        XCTAssert(app.staticTexts["Your month in money."].waitForExistence(timeout: 10))
        snapshot("01-sign-in")

        // Cancel the system web sheet: InlineError inside the card, buttons re-enabled.
        app.buttons["Continue with email"].tap()
        try systemButton("Cancel", in: app).tap()
        XCTAssert(app.staticTexts["We couldn't sign you in. Try again."].waitForExistence(timeout: 5))
        XCTAssert(app.buttons["Continue with email"].isEnabled)
        snapshot("02-cancelled-inline-error")

        // The next tap clears the error and opens Keycloak's login page in the web sheet.
        app.buttons["Continue with email"].tap()
        try systemButton("Continue", in: app).tap()
        XCTAssertFalse(app.staticTexts["We couldn't sign you in. Try again."].exists)
        guard let username = first(hosts.map { $0.webViews.textFields.firstMatch }, attempts: 60) else {
            snapshot("03-keycloak-login-page-missing")
            return XCTFail("Keycloak login page did not load; is Compose up on localhost:8080?")
        }
        snapshot("03-keycloak-login-page")
        username.tap()
        username.typeText("test")
        let sheet = try XCTUnwrap(hosts.first { $0.webViews.firstMatch.exists })
        sheet.webViews.secureTextFields.firstMatch.tap()
        sheet.webViews.secureTextFields.firstMatch.typeText(password)
        sheet.webViews.buttons["Sign In"].tap()

        XCTAssert(app.staticTexts["You're signed in."].waitForExistence(timeout: 30))
        snapshot("04-signed-in")

        // Relaunch: the Keychain session skips Sign in.
        app.launchArguments = []
        app.terminate()
        app.launch()
        XCTAssert(app.staticTexts["You're signed in."].waitForExistence(timeout: 10))
        XCTAssertFalse(app.buttons["Continue with email"].exists)
        snapshot("05-relaunched-still-signed-in")
    }

    /// The web sheet and its "wants to use localhost to sign in" alert are drawn by system processes
    /// (SpringBoard, SafariViewService), not the app, so element queries look there first.
    private var hosts: [XCUIApplication] {
        ["com.apple.springboard", "com.apple.SafariViewService"].map(XCUIApplication.init(bundleIdentifier:)) + [XCUIApplication()]
    }

    private func systemButton(_ title: String, in app: XCUIApplication) throws -> XCUIElement {
        try XCTUnwrap(first(hosts.map { $0.buttons[title] }, attempts: 6), "no system button titled \(title)")
    }

    /// Polls each candidate for about a second per attempt and returns the first that exists.
    private func first(_ candidates: [XCUIElement], attempts: Int) -> XCUIElement? {
        for _ in 0..<attempts {
            for candidate in candidates where candidate.waitForExistence(timeout: 1) { return candidate }
        }
        return nil
    }

    private func snapshot(_ name: String) {
        let attachment = XCTAttachment(screenshot: XCUIScreen.main.screenshot())
        attachment.name = name
        attachment.lifetime = .keepAlways
        add(attachment)
    }
}
