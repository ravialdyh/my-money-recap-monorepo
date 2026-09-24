import Connect
import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// The Log a spend sheet as drawn (SCREEN-STATES B2 to B4): empty, filled, at the 11-digit cap with its line, failed
/// with its line.
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
}
