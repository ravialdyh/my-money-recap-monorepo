import XCTest

/// The sign-in-to-Home path against local Compose (IOS I9). Runs only when the realm test user's
/// password is passed in: `TEST_RUNNER_MMR_TEST_USER_PASSWORD=… xcodebuild test …`. CI skips it.
final class SignInUITests: XCTestCase {
    func testCancelThenSignInThroughKeycloakAndRelaunch() throws {
        guard let password = ProcessInfo.processInfo.environment["MMR_TEST_USER_PASSWORD"] else {
            throw XCTSkip(
                "set MMR_TEST_USER_PASSWORD (infra/compose/.env TEST_USER_PASSWORD) to run against local Compose"
            )
        }
        let app = XCUIApplication()
        app.launchArguments = ["--reset-session", "--reset-snapshot"]
        app.launch()
        XCTAssert(app.staticTexts["Your month in money."].waitForExistence(timeout: 10))
        snapshot("01-sign-in")

        app.buttons["Continue with email"].tap()
        try systemButton("Cancel", in: app).tap()
        XCTAssert(app.staticTexts["We couldn't sign you in. Try again."].waitForExistence(timeout: 5))
        XCTAssert(app.buttons["Continue with email"].isEnabled, "cancelling the web sheet re-enables the buttons")
        snapshot("02-cancelled-inline-error")

        app.buttons["Continue with email"].tap()
        try systemButton("Continue", in: app).tap()
        XCTAssertFalse(
            app.staticTexts["We couldn't sign you in. Try again."].exists,
            "the next tap clears the inline error"
        )
        guard let username = first(hosts.map(\.webViews.textFields.firstMatch), attempts: 60) else {
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

        let zero = Decimal(0).formatted(.currency(code: "GBP").precision(.fractionLength(0)))
        XCTAssert(
            app.staticTexts["Spent so far"].waitForExistence(timeout: 30),
            "Home at zero, the B1 \"no salary set\" layout (PRD §6.8)"
        )
        XCTAssert(app.staticTexts[zero].exists, "the hero shows Spent at \(zero)")
        XCTAssert(app.staticTexts["Money in \(zero)"].exists)
        XCTAssertEqual(
            app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH 'Saved'")).count,
            0,
            "no Saved chip while Money In is 0 (D41)"
        )
        XCTAssert(app.buttons["Set your monthly salary"].exists)
        XCTAssert(app.buttons["Set a saving target"].exists)
        XCTAssert(app.buttons["Log a spend"].exists)
        XCTAssertFalse(app.buttons["See last month"].isEnabled)
        XCTAssert(app.buttons.matching(NSPredicate(format: "label BEGINSWITH 'Your first recap arrives on'")).firstMatch
            .exists)
        XCTAssertFalse(
            app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH 'Hi,'")).firstMatch.exists,
            "no greeting while display_name is empty"
        )
        snapshot("04-home-at-zero")

        app.launchArguments = []
        app.terminate()
        app.launch()
        XCTAssert(app.staticTexts["Spent so far"].waitForExistence(timeout: 10))
        XCTAssertFalse(
            app.buttons["Continue with email"].exists,
            "the Keychain session skips Sign in on relaunch (IOS I2)"
        )
        snapshot("05-relaunched-home-from-snapshot")

        app.launchEnvironment["MMR_BFF_URL"] = "http://localhost:1"
        app.terminate()
        app.launch()
        XCTAssert(app.staticTexts["Spent so far"].waitForExistence(timeout: 10))
        XCTAssert(
            app.staticTexts["Money in \(zero)"].exists,
            "an unreachable BFF leaves the last snapshot on screen (PRD D61)"
        )
        snapshot("06-bff-unreachable-snapshot-still-shown")

        app.launchArguments = ["--reset-snapshot"]
        app.terminate()
        app.launch()
        XCTAssert(app.otherElements["home-skeleton"].waitForExistence(timeout: 10))
        XCTAssertEqual(
            app.activityIndicators.count,
            0,
            "no snapshot and no answer is the §6.19 skeleton, never a spinner"
        )
        XCTAssertFalse(app.staticTexts["Spent so far"].exists)
        snapshot("07-first-load-skeleton")
    }

    /// The web sheet and its "wants to use localhost to sign in" alert are drawn by system processes
    /// (SpringBoard, SafariViewService), not the app, so element queries look there first.
    private let hosts = ["com.apple.springboard", "com.apple.SafariViewService"]
        .map(XCUIApplication.init(bundleIdentifier:)) + [XCUIApplication()]

    private func systemButton(_ title: String, in app: XCUIApplication) throws -> XCUIElement {
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

    private func snapshot(_ name: String) {
        let attachment = XCTAttachment(screenshot: XCUIScreen.main.screenshot())
        attachment.name = name
        attachment.lifetime = .keepAlways
        add(attachment)
    }
}
