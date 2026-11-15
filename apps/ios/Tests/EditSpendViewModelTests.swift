import Connect
import Foundation
import MoneyRecapContract
import SwiftData
import XCTest
@testable import MyMoneyRecap

@MainActor
final class EditSpendViewModelTests: XCTestCase {
    private var calendar: Calendar {
        var gregorian = Calendar(identifier: .gregorian)
        gregorian.timeZone = TimeZone(identifier: "Asia/Jakarta") ?? .gmt
        return gregorian
    }

    private let unavailable = ConnectError(code: .unavailable, message: nil)

    private func sheet(
        _ spend: MoneyRecap_V1_Spend,
        edit: @escaping LogSpendViewModel.Edit = { _, _ in .failure(ConnectError(code: .unavailable, message: nil)) },
        delete: @escaping LogSpendViewModel
            .Delete = { _, _ in .failure(ConnectError(code: .unavailable, message: nil)) },
        connection: Connection? = nil
    ) throws -> LogSpendViewModel {
        let today = try XCTUnwrap(calendar.date(from: DateComponents(year: 2026, month: 9, day: 26, hour: 9)))
        return LogSpendViewModel(
            editing: spend, currency: "IDR", digits: 0, earliest: "2026-07-29", today: today, calendar: calendar,
            locale: Locale(identifier: "en_GB"), connection: connection ?? Connection(online: true), edit: edit,
            delete: delete
        )
    }

    func testOfflineTheEditCanBeReadButNotSavedOrDeleted() async throws {
        let recorded = try LogSpendViewModelTests.recordedAnswer()
        var calls = 0
        let connection = Connection(online: false)
        let model = try sheet(recorded.spend, edit: { _, _ in
            calls += 1
            return .failure(self.unavailable)
        }, delete: { _, _ in
            calls += 1
            return .failure(self.unavailable)
        }, connection: connection)
        XCTAssertTrue(model.offline)
        XCTAssertEqual(model.amount, "25000", "the Spend stays readable")
        XCTAssertFalse(model.canSave)
        let saved = await model.save()
        let deleted = await model.deleteSpend()
        XCTAssertNil(saved)
        XCTAssertNil(deleted)
        XCTAssertEqual(calls, 0, "nothing is sent with no network path")
        XCTAssertFalse(model.failed || model.deleteFailed, "no failure line: nothing was tried")
        connection.online = true
        XCTAssertFalse(model.offline)
        XCTAssertTrue(model.canSave)
    }

    func testTheSheetIsPrefilledAndSavesEveryFieldUnderOneKeyPerAttempt() async throws {
        let recorded = try LogSpendViewModelTests.recordedAnswer()
        var sent: [(MoneyRecap_V1_BffServiceEditSpendRequest, String)] = []
        var answer: Result<MoneyRecap_V1_BffServiceEditSpendResponse, ConnectError> = .failure(unavailable)
        let model = try sheet(recorded.spend) { request, headers in
            sent.append((request, headers["Idempotency-Key"]?.first ?? ""))
            return answer
        }
        XCTAssertEqual(model.amount, "25000")
        XCTAssertEqual(model.category, .snacks)
        XCTAssertEqual(model.note, "Kopi susu")
        XCTAssertEqual(model.dateLabel, "23 September")
        XCTAssertFalse(model.dateLocked, "a date in the last 60 days opens the picker")
        XCTAssertTrue(model.canSave)
        model.choose(.rides)
        model.date = model.dates.lowerBound
        let failed = await model.save()
        XCTAssertNil(failed)
        XCTAssertTrue(model.failed, "We couldn't save this spend. Try again.")
        XCTAssertFalse(model.deleteFailed)
        XCTAssertEqual(model.category, .rides, "the values are intact")
        answer = .success(.with {
            $0.home = recorded.home
            $0.spend = recorded.spend
        })
        let home = await model.save()
        XCTAssertEqual(home, recorded.home, "the answer's Home replaces the snapshot")
        XCTAssertEqual(sent.count, 2)
        XCTAssertEqual(sent[0].1, sent[1].1, "a retry of the same values sends the same key (IOS I3)")
        let request = sent[1].0
        XCTAssertEqual(request.spendID, recorded.spend.id)
        XCTAssertEqual(request.amount.amount, "25000")
        XCTAssertEqual(request.category, .rides)
        XCTAssertEqual(request.date, "2026-07-29")
        XCTAssertEqual(request.note, "Kopi susu")
        XCTAssertEqual(model.months, ["2026-09", "2026-07"], "both months the move changed")
    }

    func testAnOldDateStaysReadOnly() async throws {
        var old = try LogSpendViewModelTests.recordedAnswer().spend
        old.date = "2026-07-10"
        var sent: MoneyRecap_V1_BffServiceEditSpendRequest?
        let model = try sheet(old) { request, _ in
            sent = request
            return .failure(ConnectError(code: .internalError, message: nil))
        }
        XCTAssertTrue(model.dateLocked)
        XCTAssertEqual(model.dateLabel, "10 July")
        XCTAssertEqual(LogSpendViewModel.oldDateHelper, "Dates older than 60 days can't be changed.")
        model.amount = "30000"
        _ = await model.save()
        XCTAssertEqual(sent?.date, "2026-07-10")
        XCTAssertEqual(sent?.amount.amount, "30000")
        XCTAssertEqual(model.months, ["2026-07"])
    }

    func testAFailedDeleteKeepsTheSpendAndARetryKeepsItsKey() async throws {
        let recorded = try LogSpendViewModelTests.recordedAnswer()
        var keys: [String] = []
        var ids: [String] = []
        var answer: Result<MoneyRecap_V1_BffServiceDeleteSpendResponse, ConnectError> = .failure(unavailable)
        let model = try sheet(recorded.spend, delete: { request, headers in
            keys.append(headers["Idempotency-Key"]?.first ?? "")
            ids.append(request.spendID)
            return answer
        })
        _ = await model.save()
        XCTAssertTrue(model.failed)
        let failed = await model.deleteSpend()
        XCTAssertNil(failed)
        XCTAssertTrue(model.deleteFailed)
        XCTAssertFalse(model.failed, "one failure line at a time")
        XCTAssertEqual(LogSpendViewModel.deleteFailure, "We couldn't delete this spend. Try again.")
        XCTAssertEqual(model.amount, "25000", "the Spend is kept")
        answer = .success(.with { $0.home = recorded.home })
        let home = await model.deleteSpend()
        XCTAssertEqual(home, recorded.home)
        XCTAssertFalse(model.deleteFailed)
        XCTAssertEqual(ids, [recorded.spend.id, recorded.spend.id])
        XCTAssertEqual(keys[0], keys[1], "a retried delete is one delete (A14)")
        let log = LogSpendViewModel(currency: "IDR", digits: 0, earliest: "") { _, _ in .failure(self.unavailable) }
        let none = await log.deleteSpend()
        XCTAssertNil(none, "Log a spend has nothing to delete")
    }

    func testHomeOpensTheSheetAndTheListAsksForEveryChangedMonth() async throws {
        let recorded = try LogSpendViewModelTests.recordedAnswer()
        let context = try ModelContext(ModelContainer(
            for: HomeSnapshot.self, SpendListSnapshot.self,
            configurations: ModelConfiguration(isStoredInMemoryOnly: true)
        ))
        var asked: [String] = []
        let home = HomeViewModel(
            fetch: { recorded.home },
            screens: HomeScreens(context: context, spendCalls: SpendCalls(list: { month in
                asked.append(month)
                return .with {
                    $0.month = month
                    $0.months = ["2026-09", "2026-08"]
                }
            }))
        )
        XCTAssertNil(home.editSheet(recorded.spend), "no sheet before Home has an answer")
        await home.refresh()
        let sheet = try XCTUnwrap(home.editSheet(recorded.spend))
        XCTAssertEqual(sheet.currency, "IDR")
        XCTAssertEqual(sheet.spend, recorded.spend)
        let list = try XCTUnwrap(home.spendList(switching: false))
        list.category = .snacks
        await list.refresh(["2026-08"])
        XCTAssertEqual(asked, ["2026-08", "2026-09"])
        XCTAssertNotNil(SpendListSnapshot.load("2026-08", from: context), "the month the Spend moved to is stored anew")
        XCTAssertNil(list.category, "Snacks has no Spend left, so its detail closes")
    }
}
