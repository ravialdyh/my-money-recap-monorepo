import XCTest

extension SignInUITests {
    /// Spec #151, #153: before any Recap the dashed card under `Past recaps` opens the month so far on P1, with no
    /// Spend left after the delete; held sideways the close control is reachable, and it returns to Home.
    func openTheMonthSoFar(_ app: XCUIApplication) {
        XCTAssert(app.staticTexts["Past recaps"].exists, "the label over the dashed card")
        app.buttons.matching(NSPredicate(format: "label BEGINSWITH 'Your first recap arrives on '")).firstMatch.tap()
        let headline = app.staticTexts["\(Date().formatted(.dateTime.month(.wide))) so far"]
        XCTAssert(headline.waitForExistence(timeout: 30), "the dashed card opens P1 once GetRecap answers")
        XCTAssert(app.staticTexts["What you've spent"].exists)
        XCTAssert(app.staticTexts["No spends"].exists, "the deleted Spend is gone from the month")
        XCTAssert(app.staticTexts["Swipe to see where it all went."].exists)
        XCTAssertFalse(app.staticTexts["Money that came your way"].exists, "P1 shows no Money In")
        let dot = app.buttons["Page 1 of 1"]
        XCTAssert(dot.exists && dot.isSelected, "the page dot names its page")
        dot.tap()
        snapshot("21-recap-month-so-far")
        let close = app.buttons["Close"]
        sideways(app, reaching: close, "recap-month-so-far")
        close.tap()
        XCTAssert(spentSoFar(app).waitForExistence(timeout: 10), "Close returns to Home")
        XCTAssertFalse(headline.exists)
    }
}
