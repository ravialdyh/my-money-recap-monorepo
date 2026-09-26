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

    private func loaded(_ name: String) async throws -> RecapViewModel {
        let answer = try Self.recorded(name)
        let model = RecapViewModel(
            month: answer.recap.month,
            fetch: { _ in answer },
            locale: Locale(identifier: "en_GB")
        )
        await model.load()
        return model
    }

    func testAFullMonthDrawsTheCoverWithMoneyInSavedAndSpent() async throws {
        let model = try await loaded("recap-squirrel")
        XCTAssertEqual(model.pages, [.cover])
        XCTAssertEqual(model.page, 0, "a Recap opens on its first page")
        XCTAssertEqual(model.monthChip, "September 2026")
        XCTAssertEqual(model.nameChip, "Maya's recap")
        XCTAssertEqual(model.headline, "Your September in money")
        XCTAssertEqual(model.eyebrow, "Money that came your way")
        XCTAssertEqual(model.hero, "£2,400")
        XCTAssertEqual(model.chips.map(\.text), ["Saved £640", "Spent £1,760"])
    }

    func testAnOverspentMonthPrintsItsNegativeSavedAsItIs() async throws {
        var answer = try Self.recorded("recap-squirrel")
        answer.recap.saved.amount = "-200.00"
        let model = RecapViewModel(month: "2026-09", fetch: { _ in answer }, locale: Locale(identifier: "en_GB"))
        await model.load()
        XCTAssertEqual(model.chips.first?.text, "Saved -£200")
    }

    /// The three partial situations draw P1, headed by situation, Spent as the hero and never Money In (PRD §6.17).
    func testEachPartialSituationDrawsP1WithItsHeadline() async throws {
        let cases = [
            ("recap-preview", "September so far", "£0", "No spends"),
            ("recap-sign-up", "Your first 4 days", "£200", "4 spends"),
            ("recap-no-money-in", "Your July spending", "£70", "1 spend"),
        ]
        for (name, headline, spent, count) in cases {
            let model = try await loaded(name)
            XCTAssertEqual(model.pages, [.partialCover], name)
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
