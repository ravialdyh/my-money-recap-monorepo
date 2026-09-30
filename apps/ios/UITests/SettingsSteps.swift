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
        XCTAssert(app.buttons["Notifications"].exists && app.buttons["Export spends"].exists)
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
}
