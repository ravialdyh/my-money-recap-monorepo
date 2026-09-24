import Connect
import Foundation
import MoneyRecapContract
import SwiftData
import XCTest
@testable import MyMoneyRecap

/// Log a spend (PRD §6.9) from the recorded answer `Fixtures/log-spend.json`: Save waits for an amount and a Category,
/// a failure keeps every value, a retry keeps its key, and the date and note stay inside what ledger takes.
@MainActor
final class LogSpendViewModelTests: XCTestCase {
    private var calendar: Calendar {
        var gregorian = Calendar(identifier: .gregorian)
        gregorian.timeZone = TimeZone(identifier: "Asia/Jakarta") ?? .gmt
        return gregorian
    }

    static func recordedAnswer() throws -> MoneyRecap_V1_BffServiceLogSpendResponse {
        let url = try XCTUnwrap(Bundle(for: Self.self).url(forResource: "log-spend", withExtension: "json"))
        return try MoneyRecap_V1_BffServiceLogSpendResponse(jsonString: String(contentsOf: url, encoding: .utf8))
    }

    private func sheet(send: @escaping LogSpendViewModel.Send) throws -> LogSpendViewModel {
        let today = try XCTUnwrap(calendar.date(from: DateComponents(
            year: 2026,
            month: 9,
            day: 26,
            hour: 23,
            minute: 30
        )))
        return LogSpendViewModel(
            currency: "IDR", digits: 0, earliest: "2026-07-29", today: today, calendar: calendar,
            locale: Locale(identifier: "en_GB"), send: send
        )
    }

    func testSaveWaitsForAnAmountAboveZeroAndACategory() throws {
        let model = try sheet { _, _ in .failure(ConnectError(code: .unavailable, message: nil)) }
        XCTAssertFalse(model.canSave, "B2 empty: Save disabled")
        model.amount = "25000"
        XCTAssertFalse(model.canSave, "no Category yet")
        model.amount = "0"
        model.choose(.snacks)
        XCTAssertFalse(model.canSave, "zero is no Spend")
        model.amount = "25000."
        XCTAssertTrue(model.canSave, "B3 filled")
    }

    func testAFailureKeepsTheValuesAndARetryKeepsItsKey() async throws {
        let recorded = try Self.recordedAnswer()
        var sent: [(MoneyRecap_V1_BffServiceLogSpendRequest, String)] = []
        var answer: Result<MoneyRecap_V1_BffServiceLogSpendResponse, ConnectError> =
            .failure(ConnectError(code: .internalError, message: nil))
        let model = try sheet { request, headers in
            sent.append((request, headers["Idempotency-Key"]?.first ?? ""))
            return answer
        }
        model.amount = "25000."
        model.choose(.rides)
        model.type("Kopi susu")
        let failed = await model.save()
        XCTAssertNil(failed)
        XCTAssertTrue(model.failed, "the one failure line (B2 write failed)")
        XCTAssertEqual(LogSpendViewModel.failure, "We couldn't save this spend. Try again.")
        XCTAssertTrue(model.canSave, "Save stays enabled")
        XCTAssertEqual(model.amount, "25000.", "the values are intact")
        XCTAssertEqual(model.category, .rides)
        answer = .success(recorded)
        let home = await model.save()
        XCTAssertEqual(home, recorded.home, "the answer's Home replaces the snapshot")
        XCTAssertFalse(model.failed)
        XCTAssertEqual(sent.count, 2)
        XCTAssertEqual(sent[0].1, sent[1].1, "a retry of the same values sends the same key (IOS I3)")
        let request = sent[1].0
        XCTAssertEqual(request.amount.amount, "25000", "a trailing point is not sent")
        XCTAssertEqual(request.amount.currencyCode, "IDR")
        XCTAssertEqual(request.category, .rides)
        XCTAssertEqual(request.date, "2026-09-26", "the device's own day, at 23:30 in Jakarta (D34)")
        XCTAssertEqual(request.note, "Kopi susu")
        model.amount = "30000"
        _ = await model.save()
        XCTAssertNotEqual(sent[2].1, sent[1].1, "changed values start a new attempt")
    }

    /// The picker runs from the earliest date ledger answered to today (D72); the phone computes neither end.
    func testTheDatesRunFromLedgersEarliestToToday() throws {
        let model = try sheet { _, _ in .failure(ConnectError(code: .unavailable, message: nil)) }
        XCTAssertEqual(model.dateLabel, "Today, 26 September")
        model.date = model.dates.lowerBound
        XCTAssertEqual(model.dateLabel, "29 July", "ledger's earliest date")
        XCTAssertEqual(LogSpendViewModel.dateHelper, "You can log up to 60 days back, never ahead.")
        let unknown = LogSpendViewModel(currency: "IDR", digits: 0, earliest: "") { _, _ in
            .failure(ConnectError(code: .unavailable, message: nil))
        }
        XCTAssertEqual(unknown.dates.lowerBound, unknown.dates.upperBound, "today alone before Home names a date")
    }

    /// PRD §6.9: at most 80 characters, counted as ledger counts them, code points of the NFC form.
    func testTheNoteStopsAtEightyCharacters() throws {
        let model = try sheet { _, _ in .failure(ConnectError(code: .unavailable, message: nil)) }
        model.type(String(repeating: "e\u{301}", count: 81))
        XCTAssertEqual(model.note.unicodeScalars.count, 80, "each decomposed é counts once")
        XCTAssertEqual(model.note, String(repeating: "\u{e9}", count: 80))
    }

    func testTheEightCategoriesFollowTheContractsOrder() {
        XCTAssertEqual(
            Icon.Name.allCases.map(LogSpendViewModel.contract),
            [.snacks, .games, .toys, .rides, .clothes, .bills, .gifts, .other]
        )
    }

    /// Home opens the sheet in its Currency and decimals, and the save's Home replaces the stored one whole (IOS I4).
    func testHomeOpensTheSheetAndStoresTheAnswer() async throws {
        let recorded = try Self.recordedAnswer()
        let context = try ModelContext(ModelContainer(
            for: HomeSnapshot.self, SpendListSnapshot.self,
            configurations: ModelConfiguration(isStoredInMemoryOnly: true)
        ))
        var gbp = recorded.home
        gbp.currencyCode = "GBP"
        gbp.currencyDigits = 2
        gbp.earliestSpendDate = "2026-09-01"
        let home = HomeViewModel(fetch: { gbp }, context: context, logSpend: { _, _ in .success(recorded) })
        XCTAssertNil(home.logSheet(), "no sheet before Home has an answer")
        await home.refresh()
        let sheet = try XCTUnwrap(home.logSheet())
        XCTAssertEqual(sheet.currency, "GBP")
        XCTAssertEqual(sheet.digits, 2)
        let iso = Date.ISO8601FormatStyle(timeZone: .current).year().month().day()
        XCTAssertEqual(sheet.dates.lowerBound.formatted(iso), "2026-09-01", "ledger's earliest date (D72)")
        sheet.amount = "12.40"
        sheet.choose(.snacks)
        let saved = await sheet.save()
        let answer = try XCTUnwrap(saved)
        home.replace(with: answer)
        XCTAssertEqual(home.home?.spent.amount, recorded.home.spent.amount)
        let stored = try XCTUnwrap(HomeSnapshot.load(from: context))
        XCTAssertEqual(try MoneyRecap_V1_HomeView(serializedBytes: stored.bytes), recorded.home)
    }
}
