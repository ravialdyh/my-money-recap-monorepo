import Connect
import Foundation
import MoneyRecapContract
import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// The Log a spend sheet as drawn (SCREEN-STATES B2 to B4): empty, filled, at the 11-digit cap with its line, failed
/// with its line; and as Edit spend (B7): prefilled, an old date, a failed save and a failed delete.
@MainActor
final class LogSpendViewTests: XCTestCase {
    private func drawn(_ model: LogSpendViewModel) -> Data? {
        picture(LogSpendView(model: model) { _ in })
    }

    func testEveryStateIsDrawnApart() async {
        let model = LogSpendViewModel(currency: "IDR", digits: 0, earliest: "") { _, _ in .failure(ConnectError(
            code: .unavailable,
            message: nil
        )) }
        let empty = drawn(model)
        model.amount = "25000"
        model.choose(.snacks)
        let filled = drawn(model)
        model.amount = "99999999999"
        let capped = drawn(model)
        model.amount = "25000"
        _ = await model.save()
        XCTAssertTrue(model.failed)
        let failed = drawn(model)
        XCTAssertNotNil(empty)
        XCTAssertEqual(Set([empty, filled, capped, failed]).count, 4, "empty, filled, at the cap and failed differ")
    }

    func testEveryEditStateIsDrawnApart() async throws {
        var spend = try LogSpendViewModelTests.recordedAnswer().spend
        let refused: LogSpendViewModel.Edit = { _, _ in .failure(ConnectError(code: .internalError, message: nil)) }
        let sheet = { (spend: MoneyRecap_V1_Spend, online: Bool) in
            LogSpendViewModel(
                editing: spend, currency: "IDR", digits: 0, earliest: "2026-09-01",
                connection: Connection(online: online), edit: refused,
                delete: { _, _ in .failure(ConnectError(code: .internalError, message: nil)) }
            )
        }
        spend.date = Date.now.formatted(Date.ISO8601FormatStyle(timeZone: .current).year().month().day())
        let model = sheet(spend, true)
        let edit = drawn(model)
        _ = await model.save()
        let saveFailed = drawn(model)
        _ = await model.deleteSpend()
        let deleteFailed = drawn(model)
        spend.date = "2026-07-10"
        let old = drawn(sheet(spend, true))
        let log = drawn(LogSpendViewModel(currency: "IDR", digits: 0, earliest: "") { _, _ in
            .failure(ConnectError(code: .unavailable, message: nil))
        })
        XCTAssertNotNil(edit)
        let offline = drawn(sheet(spend, false))
        XCTAssertEqual(
            Set([edit, saveFailed, deleteFailed, old, log, offline]).count, 6,
            "prefilled, save failed, delete failed, the old date, Log a spend and offline differ"
        )
    }
}
