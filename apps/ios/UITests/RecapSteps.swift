import XCTest

extension SignInUITests {
    /// Spec #151, #153: before any Recap the dashed card under `Past recaps` opens the month so far on P1, with no
    /// Spend left after the delete; held sideways the close control is reachable, and it returns to Home.
    func openTheMonthSoFar(_ app: XCUIApplication) {
        let label = app.staticTexts.matching(NSPredicate(format: "label ==[c] 'Past recaps'")).firstMatch
        XCTAssert(label.waitForExistence(timeout: 10), "the label over the dashed card")
        app.buttons.matching(NSPredicate(format: "label BEGINSWITH 'Your first recap arrives on '")).firstMatch.tap()
        let headline = app.staticTexts["\(Date().formatted(.dateTime.month(.wide))) so far"]
        XCTAssert(headline.waitForExistence(timeout: 30), "the dashed card opens P1 once GetRecap answers")
        XCTAssert(app.staticTexts["What you've spent"].exists)
        XCTAssert(app.staticTexts["No spends"].exists, "the deleted Spend is gone from the month")
        XCTAssert(app.staticTexts["Swipe to see where it all went."].exists)
        XCTAssertFalse(app.staticTexts["Money that came your way"].exists, "P1 shows no Money In")
        let dot = app.buttons.matching(NSPredicate(format: "label BEGINSWITH 'Page 1 of '")).firstMatch
        XCTAssert(dot.exists && dot.isSelected, "the page dot names its page")
        dot.tap()
        snapshot("21-recap-month-so-far")
        let close = app.buttons["Close"]
        sideways(app, reaching: close, "recap-month-so-far")
        swipeToTheSpendingPage(app)
        sideways(app, reaching: close, "recap-spending")
        close.tap()
        XCTAssert(spentSoFar(app).waitForExistence(timeout: 10), "Close returns to Home")
        XCTAssertFalse(headline.exists)
    }

    /// #155: P3 with the Spend deleted: its chips, `Where {Spent} went`, nothing spent and no splurge, and the Payday
    /// card only before Payday, so it holds whichever day it runs.
    func swipeToTheSpendingPage(_ app: XCUIApplication) {
        let headline = app.staticTexts["Where \(Self.zero("IDR")) went"]
        var swipes = 0
        while !headline.waitForExistence(timeout: 2), swipes < 3 {
            app.swipeLeft()
            swipes += 1
        }
        XCTAssert(headline.exists, "a swipe reaches P3")
        let chip = { (text: String) in
            app.staticTexts.matching(NSPredicate(format: "label ==[c] %@", text)).firstMatch
        }
        XCTAssert(chip("Target").exists && chip("Spending").exists, "P3's chips")
        XCTAssert(app.staticTexts["Where it went"].exists)
        XCTAssert(app.staticTexts["Nothing spent this month"].exists, "the deleted Spend is gone")
        XCTAssertFalse(chip("The splurge").exists, "no splurge without a Spend")
        let payday = app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH 'Your salary lands on the '"))
            .firstMatch
        let day = Calendar.current.component(.day, from: .now) == 1 ? "2nd" : "1st"
        if payday.exists {
            XCTAssertEqual(payday.label, "Your salary lands on the \(day).", "the step-3 Payday, before it (D55)")
        }
        XCTAssertEqual(app.staticTexts["Saving target"].exists, payday.exists, "the target card only before Payday")
        let selected = app.buttons.matching(NSPredicate(format: "label BEGINSWITH 'Page ' AND selected == true"))
            .firstMatch
        XCTAssert(selected.label.hasPrefix("Page \(swipes + 1) of "), "the page dots follow the swipe")
        snapshot("22-recap-spending")
    }
}
