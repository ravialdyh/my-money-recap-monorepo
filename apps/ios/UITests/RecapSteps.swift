import XCTest

extension SignInUITests {
    /// Spec #151, #153, #154: before any Recap the dashed card under `Past recaps` opens the month so far on P1, with
    /// no Spend left after the delete; swipes turn to P2 and P3, each turned sideways reaching Close, which returns
    /// Home.
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
        let dot = onScreen(app.buttons, "Page 1 of 3")
        XCTAssert(dot?.isSelected == true, "the page dot names its page")
        dot?.tap()
        snapshot("21-recap-month-so-far")
        let close = app.buttons["Close"]
        sideways(app, reaching: close, "recap-month-so-far")
        app.swipeLeft()
        let streak = app.staticTexts["Your longest no-spend streak"]
        let turned = XCTNSPredicateExpectation(predicate: NSPredicate(format: "hittable == true"), object: streak)
        XCTAssertEqual(XCTWaiter().wait(for: [turned], timeout: 10), .completed, "a swipe turns to P2")
        XCTAssert(onScreen(app.buttons, "Page 2 of 3")?.isSelected == true, "the second dot is lit")
        XCTAssertFalse(app.staticTexts["Kept against your target"].exists, "P2 has no bars card")
        snapshot("22-recap-streak")
        sideways(app, reaching: close, "recap-streak")
        swipeToTheSpendingPage(app, from: 2)
        sideways(app, reaching: close, "recap-spending")
        close.tap()
        XCTAssert(spentSoFar(app).waitForExistence(timeout: 10), "Close returns to Home")
        XCTAssertFalse(streak.exists)
    }

    /// The one element labelled `label` on screen: the pager keeps its neighbouring page, dots included, in the tree.
    private func onScreen(_ query: XCUIElementQuery, _ label: String) -> XCUIElement? {
        query.matching(NSPredicate(format: "label == %@", label)).allElementsBoundByIndex.first { $0.isHittable }
    }

    /// #155: P3 with the Spend deleted: its chips, `Where {Spent} went`, nothing spent and no splurge, and the Payday
    /// card only before Payday, so it holds whichever day it runs.
    func swipeToTheSpendingPage(_ app: XCUIApplication, from page: Int) {
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
        XCTAssert(selected.label.hasPrefix("Page \(page + swipes) of "), "the page dots follow the swipe")
        snapshot("23-recap-spending")
    }
}
