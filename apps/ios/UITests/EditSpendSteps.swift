import XCTest

extension SignInUITests {
    /// PRD §6.10 / SCREEN-STATES B7, B8: the logged Spend opens Edit spend prefilled; a new amount and Category move
    /// Home and the list; the delete alert cancels, then deletes it, and the list and Home are back at nothing.
    func editAndDeleteTheSpend(_ app: XCUIApplication, from amount: Int, to edited: Int) {
        spentSoFar(app).tap()
        openTheSpend(app, in: "Snacks")
        let field = app.textFields["Amount"]
        XCTAssert(app.staticTexts["Edit spend"].waitForExistence(timeout: 10), "B7 opens titled Edit spend")
        XCTAssert(app.buttons["Save changes"].isEnabled, "prefilled, so Save changes is on")
        XCTAssert(app.textFields["Note"].value as? String == "Kopi susu", "the note prefilled")
        snapshot("17-edit-spend")
        field.coordinate(withNormalizedOffset: CGVector(dx: 0.95, dy: 0.5)).tap()
        field.typeText(String(repeating: XCUIKeyboardKey.delete.rawValue, count: String(amount).count + 3))
        field.typeText(String(edited))
        app.buttons.matching(NSPredicate(format: "label ==[c] 'Rides'")).firstMatch.tap()
        sideways(app, reaching: app.buttons["Delete this spend"], "edit-spend")
        app.buttons["Save changes"].tap()
        let rides = app.buttons.matching(NSPredicate(format: "label CONTAINS 'Rides' AND label CONTAINS '1 spend'"))
            .firstMatch
        XCTAssert(rides.waitForExistence(timeout: 30), "the edit moved the Spend to Rides and back to the list")
        XCTAssert(app.staticTexts["\(Self.rupiah(edited)) so far, 1 spend"].exists, "the list shows the new amount")
        snapshot("18-spend-list-after-an-edit")

        openTheSpend(app, in: "Rides")
        XCTAssert(app.buttons["Delete this spend"].waitForExistence(timeout: 10))
        app.buttons["Delete this spend"].tap()
        let alert = app.alerts["Delete this spend?"]
        XCTAssert(alert.waitForExistence(timeout: 10), "B8 the one confirm, an alert")
        XCTAssert(alert.staticTexts["You can't undo this."].exists)
        XCTAssert(alert.buttons["Cancel"].exists, "both buttons shown (owner, 2026-09-24)")
        snapshot("19-delete-confirm")
        alert.buttons["Cancel"].tap()
        XCTAssert(app.staticTexts["Edit spend"].waitForExistence(timeout: 10), "Cancel keeps the Spend and its sheet")
        app.buttons["Delete this spend"].tap()
        XCTAssert(alert.waitForExistence(timeout: 10))
        alert.buttons["Delete"].tap()
        let empty = app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH 'Nothing spent yet. Streak day '"))
            .firstMatch
        XCTAssert(empty.waitForExistence(timeout: 30), "the delete returns to the list with the row gone")
        snapshot("20-spend-list-after-a-delete")
        app.buttons["Back"].tap()
        XCTAssert(app.staticTexts[Self.zero("IDR")].waitForExistence(timeout: 10), "Home's hero is back at 0")
    }

    /// Opens the one Spend in `category` from the Spend list, by its note.
    private func openTheSpend(_ app: XCUIApplication, in category: String) {
        let row = app.buttons.matching(NSPredicate(format: "label CONTAINS %@ AND label CONTAINS '1 spend'", category))
            .firstMatch
        XCTAssert(row.waitForExistence(timeout: 30), "the list shows the Spend under \(category)")
        row.tap()
        let spend = app.buttons.matching(NSPredicate(format: "label CONTAINS 'Kopi susu'")).firstMatch
        XCTAssert(spend.waitForExistence(timeout: 10), "B6b lists the Spend by its note")
        spend.tap()
    }
}
