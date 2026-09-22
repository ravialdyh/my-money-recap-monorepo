import XCTest

/// The sign-in-to-Home path against local Compose (IOS I9), by magic link (spec #61 seam 6). Skips unless the
/// runner names the stack's Mailpit, as `scripts/ui-test.sh` does.
@MainActor
final class SignInUITests: XCTestCase {
    override func setUp() {
        super.setUp()
        continueAfterFailure = false
    }

    func testSignInByMagicLinkThenRelaunch() async throws {
        let stack = try Stack()
        let address = "mmr-test-\(UUID().uuidString.lowercased())@example.test"
        let app = XCUIApplication()
        app.launchEnvironment = Self.appEnvironment
        app.launchArguments = ["--reset-session", "--reset-snapshot"]
        app.launch()
        XCTAssert(app.staticTexts["Your month in money."].waitForExistence(timeout: 10))
        snapshot("01-sign-in")

        app.buttons["Continue with email"].tap()
        let field = app.textFields["Email address"]
        XCTAssert(field.waitForExistence(timeout: 5), "Continue with email pushes Enter email")
        XCTAssertFalse(app.buttons["Send me a link"].isEnabled, "Send stays disabled until the address is one")
        field.tap()
        field.typeText(address)
        XCTAssert(app.buttons["Send me a link"].isEnabled)
        snapshot("02-enter-email-typed")

        field.typeText("\n")
        XCTAssert(app.staticTexts["Check your inbox."].waitForExistence(timeout: 15), "RequestMagicLink answered")
        XCTAssert(app.staticTexts["We sent a link to \(address). It works for 15 minutes."].exists)
        let countdown = app.buttons.matching(NSPredicate(format: "label BEGINSWITH 'Resend in 0:'")).firstMatch
        XCTAssert(countdown.exists, "Resend sleeps behind a live m:ss countdown")
        XCTAssertFalse(countdown.isEnabled)
        snapshot("03-check-your-inbox-countdown")
        let first = countdown.label
        try await Task.sleep(for: .seconds(3))
        XCTAssertNotEqual(countdown.label, first, "the countdown is live")
        snapshot("04-check-your-inbox-countdown-later")

        let link = try await stack.mailpit.link(to: address)
        app.terminate()
        app.launchArguments = ["--magic-link", link]
        app.launch()
        XCTAssert(
            app.staticTexts["What should we call you?"].waitForExistence(timeout: 30),
            "a new Member meets step 1"
        )
        app.launchArguments = []
        app.terminate()
        app.launch()
        let nameField = app.textFields["First name"]
        XCTAssert(nameField.waitForExistence(timeout: 30), "a force-quit on step 1 reopens step 1")
        nameField.tap()
        nameField.typeText("Maya2")
        XCTAssertEqual(nameField.value as? String, "Maya", "the field refuses a digit (PRD §6.4)")
        answerStepOne(app, avatar: "Otter")
        assertHomeAtZero(app, greeting: "Hi, Maya")
        snapshot("05-home-at-zero-signed-in-by-link")

        app.terminate()
        app.launch()
        XCTAssert(app.staticTexts["Spent so far"].waitForExistence(timeout: 10))
        XCTAssertFalse(
            app.buttons["Continue with email"].exists,
            "the Keychain session skips Sign in on relaunch (IOS I2)"
        )
        snapshot("06-relaunched-home-from-snapshot")

        app.launchEnvironment["MMR_BFF_URL"] = "http://localhost:1"
        app.terminate()
        app.launch()
        XCTAssert(app.staticTexts["Spent so far"].waitForExistence(timeout: 10))
        XCTAssert(
            app.staticTexts["Money in \(Self.zero)"].exists,
            "an unreachable BFF leaves the last snapshot on screen (PRD D61)"
        )
        snapshot("07-bff-unreachable-snapshot-still-shown")

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
        snapshot("08-first-load-skeleton")
    }
}

extension XCTestCase {
    /// The runner's `MMR_BFF_URL` and `MMR_ISSUER`, the two the app reads, handed on at launch.
    static var appEnvironment: [String: String] {
        ProcessInfo.processInfo.environment.filter { ["MMR_BFF_URL", "MMR_ISSUER"].contains($0.key) }
    }

    static var zero: String {
        Decimal(0).formatted(.currency(code: "GBP").precision(.fractionLength(0)))
    }

    /// Onboarding step 1 (PRD §6.4), its name already in the field: no Next without an Avatar, then the Avatar and
    /// Next.
    @MainActor
    func answerStepOne(_ app: XCUIApplication, avatar: String) {
        XCTAssert(app.staticTexts["What should we call you?"].waitForExistence(timeout: 30), "step 1 (PRD §6.4)")
        XCTAssert(app.staticTexts.matching(NSPredicate(format: "label ==[c] 'Step 1 of 4'")).firstMatch.exists)
        XCTAssertFalse(app.buttons["Next"].isEnabled, "Next stays off until an Avatar is picked")
        app.buttons[avatar].tap()
        XCTAssert(app.buttons[avatar].isSelected, "the picked Avatar reads as selected")
        XCTAssert(app.buttons["Next"].isEnabled)
        snapshot("step-1-answered")
        app.buttons["Next"].tap()
    }

    /// Home at zero, the B1 "no salary set" layout (PRD §6.8), greeting the Member once step 1 is saved.
    @MainActor
    func assertHomeAtZero(_ app: XCUIApplication, greeting: String) {
        XCTAssert(app.staticTexts["Spent so far"].waitForExistence(timeout: 30), "Home at zero (PRD §6.8)")
        XCTAssert(app.staticTexts[Self.zero].exists, "the hero shows Spent at \(Self.zero)")
        XCTAssert(app.staticTexts["Money in \(Self.zero)"].exists)
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
        XCTAssert(app.staticTexts[greeting].exists, "Home greets the Member by name (PRD §6.8)")
    }

    @MainActor
    func snapshot(_ name: String) {
        let attachment = XCTAttachment(screenshot: XCUIScreen.main.screenshot())
        attachment.name = name
        attachment.lifetime = .keepAlways
        add(attachment)
    }
}

/// The Compose stack the runner points at (`scripts/ui-test.sh` reads it from `infra/compose/.env`). The
/// app gets the BFF and the issuer; Mailpit is read from this process, never from the app.
struct Stack {
    let mailpit: Mailpit

    init() throws {
        guard let mailpit = ProcessInfo.processInfo.environment["MMR_MAILPIT_URL"].flatMap(URL.init(string:)) else {
            throw XCTSkip("set MMR_MAILPIT_URL (the stack's Mailpit, infra/compose/README.md) to run against Compose")
        }
        self.mailpit = Mailpit(base: mailpit)
    }
}

/// Mailpit's HTTP API: the newest message to an address, and the magic link inside it.
struct Mailpit {
    let base: URL

    func link(to address: String) async throws -> String {
        for _ in 0 ..< 20 {
            if let id = try await newestMessage(to: address) {
                let message = try await json(base.appending(path: "api/v1/message/\(id)"))
                let text = message["Text"] as? String ?? ""
                if let link = text.range(of: #"\S+\?lt=\S+"#, options: .regularExpression) {
                    return String(text[link])
                }
            }
            try await Task.sleep(for: .milliseconds(500))
        }
        throw MailpitError.noLink(address)
    }

    private func newestMessage(to address: String) async throws -> String? {
        let search = base.appending(path: "api/v1/search").appending(queryItems: [
            URLQueryItem(name: "query", value: "to:\(address)"),
        ])
        let messages = try await json(search)["messages"] as? [[String: Any]]
        return messages?.first?["ID"] as? String
    }

    private func json(_ url: URL) async throws -> [String: Any] {
        let (data, _) = try await URLSession.shared.data(from: url)
        return try JSONSerialization.jsonObject(with: data) as? [String: Any] ?? [:]
    }
}

enum MailpitError: Error {
    case noLink(String)
}
