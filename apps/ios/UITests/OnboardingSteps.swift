import XCTest

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
        sideways(app, reaching: amount, "step-3")
        app.buttons["Next"].tap()
    }

    /// Home's hero eyebrow, "Spent so far", with the countdown to the Salary once one is set (PRD §6.8).
    @MainActor
    func spentSoFar(_ app: XCUIApplication) -> XCUIElement {
        app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH 'Spent so far'")).firstMatch
    }

    /// Step 4 (PRD §6.7, SCREEN-STATES A7): back to step 3 and forward, 20% by default with its amount, the knob
    /// dragged
    /// to 25% and the preview following it, then Done.
    @MainActor
    func answerStepFour(_ app: XCUIApplication, salary: Int) {
        XCTAssert(app.staticTexts["How much do you want to keep?"].waitForExistence(timeout: 30), "step 4 (PRD §6.7)")
        app.buttons["Back"].tap()
        XCTAssert(app.staticTexts["Your monthly salary"].waitForExistence(timeout: 10), "Back returns to step 3")
        app.buttons["Next"].tap()
        XCTAssert(app.staticTexts["How much do you want to keep?"].waitForExistence(timeout: 30), "and Next to step 4")
        XCTAssert(app.staticTexts[Self.preview(salary / 5, of: salary)].exists, "20% by default (PRD §6.7)")
        let slider = app.sliders["Saving target"]
        slider.coordinate(withNormalizedOffset: CGVector(dx: 0.2, dy: 0.5))
            .press(forDuration: 0.1, thenDragTo: slider.coordinate(withNormalizedOffset: CGVector(dx: 0.25, dy: 0.5)))
        XCTAssert(app.staticTexts[Self.preview(salary / 4, of: salary)].waitForExistence(timeout: 5), "live at 25%")
        snapshot("step-4-answered")
        sideways(app, reaching: app.buttons["Done"], "step-4")
        app.buttons["Done"].tap()
    }

    /// "{share} of {salary} each month" in rupiah, as the app prints a whole amount (`Money`).
    static func preview(_ share: Int, of salary: Int) -> String {
        "\(rupiah(share)) of \(rupiah(salary)) each month"
    }

    static func rupiah(_ amount: Int) -> String {
        Decimal(amount).formatted(.currency(code: "IDR").precision(.fractionLength(0)))
    }

    /// Home at zero (PRD §6.8), greeting the Member and printing every figure in the Currency step 2 saved: the B1
    /// "no salary set" layout; with `salary` the Salary prompt gone; with a `target` amount its card filled in.
    @MainActor
    func assertHomeAtZero(
        _ app: XCUIApplication,
        greeting: String,
        currency: String,
        salary: Bool = false,
        target: Int? = nil
    ) {
        XCTAssert(spentSoFar(app).waitForExistence(timeout: 30), "Home at zero (PRD §6.8)")
        XCTAssert(app.staticTexts[Self.zero(currency)].exists, "the hero shows Spent at \(Self.zero(currency))")
        XCTAssert(app.staticTexts["Money in \(Self.zero(currency))"].exists)
        XCTAssertEqual(
            app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH 'Saved'")).count,
            0,
            "no Saved chip while Money In is 0 (D41)"
        )
        XCTAssertEqual(app.buttons["Set your monthly salary"].exists, !salary, "the prompt only without a Salary (D54)")
        XCTAssertEqual(app.buttons["Set a saving target"].exists, target == nil, "the prompt only without a Target")
        if let target {
            XCTAssert(app.staticTexts["\(Self.zero(currency)) of \(Self.rupiah(target))"].exists, "Saved of the Target")
            XCTAssert(
                app.staticTexts
                    .matching(NSPredicate(format: "label BEGINSWITH 'Nothing to keep yet. Your salary lands on the '"))
                    .firstMatch.exists,
                "before the first Payday, the salary-lands line (spec #92)"
            )
        }
        XCTAssert(app.buttons["Log a spend"].exists)
        XCTAssertFalse(app.buttons["See last month"].isEnabled)
        XCTAssert(app.buttons.matching(NSPredicate(format: "label BEGINSWITH 'Your first recap arrives on'")).firstMatch
            .exists)
        XCTAssert(app.staticTexts[greeting].exists, "Home greets the Member by name (PRD §6.8)")
    }

    /// PRD §11 "Platform": the screen held sideways, `control` reached by scrolling where the screen scrolls, a
    /// picture, then upright again (#106).
    @MainActor
    func sideways(_ app: XCUIApplication, reaching control: XCUIElement, _ name: String) {
        turn(app, to: .landscapeLeft, name)
        var swipes = 0
        while !control.isHittable, swipes < 4 {
            app.swipeUp()
            swipes += 1
        }
        XCTAssert(control.isHittable, "\(name): reachable held sideways")
        snapshot("sideways-\(name)")
        turn(app, to: .portrait, name)
    }

    /// Turns the phone and waits until the window has turned too, then for the rotation to settle.
    @MainActor
    private func turn(_ app: XCUIApplication, to orientation: UIDeviceOrientation, _ name: String) {
        XCUIDevice.shared.orientation = orientation
        let window = app.windows.firstMatch
        let wide = orientation.isLandscape
        let turned = XCTNSPredicateExpectation(
            predicate: NSPredicate { _, _ in (window.frame.width > window.frame.height) == wide },
            object: nil
        )
        XCTAssertEqual(XCTWaiter().wait(for: [turned], timeout: 10), .completed, "\(name): turned")
        _ = XCTWaiter().wait(for: [XCTestExpectation(description: "the rotation settles")], timeout: 1)
    }

    @MainActor
    func snapshot(_ name: String) {
        let attachment = XCTAttachment(screenshot: XCUIScreen.main.screenshot())
        attachment.name = name
        attachment.lifetime = .keepAlways
        add(attachment)
    }
}
