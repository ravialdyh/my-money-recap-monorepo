import XCTest

extension SignInUITests {
    /// #170: the Settings chip opens D1, each row read with its value; Profile renames the Member and Save returns to
    /// D1 with the new name; D1 sideways reaches Log out; back closes it and Home greets the new name.
    func renameInSettings(_ app: XCUIApplication, from old: String, to new: String) {
        app.buttons["Settings"].tap()
        XCTAssert(app.staticTexts["Settings"].waitForExistence(timeout: 10), "the chip opens D1")
        let row = { (label: String, value: String) in
            app.buttons.matching(NSPredicate(format: "label BEGINSWITH %@ AND label CONTAINS %@", label, value))
                .firstMatch
        }
        XCTAssert(row("Profile", old).waitForExistence(timeout: 30), "Profile reads the Display Name")
        XCTAssert(row("Currency", "IDR").exists)
        let day = Calendar.current.component(.day, from: .now) == 1 ? "2nd" : "1st"
        XCTAssert(row("Salary and payday", " · \(day)").exists, "the Salary with its Payday")
        XCTAssert(row("Saving target", "25%").exists)
        XCTAssert(row("Notifications", "2 on").exists && app.buttons["Export spends"].exists)
        XCTAssert(app.buttons["Delete account"].exists && app.buttons["Log out"].exists)
        snapshot("27-settings")
        sideways(app, reaching: app.buttons["Log out"], "settings")

        row("Profile", old).tap()
        let field = app.textFields["First name"]
        XCTAssert(field.waitForExistence(timeout: 10), "Profile opens D2 with setup's field")
        XCTAssertEqual(field.value as? String, old)
        XCTAssertFalse(app.buttons["Save"].isEnabled, "nothing changed yet")
        field.tap()
        field.typeText(String(repeating: XCUIKeyboardKey.delete.rawValue, count: old.count) + new)
        let typed = XCTNSPredicateExpectation(predicate: NSPredicate(format: "value == %@", new), object: field)
        XCTAssertEqual(XCTWaiter().wait(for: [typed], timeout: 5), .completed, "the new name typed")
        snapshot("28-profile-renamed")
        sideways(app, reaching: app.buttons["Save"], "profile")
        app.buttons["Save"].tap()
        XCTAssert(row("Profile", new).waitForExistence(timeout: 30), "Save returns to D1, which reads the new name")
        app.buttons["Back"].tap()
        XCTAssert(app.staticTexts["Hi, \(new)"].waitForExistence(timeout: 30), "Home greets the new name")
        snapshot("29-home-greets-new-name")
    }

    /// #172: Saving target opens D4 over the setup Target, 25%, with the forward-only line, since the setup save was
    /// the first; the slider moves the preview, D4 sideways reaches Save, and Save returns to D1, where 25% stays in
    /// force this month (D65).
    func editTargetInSettings(_ app: XCUIApplication, salary: Int) {
        app.buttons["Settings"].tap()
        let row = app.buttons.matching(NSPredicate(format: "label BEGINSWITH 'Saving target' AND label CONTAINS '25%'"))
            .firstMatch
        XCTAssert(row.waitForExistence(timeout: 30), "the row reads the setup Target")
        row.tap()
        XCTAssert(app.staticTexts["Share of salary to keep"].waitForExistence(timeout: 10), "D4 opens setup's card")
        XCTAssert(app.staticTexts[Self.preview(salary / 4, of: salary)].exists, "over the Target in force")
        let next = Calendar.current.date(byAdding: .month, value: 1, to: .now) ?? .now
        XCTAssert(app.staticTexts["Your new target starts in \(next.formatted(.dateTime.month(.wide)))."].exists)
        let slider = app.sliders["Saving target"]
        slider.coordinate(withNormalizedOffset: CGVector(dx: 0.25, dy: 0.5))
            .press(forDuration: 0.1, thenDragTo: slider.coordinate(withNormalizedOffset: CGVector(dx: 0.3, dy: 0.5)))
        XCTAssert(
            app.staticTexts[Self.preview(salary * 3 / 10, of: salary)].waitForExistence(timeout: 5),
            "live at 30%"
        )
        snapshot("30-saving-target-edit")
        sideways(app, reaching: app.buttons["Save"], "saving-target")
        app.buttons["Save"].tap()
        XCTAssert(row.waitForExistence(timeout: 30), "Save returns to D1, 25% still in force this month")
        XCTAssertFalse(app.staticTexts["Share of salary to keep"].exists)
        app.buttons["Back"].tap()
        XCTAssert(spentSoFar(app).waitForExistence(timeout: 10), "back to Home")
    }

    /// #174: D5 with the rupiah checked; the pound meets the failed fetch (no rates App ID) twice; D7's typed rate
    /// previews the Salary; Convert returns to D5 on the pound; D1 and Home read pounds; the month so far is stored
    /// again.
    func convertToPoundsInSettings(_ app: XCUIApplication, salary: Int, rate: String, pounds: Decimal) {
        app.buttons["Settings"].tap()
        let row = { (label: String, value: String) in
            app.buttons.matching(NSPredicate(format: "label BEGINSWITH %@ AND label CONTAINS %@", label, value))
                .firstMatch
        }
        XCTAssert(row("Currency", "IDR").waitForExistence(timeout: 30), "the row reads the rupiah")
        row("Currency", "IDR").tap()
        let search = app.textFields["Search by name or code"]
        XCTAssert(search.waitForExistence(timeout: 10), "Currency opens D5 with setup's search")
        XCTAssert(app.buttons["Indonesian Rupiah, IDR"].isSelected, "the Currency held is checked")
        snapshot("31-currency")
        sideways(app, reaching: search, "currency")
        search.tap()
        search.typeText("pound")
        app.buttons["British Pound, GBP"].tap()
        let failed = app.staticTexts["We couldn't get today's rate."]
        XCTAssert(failed.waitForExistence(timeout: 30), "no rates App ID: the fetch fails (D6)")
        XCTAssert(app.staticTexts["Convert everything to GBP?"].exists)
        let enter = app.buttons["Enter it myself"]
        XCTAssert(enter.exists && app.buttons["Cancel"].exists)
        snapshot("32-currency-fetch-failed")
        sideways(app, reaching: enter, "currency-fetch-failed")
        app.buttons["Try again"].tap()
        XCTAssert(enter.waitForExistence(timeout: 30) && failed.exists, "Try again meets the same failure")
        enter.tap()
        let field = app.textFields["Rate: 1 IDR ="]
        XCTAssert(field.waitForExistence(timeout: 10), "Enter it myself opens D7")
        XCTAssert(app.staticTexts["We couldn't fetch this rate. Enter it yourself to continue."].exists)
        let untyped = "All your amounts will be converted at the rate you enter. This can't be undone."
        XCTAssert(app.staticTexts[untyped].exists)
        let convert = app.buttons["Convert"]
        XCTAssertFalse(convert.isEnabled, "no rate yet")
        snapshot("33-currency-manual-rate")
        field.tap()
        field.typeText(rate)
        let after = Self.pounds(pounds)
        let sentence = "Your \(Self.rupiah(salary)) salary becomes \(after). This can't be undone."
        XCTAssert(app.staticTexts[sentence].waitForExistence(timeout: 30), "the preview answers after typing stops")
        XCTAssert(row("Salary", after).exists, "the Salary row, before and after")
        XCTAssert(convert.isEnabled)
        snapshot("34-currency-manual-rate-typed")
        sideways(app, reaching: convert, "currency-manual-rate")
        convert.tap()
        let pound = app.buttons.matching(NSPredicate(format: "label == 'British Pound, GBP' AND selected == true"))
        XCTAssert(pound.firstMatch.waitForExistence(timeout: 30), "Convert returns to D5 on the pound")
        app.buttons["Back"].tap()
        XCTAssert(row("Currency", "GBP").waitForExistence(timeout: 10), "D1 reads the answer's Settings")
        XCTAssert(row("Salary and payday", "\(after) · ").exists)
        app.buttons["Back"].tap()
        XCTAssert(app.staticTexts["Money in \(Self.zero("GBP"))"].waitForExistence(timeout: 30), "Home in pounds")
        XCTAssert(app.staticTexts["\(Self.zero("GBP")) of \(Self.pounds(pounds / 4))"].exists, "the 25% Target")
        snapshot("35-home-in-pounds")
        sideways(app, reaching: app.buttons["Log a spend"], "home-in-pounds")
        app.buttons.matching(NSPredicate(format: "label BEGINSWITH 'Your first recap arrives on '")).firstMatch.tap()
        let cover = app.staticTexts["\(Date().formatted(.dateTime.month(.wide))) so far"]
        XCTAssert(cover.waitForExistence(timeout: 30), "the erased Recap is asked again, now in pounds")
        app.buttons["Close"].tap()
        XCTAssert(spentSoFar(app).waitForExistence(timeout: 10))
    }

    /// An amount in pounds as the app prints it (`Money`): no fraction when whole, else two decimals.
    static func pounds(_ amount: Decimal) -> String {
        var source = amount, rounded = Decimal()
        NSDecimalRound(&rounded, &source, 0, .plain)
        let whole = rounded == amount
        return amount.formatted(.currency(code: "GBP").precision(.fractionLength(whole ? 0 : 2)))
    }

    /// #179: with notifications allowed after the first Spend, D1's row counts PRD §10's two defaults; D8 turns the
    /// Daily reminder on, which grows its `Remind me at` row and saves at once, and D1 then counts three.
    func turnOnTheReminder(_ app: XCUIApplication) {
        app.buttons["Settings"].tap()
        let row = { (value: String) in
            app.buttons.matching(NSPredicate(format: "label BEGINSWITH 'Notifications' AND label CONTAINS %@", value))
                .firstMatch
        }
        XCTAssert(row("2 on").waitForExistence(timeout: 30), "Recap ready and Target warning are on by default")
        row("2 on").tap()
        let reminder = app.switches["Daily reminder"]
        XCTAssert(reminder.waitForExistence(timeout: 10), "the row opens D8")
        XCTAssert(app.switches["Recap ready"].exists && app.switches["Target warning"].exists)
        XCTAssertFalse(app.staticTexts["Remind me at"].exists, "no time row while the reminder is off")
        reminder.tap()
        XCTAssert(app.staticTexts["Remind me at"].waitForExistence(timeout: 10), "the reminder's time row")
        XCTAssertFalse(app.staticTexts["We couldn't save this. Try again."].waitForExistence(timeout: 3), "saved")
        snapshot("36-notifications-reminder-on")
        sideways(app, reaching: app.switches["Target warning"], "notifications")
        app.buttons["Back"].tap()
        XCTAssert(row("3 on").waitForExistence(timeout: 30), "D1 asks again and counts the reminder")
        app.buttons["Back"].tap()
        XCTAssert(app.buttons["Log a spend"].waitForExistence(timeout: 30), "back on Home")
    }
}
