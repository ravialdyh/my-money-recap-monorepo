import Connect
import Foundation
import MoneyRecapContract
import XCTest
@testable import MyMoneyRecap

/// Step 3 (PRD §6.6, spec #92): Next waits for an amount above zero and saves it with the Payday and "this month",
/// one key per answer; Skip for now records step 3 alone; the helper shows the three paydays from today on.
@MainActor
final class SalaryViewModelTests: XCTestCase {
    private var calendar: Calendar {
        var gregorian = Calendar(identifier: .gregorian)
        gregorian.timeZone = TimeZone(identifier: "Asia/Jakarta") ?? .gmt
        return gregorian
    }

    private func day(_ text: String) throws -> Date {
        try XCTUnwrap(calendar.date(from: DateComponents(
            year: Int(text.prefix(4)), month: Int(text.dropFirst(5).prefix(2)), day: Int(text.suffix(2)), hour: 9
        )))
    }

    private func model(
        on today: String = "2026-09-04",
        saved: MoneyRecap_V1_SalarySetting? = nil,
        saveSalary: @escaping SalaryViewModel.SaveSalary = { _, _ in .saved },
        saveProfile: @escaping CurrencyViewModel.Save = { _, _ in .saved }
    ) throws -> SalaryViewModel {
        try SalaryViewModel(
            currency: "IDR", digits: 0, saved: saved, today: day(today), calendar: calendar,
            locale: Locale(identifier: "en_GB"), saveSalary: saveSalary, saveProfile: saveProfile
        )
    }

    func testNextWaitsForAnAmountAboveZeroAndSavesItForThisMonth() async throws {
        var sent: [(MoneyRecap_V1_BffServiceSaveSalaryRequest, Headers)] = []
        var answer = SaveResult.offline
        let step = try model { request, headers in
            sent.append((request, headers))
            return answer
        }
        XCTAssertEqual(step.amount, "", "the field starts empty")
        XCTAssertEqual(step.payday, "25", "PRD §6.6 default")
        XCTAssertFalse(step.canContinue)
        step.amount = "0"
        XCTAssertFalse(step.canContinue, "zero is no Salary")
        step.amount = "150000000."
        XCTAssertTrue(step.canContinue)
        step.choose("last")
        let offline = await step.next()
        XCTAssertFalse(offline)
        XCTAssertTrue(step.offline, "the card says so and Next tries again")
        answer = .saved
        let saved = await step.next()
        XCTAssertTrue(saved)
        XCTAssertFalse(step.offline)
        XCTAssertEqual(sent.count, 2)
        let request = try XCTUnwrap(sent.last?.0)
        XCTAssertEqual(request.amount.amount, "150000000", "a trailing point is not sent")
        XCTAssertEqual(request.amount.currencyCode, "IDR")
        XCTAssertEqual(request.payday, "last")
        XCTAssertEqual(request.appliesFrom, .thisMonth, "setup never asks which month")
        XCTAssertEqual(sent[0].1, sent[1].1, "a retry of the same answers keeps its key")
        step.amount = "16000000"
        _ = await step.next()
        XCTAssertNotEqual(sent[1].1, sent[2].1, "changed answers, a new key")
    }

    func testSkipForNowRecordsStepThreeAloneUnderItsOwnKey() async throws {
        var salaries = 0
        var skips: [(MoneyRecap_V1_BffServiceSaveProfileRequest, Headers)] = []
        let step = try model(saveSalary: { _, _ in
            salaries += 1
            return .offline
        }, saveProfile: { request, headers in
            skips.append((request, headers))
            return skips.count == 1 ? .offline : .saved
        })
        step.amount = "150000000"
        _ = await step.next()
        let first = await step.skip()
        let second = await step.skip()
        XCTAssertFalse(first)
        XCTAssertTrue(second)
        XCTAssertEqual(salaries, 1)
        XCTAssertEqual(skips.map(\.0.onboardingStep), [3, 3])
        XCTAssertFalse(skips[0].0.hasCurrencyCode || skips[0].0.hasDisplayName, "step 3 alone")
        XCTAssertEqual(skips[0].1, skips[1].1, "the skip retried keeps its key")
    }

    /// A resume after a crash, or step 3 after going back, fills in ledger's answer (spec #92).
    func testTheSavedAnswerIsFilledIn() throws {
        let step = try model(saved: .with {
            $0.amount.amount = "150000000"
            $0.payday = "1"
        })
        XCTAssertEqual(step.amount, "150000000")
        XCTAssertEqual(step.payday, "1")
        XCTAssertTrue(step.canContinue)
    }

    /// PRD D51 and §8.2: today counts, a short month pays on its last day, and every tap redraws the three dates.
    func testTheHelperShowsTheNextThreePaydaysFromToday() throws {
        let step = try model(on: "2026-09-25")
        XCTAssertEqual(step.nextPaydays, ["25 Sep", "25 Oct", "25 Nov"], "picking 25 on the 25th shows today first")
        step.choose("24")
        XCTAssertEqual(step.nextPaydays, ["24 Oct", "24 Nov", "24 Dec"])
        step.choose("31")
        XCTAssertEqual(step.nextPaydays, ["30 Sep", "31 Oct", "30 Nov"], "PRD §12, criterion 13")
        step.choose("last")
        XCTAssertEqual(step.nextPaydays, ["30 Sep", "31 Oct", "30 Nov"])
        let january = try model(on: "2027-01-31")
        january.choose("30")
        XCTAssertEqual(january.nextPaydays, ["28 Feb", "30 Mar", "30 Apr"])
    }
}
