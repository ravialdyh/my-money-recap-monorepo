import MoneyRecapContract
import XCTest
@testable import MyMoneyRecap

/// Screen 3's words from recorded `GetRecap` answers (IOS I9): Maya's months in pounds, recorded from the stack on
/// 26 September 2026 (UTC), the day `today` stands for, and the partial months of `RecapViewModelTests`.
@MainActor
final class RecapGoalViewModelTests: XCTestCase {
    static let recordedOn = Calendar.current.date(from: DateComponents(year: 2026, month: 9, day: 26, hour: 12))!

    static func loaded(_ answer: MoneyRecap_V1_RecapView) async -> RecapViewModel {
        let model = RecapViewModel(
            month: answer.recap.month,
            fetch: { _ in answer },
            locale: Locale(identifier: "en_GB"),
            today: recordedOn
        )
        await model.load()
        return model
    }

    private func loaded(_ name: String) async throws -> RecapViewModel {
        try await Self.loaded(RecapViewModelTests.recorded(name))
    }

    /// Canvas C3: June, Saved £640 of a £480 target, eight Categories, the splurge with its note.
    func testTargetMetDrawsThreeBarsTheRollUpAndTheSplurge() async throws {
        let model = try await loaded("recap-target-met")
        XCTAssertEqual(model.pages, [.cover, .kept, .goal])
        XCTAssertEqual(model.goalHeadline, "133% of your saving target")
        XCTAssertEqual(model.target?.progress, "£640 of £480")
        XCTAssertEqual(model.answer?.recap.savingTarget.progressSentence, "Target met. £160 to spare.")
        XCTAssertNil(model.paydayLine)
        XCTAssertEqual(model.spendingEyebrow, "Where £1,760 went")
        XCTAssertEqual(model.spendingTitle, "Mostly bills")
        XCTAssertEqual(model.bars.map(\.category), [.bills, .rides, .snacks])
        XCTAssertEqual(model.bars.map(\.percent), [35, 26, 14])
        XCTAssertEqual(model.bars.map(\.amount), ["£616", "£458", "£246"])
        XCTAssertEqual(
            model.rollUp,
            "Clothes, Gifts, Games, Toys, Other · 25%",
            "10 + 7 + 5 + 2 + 1, ledger's rounding"
        )
        XCTAssertFalse(model.everyPenny)
        XCTAssertEqual(model.splurge, "£420 on a flight home. Your biggest single buy.")
        XCTAssertEqual(model.bars.first?.spoken, "Bills, 35 percent, £616")
    }

    /// October's canvas in July: 65%, a part-full bar, the splurge with no note named by its Category.
    func testTargetMissedDrawsAPartFullBarAndASplurgeWithNoNote() async throws {
        let model = try await loaded("recap-target-missed")
        XCTAssertEqual(model.goalHeadline, "65% of your saving target")
        XCTAssertEqual(model.target?.progress, "£310 of £480")
        XCTAssertEqual(try XCTUnwrap(model.target?.fraction), 310.0 / 480, accuracy: 0.0001)
        XCTAssertEqual(model.spendingTitle, "Mostly bills")
        XCTAssertEqual(model.rollUp, "Rides, Games, Other · 27%", "the canvas's fourth bar joins the roll-up (D71)")
        XCTAssertEqual(model.splurge, "£560 on clothes. Your biggest single buy.")
    }

    func testNoSpendKeepsEveryPennyAndDrawsNoSplurge() async throws {
        let model = try await loaded("recap-no-spends")
        XCTAssertEqual(model.goalHeadline, "500% of your saving target")
        XCTAssertEqual(model.spendingEyebrow, "Where it went")
        XCTAssertEqual(model.spendingTitle, "Nothing spent this month")
        XCTAssertTrue(model.bars.isEmpty)
        XCTAssertNil(model.rollUp)
        XCTAssertTrue(model.everyPenny, "Money In is above 0")
        XCTAssertNil(model.splurge)
    }

    /// May, before the Saving Target's first month: no percentage anywhere, the target card instead (PRD §6.13).
    func testNoSavingTargetHeadsWithSpentAndDrawsTheTargetCard() async throws {
        let model = try await loaded("recap-no-target")
        XCTAssertEqual(model.pages, [.cover, .kept, .goal])
        XCTAssertEqual(model.goalHeadline, "Where £1,760 went")
        XCTAssertNil(model.target)
        XCTAssertNil(model.paydayLine)
        XCTAssertFalse(model.partial, "the page draws the target card")
        XCTAssertEqual(model.splurge, "£420 on a flight home. Your biggest single buy.")
    }

    /// P3 (PRD §6.17, D55): September so far, Payday moved to the 30th, 4 days after the recording.
    func testTheMonthSoFarBeforePaydayNamesThePayday() async throws {
        let model = try await loaded("recap-before-payday")
        XCTAssertEqual(model.pages, [.partialCover, .streak, .partialGoal])
        XCTAssertEqual(model.goalHeadline, "Where £900 went", "P3 has no percentage")
        XCTAssertNil(model.target)
        XCTAssertEqual(model.paydayLine, "Your salary lands on the 30th.")
        XCTAssertEqual(model.bars.map(\.percent), [56, 20, 18])
        XCTAssertEqual(model.rollUp, "Rides, Other · 6%")
        XCTAssertEqual(model.splurge, "£280 on phone and internet. Your biggest single buy.")
    }

    /// After Payday, and in both past partial months, P3 has no target card.
    func testP3AfterPaydayAndInAPastPartialMonthHasNoTargetCard() async throws {
        for name in ["recap-after-payday", "recap-sign-up", "recap-no-money-in"] {
            let model = try await loaded(name)
            XCTAssertEqual(model.pages, [.partialCover, .streak, .partialGoal], name)
            XCTAssertNil(model.target, name)
            XCTAssertNil(model.paydayLine, name)
        }
        let sign = try await loaded("recap-sign-up")
        XCTAssertEqual(sign.goalHeadline, "Where IDR\u{00A0}350,000 went")
        XCTAssertEqual(sign.rollUp, nil, "two Categories, no roll-up")
    }

    /// "You kept every penny that came in." only while Money In is above 0.
    func testNoSpendAndNoMoneyInKeepsNoPenny() async throws {
        var answer = try RecapViewModelTests.recorded("recap-preview")
        let paid = await Self.loaded(answer)
        XCTAssertTrue(paid.everyPenny)
        answer.recap.moneyIn.amount = "0"
        let unpaid = await Self.loaded(answer)
        XCTAssertFalse(unpaid.everyPenny)
    }
}
