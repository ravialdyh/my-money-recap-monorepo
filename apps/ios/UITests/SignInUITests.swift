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
        app.buttons["Next"].tap()
        XCTAssert(app.staticTexts["Which currency?"].waitForExistence(timeout: 30))

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

        app.launchEnvironment = Self.appEnvironment
        app.terminate()
        app.launch()
        searchStepTwo(app, for: "rupiah", choosing: "Indonesian Rupiah, IDR")
        answerStepThree(app)
        assertHomeAtZero(app, greeting: "Hi, Maya", currency: "IDR", salary: true)
        snapshot("07-home-at-zero-signed-in-by-link")

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

extension XCTestCase {
    /// The runner's `MMR_BFF_URL` and `MMR_ISSUER`, the two the app reads, handed on at launch.
    static var appEnvironment: [String: String] {
        ProcessInfo.processInfo.environment.filter { ["MMR_BFF_URL", "MMR_ISSUER"].contains($0.key) }
    }

    /// 0 in `currency`, as the app prints a whole amount (`Money`): in the device locale, without a zero fraction.
    static func zero(_ currency: String) -> String {
        Decimal(0).formatted(.currency(code: currency).precision(.fractionLength(0)))
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

    /// Step 2's back chevron (spec #92): step 1 again, its saved name in the field and its Avatar re-selected.
    @MainActor
    func backToStepOne(_ app: XCUIApplication, finding name: String, _ avatar: String) {
        XCTAssert(app.staticTexts["Which currency?"].waitForExistence(timeout: 30), "step 2 (PRD §6.5)")
        app.buttons["Back"].tap()
        let field = app.textFields["First name"]
        XCTAssert(field.waitForExistence(timeout: 10), "Back returns to step 1")
        XCTAssertEqual(field.value as? String, name, "the saved name")
        XCTAssert(app.buttons[avatar].isSelected, "the saved Avatar, re-selected")
        XCTAssert(app.buttons["Next"].isEnabled)
    }

    /// Onboarding step 2 (PRD §6.5, SCREEN-STATES A5): a search that matches nothing says so and clears, a search by
    /// name finds the Currency, and Next saves it.
    @MainActor
    func searchStepTwo(_ app: XCUIApplication, for query: String, choosing row: String) {
        XCTAssert(app.staticTexts["Which currency?"].waitForExistence(timeout: 30), "step 2 (PRD §6.5)")
        XCTAssert(app.staticTexts.matching(NSPredicate(format: "label ==[c] 'Step 2 of 4'")).firstMatch.exists)
        let search = app.textFields["Search by name or code"]
        search.tap()
        search.typeText("ducat")
        XCTAssert(app.staticTexts["No currency matches \"ducat\"."].waitForExistence(timeout: 5), "the empty search")
        snapshot("step-2-no-match")
        app.buttons["Clear search"].tap()
        search.tap()
        search.typeText(query)
        app.buttons[row].tap()
        XCTAssert(app.buttons[row].isSelected, "the tapped Currency reads as chosen")
        snapshot("step-2-searched")
        app.buttons["Next"].tap()
    }

    /// Step 3 (PRD §6.6, SCREEN-STATES A6): back to step 2 and forward, Next off until an amount, a 12th digit
    /// refused with the D57 line, and a Payday not today's, so the sign-up month shows Money In 0 (PRD §8.11).
    @MainActor
    func answerStepThree(_ app: XCUIApplication) {
        XCTAssert(app.staticTexts["Your monthly salary"].waitForExistence(timeout: 30), "step 3 (PRD §6.6)")
        XCTAssert(app.staticTexts.matching(NSPredicate(format: "label ==[c] 'Step 3 of 4'")).firstMatch.exists)
        app.buttons["Back"].tap()
        XCTAssert(app.staticTexts["Which currency?"].waitForExistence(timeout: 10), "Back returns to step 2")
        XCTAssert(app.buttons["Indonesian Rupiah, IDR"].isSelected, "the saved Currency, re-selected")
        app.buttons["Next"].tap()
        XCTAssert(app.staticTexts["Your monthly salary"].waitForExistence(timeout: 30), "Next returns to step 3")
        XCTAssertFalse(app.buttons["Next"].isEnabled, "Next stays off until an amount is typed")
        let amount = app.textFields["Amount, after tax"]
        amount.tap()
        amount.typeText("999999999999")
        XCTAssertEqual(amount.value as? String, Int64(99_999_999_999).formatted(), "the 12th digit is refused")
        XCTAssert(app.staticTexts["That's more than the app can store."].exists, "SCREEN-STATES A6 amount too long")
        snapshot("step-3-amount-too-long")
        amount.typeText(String(repeating: XCUIKeyboardKey.delete.rawValue, count: 11) + "15000000")
        XCTAssertEqual(amount.value as? String, Int64(15_000_000).formatted(), "grouped as the device writes it")
        let payday = Calendar.current.component(.day, from: .now) == 1 ? "2" : "1"
        app.buttons[payday].tap()
        XCTAssert(app.buttons[payday].isSelected, "the tapped Payday reads as chosen")
        XCTAssert(app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH 'Next paydays: '")).firstMatch.exists)
        snapshot("step-3-answered")
        app.buttons["Next"].tap()
    }

    /// Home's hero eyebrow, "Spent so far", with the countdown to the Salary once one is set (PRD §6.8).
    @MainActor
    func spentSoFar(_ app: XCUIApplication) -> XCUIElement {
        app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH 'Spent so far'")).firstMatch
    }

    /// Home at zero (PRD §6.8), greeting the Member and printing every figure in the Currency step 2 saved: the B1
    /// "no salary set" layout, or with `salary` the Salary prompt gone and the target prompt still there.
    @MainActor
    func assertHomeAtZero(_ app: XCUIApplication, greeting: String, currency: String, salary: Bool = false) {
        XCTAssert(spentSoFar(app).waitForExistence(timeout: 30), "Home at zero (PRD §6.8)")
        XCTAssert(app.staticTexts[Self.zero(currency)].exists, "the hero shows Spent at \(Self.zero(currency))")
        XCTAssert(app.staticTexts["Money in \(Self.zero(currency))"].exists)
        XCTAssertEqual(
            app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH 'Saved'")).count,
            0,
            "no Saved chip while Money In is 0 (D41)"
        )
        XCTAssertEqual(app.buttons["Set your monthly salary"].exists, !salary, "the prompt only without a Salary (D54)")
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
