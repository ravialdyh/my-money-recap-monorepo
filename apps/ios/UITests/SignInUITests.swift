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
        answerStepFour(app, salary: 150_000_000)
        assertHomeAtZero(app, greeting: "Hi, Maya", currency: "IDR", salary: true, target: 37_500_000)
        snapshot("07-home-at-zero-signed-in-by-link")
        sideways(app, reaching: app.buttons["Log a spend"], "home")
        logASpend(app, rupiah: 25000)
        openTheSpendList(app, rupiah: 25000)
        editAndDeleteTheSpend(app, from: 25000, to: 40000)
        openTheMonthSoFar(app)
        renameInSettings(app, from: "Maya", to: "Tess")
        editTargetInSettings(app, salary: 150_000_000)
        convertToPoundsInSettings(app, salary: 150_000_000, rate: "0.000046", pounds: 6900)
        turnOnTheReminder(app)
        exportInSettings(app)

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
            app.staticTexts["Money in \(Self.zero("GBP"))"].exists,
            "a Member who finished setup reaches Home from the stored answer with no signal (PRD D61, spec #92)"
        )
        snapshot("09-bff-unreachable-snapshot-still-shown")
        app.buttons.matching(NSPredicate(format: "label BEGINSWITH 'Your first recap arrives on '")).firstMatch.tap()
        XCTAssert(
            app.staticTexts["\(Date().formatted(.dateTime.month(.wide))) so far"].waitForExistence(timeout: 10),
            "the dashed card opens the stored preview's P1 with no signal (PRD D61, IOS I4)"
        )
        XCTAssert(app.staticTexts["No spends"].exists, "the stored answer's figures")
        snapshot("26-recap-stored-no-signal")
    }
}

extension SignInUITests {
    /// PRD §6.9 / SCREEN-STATES B2 to B4: Save off while empty, the amount, one Category, today's date row and its
    /// picker, a note, the sheet sideways, then Save closes it, notifications are allowed (B5), and the hero shows the
    /// Spent the BFF answered.
    func logASpend(_ app: XCUIApplication, rupiah amount: Int) {
        app.buttons["Log a spend"].tap()
        let field = app.textFields["Amount"]
        XCTAssert(field.waitForExistence(timeout: 10), "Log a spend opens its sheet")
        XCTAssertFalse(app.buttons["Save"].isEnabled, "B2 empty: Save disabled")
        snapshot("10-log-a-spend-empty")
        field.tap()
        field.typeText(String(amount))
        let snacks = app.buttons.matching(NSPredicate(format: "label ==[c] 'Snacks'")).firstMatch
        snacks.tap()
        XCTAssert(snacks.isSelected, "the tapped Category reads as chosen")
        let date = app.buttons.matching(NSPredicate(format: "label BEGINSWITH 'Date' AND label CONTAINS 'Today, '"))
            .firstMatch
        XCTAssert(date.exists, "the date defaults to today")
        date.tap()
        XCTAssert(app.staticTexts["You can log up to 60 days back, never ahead."].waitForExistence(timeout: 5))
        XCTAssert(app.datePickers.firstMatch.exists, "B4 the wheel picker")
        snapshot("11-log-a-spend-date-picker")
        date.tap()
        let note = app.textFields["Note"]
        note.tap()
        note.typeText("Kopi susu\n")
        XCTAssert(app.buttons["Save"].isEnabled, "B3 filled")
        snapshot("12-log-a-spend-filled")
        sideways(app, reaching: app.buttons["Save"], "log-a-spend")
        app.buttons["Save"].tap()
        allowNotifications()
        XCTAssert(app.staticTexts[Self.rupiah(amount)].waitForExistence(timeout: 30), "the hero shows the new Spent")
        XCTAssertFalse(field.exists, "the sheet closed")
        snapshot("13-home-after-a-spend")
    }

    /// SCREEN-STATES B5: once the first Spend's sheet closes, iOS asks; the alert is springboard's, not the app's.
    func allowNotifications() {
        let allow = XCUIApplication(bundleIdentifier: "com.apple.springboard").buttons["Allow"]
        XCTAssert(allow.waitForExistence(timeout: 30), "the first Spend asks for notification permission (B5)")
        snapshot("30-push-permission")
        allow.tap()
        XCTAssertFalse(allow.waitForExistence(timeout: 2), "the alert closed")
    }
}

extension SignInUITests {
    /// PRD §6.10 / SCREEN-STATES B6, B6b: the hero opens the Spend list with the Spend under Snacks, its Category
    /// opens the detail naming it by its note, back returns to the list, and the month chip opens the switcher.
    func openTheSpendList(_ app: XCUIApplication, rupiah amount: Int) {
        spentSoFar(app).tap()
        let snacks = app.buttons.matching(NSPredicate(format: "label CONTAINS 'Snacks' AND label CONTAINS '1 spend'"))
            .firstMatch
        XCTAssert(snacks.waitForExistence(timeout: 30), "the hero opens the Spend list with the Spend under Snacks")
        XCTAssert(app.staticTexts["Spends"].exists)
        XCTAssert(app.staticTexts["\(Self.rupiah(amount)) so far, 1 spend"].exists, "the month line under the switcher")
        snapshot("14-spend-list")
        sideways(app, reaching: snacks, "spend-list")
        snacks.tap()
        XCTAssert(app.staticTexts["Kopi susu"].waitForExistence(timeout: 10), "B6b lists the Spend by its note")
        let share = NSPredicate(
            format: "label BEGINSWITH %@ AND label ENDSWITH ' so far'",
            "\(Self.rupiah(amount)) · 100% of "
        )
        XCTAssert(app.staticTexts.matching(share).firstMatch.exists, "the Category's share of this month")
        snapshot("15-category-detail")
        sideways(app, reaching: app.staticTexts["Kopi susu"], "category-detail")
        app.buttons["Back"].tap()
        XCTAssert(snacks.waitForExistence(timeout: 10), "back returns to the list, not Home")
        app.buttons["Back"].tap()
        XCTAssert(spentSoFar(app).waitForExistence(timeout: 10), "back from the list returns to Home")

        let now = Date()
        app.buttons[now.formatted(.dateTime.month(.wide))].tap()
        let shown = now.formatted(.dateTime.month(.wide).year())
        let current = app.buttons.matching(NSPredicate(format: "label == %@ AND selected == true", shown)).firstMatch
        XCTAssert(current.waitForExistence(timeout: 10), "the month chip opens the switcher, this month checked")
        snapshot("16-month-switcher")
        current.tap()
        XCTAssert(snacks.waitForExistence(timeout: 10))
        app.buttons["Back"].tap()
        XCTAssert(spentSoFar(app).waitForExistence(timeout: 10))
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
