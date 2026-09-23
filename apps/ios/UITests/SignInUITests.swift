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
        sideways(app, reaching: app.buttons["Continue with email"], "sign-in")

        app.buttons["Continue with email"].tap()
        let field = app.textFields["Email address"]
        XCTAssert(field.waitForExistence(timeout: 5), "Continue with email pushes Enter email")
        XCTAssertFalse(app.buttons["Send me a link"].isEnabled, "Send stays disabled until the address is one")
        sideways(app, reaching: app.buttons["Send me a link"], "enter-email")
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
        sideways(app, reaching: countdown, "check-your-inbox")
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
        XCTAssertFalse(app.buttons["Back"].exists, "step 1 has no back chevron (SCREEN-STATES A4)")
        nameField.tap()
        nameField.typeText("Maya2")
        XCTAssertEqual(nameField.value as? String, "Maya", "the field refuses a digit (PRD §6.4)")
        answerStepOne(app, avatar: "Otter")
        backToStepOne(app, finding: "Maya", "Otter")
        app.buttons["Next"].tap()
        XCTAssert(app.staticTexts["Which currency?"].waitForExistence(timeout: 30), "Next returns to step 2")

        app.terminate()
        app.launch()
        XCTAssert(
            app.staticTexts["Which currency?"].waitForExistence(timeout: 30),
            "a force-quit on step 2 reopens step 2, the first unanswered (spec #92)"
        )
        backToStepOne(app, finding: "Maya", "Otter")
        snapshot("05-resumed-answers-restored")
        sideways(app, reaching: app.buttons["Next"], "step-1")
        app.buttons["Next"].tap()
        XCTAssert(app.staticTexts["Which currency?"].waitForExistence(timeout: 30))
        sideways(app, reaching: app.buttons["Next"], "step-2")

        let offline = "No connection. We need one to save your answers."
        app.launchEnvironment["MMR_BFF_URL"] = "http://localhost:1"
        app.terminate()
        app.launch()
        XCTAssert(
            app.staticTexts[offline].waitForExistence(timeout: 30),
            "part-way through setup with no signal: the named state, never Home (spec #92)"
        )
        app.buttons["Try again"].tap()
        let settled = expectation(
            for: NSPredicate(format: "isEnabled == true"),
            evaluatedWith: app.buttons["Try again"]
        )
        await fulfillment(of: [settled], timeout: 30)
        XCTAssert(app.staticTexts[offline].exists, "still no connection, still the state and its Try again")
        snapshot("06-no-connection-part-way-through-setup")
        sideways(app, reaching: app.buttons["Try again"], "no-connection")

        app.launchEnvironment = Self.appEnvironment
        app.terminate()
        app.launch()
        searchStepTwo(app, for: "rupiah", choosing: "Indonesian Rupiah, IDR")
        answerStepThree(app)
        answerStepFour(app, salary: 15_000_000)
        assertHomeAtZero(app, greeting: "Hi, Maya", currency: "IDR", salary: true, target: 3_750_000)
        snapshot("07-home-at-zero-signed-in-by-link")
        sideways(app, reaching: app.buttons["Log a spend"], "home")

        app.terminate()
        app.launch()
        XCTAssert(spentSoFar(app).waitForExistence(timeout: 10))
        XCTAssertFalse(
            app.buttons["Continue with email"].exists,
            "the Keychain session skips Sign in on relaunch (IOS I2)"
        )
        snapshot("08-relaunched-home-from-snapshot")

        app.launchEnvironment["MMR_BFF_URL"] = "http://localhost:1"
        app.terminate()
        app.launch()
        XCTAssert(spentSoFar(app).waitForExistence(timeout: 10))
        XCTAssert(
            app.staticTexts["Money in \(Self.zero("IDR"))"].exists,
            "a Member who finished setup reaches Home from the stored answer with no signal (PRD D61, spec #92)"
        )
        snapshot("09-bff-unreachable-snapshot-still-shown")
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
