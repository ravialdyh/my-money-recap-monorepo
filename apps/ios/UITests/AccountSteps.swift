import XCTest

extension SignInUITests {
    func logOutAndBackIn(
        _ app: XCUIApplication, address: String, mailpit: Mailpit, after used: String
    ) async throws -> String {
        app.buttons["Settings"].tap()
        let row = app.buttons["Log out"]
        XCTAssert(row.waitForExistence(timeout: 30), "D1 lists Log out")
        row.tap()
        let sheet = app.sheets.firstMatch
        XCTAssert(sheet.waitForExistence(timeout: 10), "Log out asks once in the system action sheet (D12)")
        XCTAssert(sheet.staticTexts[
            "Log out of My Money Recap? You'll need your email or Apple or Google account to get back in."
        ].exists)
        snapshot("39-log-out-sheet")
        if sheet.buttons["Cancel"].exists {
            sheet.buttons["Cancel"].tap()
        } else {
            app.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.1)).tap()
        }
        let gone = expectation(for: NSPredicate(format: "exists == false"), evaluatedWith: sheet)
        await fulfillment(of: [gone], timeout: 10)
        XCTAssert(row.exists, "Cancel, a tap outside where iOS draws the sheet as a popover, keeps the Member on D1")
        row.tap()
        XCTAssert(sheet.waitForExistence(timeout: 10))
        sheet.buttons["Log out"].tap()
        XCTAssert(app.staticTexts["Your month in money."].waitForExistence(timeout: 30), "Sign in, with no message")
        snapshot("40-logged-out")
        return try await signInByLink(app, address: address, mailpit: mailpit, after: used, landing: spentSoFar(app))
    }

    func deleteThenKeep(_ app: XCUIApplication, address: String, mailpit: Mailpit, after used: String) async throws {
        app.buttons["Settings"].tap()
        let row = app.buttons["Delete account"]
        XCTAssert(row.waitForExistence(timeout: 30), "D1 lists Delete account")
        row.tap()
        XCTAssert(app.staticTexts["This is permanent"].waitForExistence(timeout: 10), "the row opens D10")
        snapshot("41-delete-account")
        let next = app.buttons["Delete my account"]
        sideways(app, reaching: next, "delete-account")
        next.tap()
        let field = app.textFields["Type DELETE to confirm"]
        XCTAssert(field.waitForExistence(timeout: 10), "D10 moves to D11")
        let delete = app.buttons["Delete my account"]
        XCTAssertFalse(delete.isEnabled, "nothing typed, nothing to delete")
        field.tap()
        field.typeText("DELET")
        XCTAssertFalse(delete.isEnabled, "only exactly DELETE")
        field.typeText("E")
        field.typeText("\n")
        XCTAssert(delete.isEnabled)
        snapshot("42-delete-confirm-typed")
        sideways(app, reaching: delete, "delete-confirm")
        delete.tap()
        XCTAssert(app.staticTexts["Your month in money."].waitForExistence(timeout: 30), "a granted request signs out")
        _ = try await signInByLink(
            app, address: address, mailpit: mailpit, after: used, landing: app.staticTexts["Welcome back"]
        )
        snapshot("43-welcome-back")
        app.buttons["Keep my account"].tap()
        XCTAssert(spentSoFar(app).waitForExistence(timeout: 30), "Keep my account keeps everything and reaches Home")
    }

    private func signInByLink(
        _ app: XCUIApplication, address: String, mailpit: Mailpit, after used: String, landing: XCUIElement
    ) async throws -> String {
        app.buttons["Continue with email"].tap()
        let field = app.textFields["Email address"]
        XCTAssert(field.waitForExistence(timeout: 5))
        field.tap()
        field.typeText(address)
        field.typeText("\n")
        XCTAssert(app.staticTexts["Check your inbox."].waitForExistence(timeout: 15), "RequestMagicLink answered")
        let link = try await mailpit.link(to: address, after: used)
        app.terminate()
        app.launchArguments = ["--magic-link", link]
        app.launch()
        XCTAssert(landing.waitForExistence(timeout: 30), "the new link signs the same Member in")
        app.launchArguments = []
        return link
    }
}

struct Stack {
    let mailpit: Mailpit

    init() throws {
        guard let mailpit = ProcessInfo.processInfo.environment["MMR_MAILPIT_URL"].flatMap(URL.init(string:)) else {
            throw XCTSkip("set MMR_MAILPIT_URL (the stack's Mailpit, infra/compose/README.md) to run against Compose")
        }
        self.mailpit = Mailpit(base: mailpit)
    }
}

struct Mailpit {
    let base: URL

    func link(to address: String, after used: String? = nil) async throws -> String {
        for _ in 0 ..< 20 {
            if let id = try await newestMessage(to: address) {
                let message = try await json(base.appending(path: "api/v1/message/\(id)"))
                let text = message["Text"] as? String ?? ""
                if let link = text.range(of: #"\S+\?lt=\S+"#, options: .regularExpression), String(text[link]) != used {
                    return String(text[link])
                }
            }
            try await Task.sleep(for: .milliseconds(500))
        }
        throw MailpitError.noLink(address)
    }

    private func newestMessage(to address: String) async throws -> String? {
        let search = base.appending(path: "api/v1/search").appending(queryItems: [
            URLQueryItem(name: "query", value: "to:\(address) subject:\"sign-in link\""),
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
