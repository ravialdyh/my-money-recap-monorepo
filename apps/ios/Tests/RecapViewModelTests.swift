import Connect
import MoneyRecapContract
import XCTest
@testable import MyMoneyRecap

/// Recorded `GetRecap` answers (`Fixtures/recap-*.json`, IOS I9) drive every case: which pages each situation draws,
/// the words page 1 prints, and the loading state kept when the BFF cannot answer.
@MainActor
final class RecapViewModelTests: XCTestCase {
    static func recorded(_ name: String) throws -> MoneyRecap_V1_RecapView {
        let url = try XCTUnwrap(Bundle(for: Self.self).url(forResource: name, withExtension: "json"))
        return try MoneyRecap_V1_BffServiceGetRecapResponse(jsonString: String(contentsOf: url, encoding: .utf8)).recap
    }

    static func loaded(_ name: String) async throws -> RecapViewModel {
        let answer = try Self.recorded(name)
        let model = RecapViewModel(
            month: answer.recap.month,
            fetch: { _ in answer },
            locale: Locale(identifier: "en_GB")
        )
        await model.load()
        return model
    }

    /// Budi's August, recorded from the stack with his sign-up, Salary and Saving Target moved to 1 August: the
    /// Squirrel.
    /// en-GB prints the rupiah's code, then a no-break space.
    func testAFullMonthDrawsTheCoverWithMoneyInSavedAndSpent() async throws {
        let model = try await Self.loaded("recap-squirrel")
        XCTAssertEqual(model.pages, [.cover, .kept, .goal, .archetype])
        XCTAssertEqual(model.page, 0, "a Recap opens on its first page")
        XCTAssertEqual(model.monthChip, "August 2026")
        XCTAssertEqual(model.nameChip, "Budi's recap")
        XCTAssertEqual(model.headline, "Your August in money")
        XCTAssertEqual(model.eyebrow, "Money that came your way")
        XCTAssertEqual(model.hero, "IDR\u{00A0}150,000,000")
        XCTAssertEqual(model.chips.map(\.text), ["Saved IDR\u{00A0}56,000,000", "Spent IDR\u{00A0}94,000,000"])
    }

    func testAnOverspentMonthPrintsItsNegativeSavedAsItIs() async throws {
        var answer = try Self.recorded("recap-squirrel")
        answer.recap.saved.amount = "-2000000"
        let model = RecapViewModel(month: "2026-08", fetch: { _ in answer }, locale: Locale(identifier: "en_GB"))
        await model.load()
        XCTAssertEqual(model.chips.first?.text, "Saved -IDR\u{00A0}2,000,000")
    }

    /// The three partial situations draw P1, headed by situation, Spent as the hero and never Money In (PRD §6.17):
    /// September so far, a sign-up on 27 August after its Payday, and an August with a Spend and no Salary.
    func testEachPartialSituationDrawsP1WithItsHeadline() async throws {
        let cases = [
            ("recap-preview", "September so far", "IDR\u{00A0}0", "No spends"),
            ("recap-sign-up", "Your first 5 days", "IDR\u{00A0}3,500,000", "2 spends"),
            ("recap-no-money-in", "Your August spending", "IDR\u{00A0}25,000", "1 spend"),
        ]
        for (name, headline, spent, count) in cases {
            let model = try await Self.loaded(name)
            XCTAssertEqual(model.pages, [.partialCover, .streak, .partialGoal, .closing], name)
            XCTAssertEqual(model.headline, headline, name)
            XCTAssertEqual(model.eyebrow, "What you've spent", name)
            XCTAssertEqual(model.hero, spent, name)
            XCTAssertEqual(model.chips.map(\.text), [count], name)
        }
    }

    /// §6.19: before any answer there is no page to turn, and with no answer the loading state stays.
    func testTheLoadingStateIsKeptOnAFailure() async {
        let model = RecapViewModel(
            month: "2026-08",
            fetch: { _ in throw ConnectError(code: .unavailable, message: nil) }
        )
        XCTAssertEqual(model.pages, [])
        await model.load()
        XCTAssertNil(model.answer)
        XCTAssertEqual(model.pages, [])
        XCTAssertEqual(model.chips.map(\.text), [" "], "one blank chip holds the card at its height")
    }
}
