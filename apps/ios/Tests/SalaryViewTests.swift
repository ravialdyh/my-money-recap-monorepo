import SwiftUI
import XCTest
@testable import MyMoneyRecap

@MainActor
final class SalaryViewTests: XCTestCase {
    private func step(_ answer: SaveResult = .saved) -> SalaryViewModel {
        SalaryViewModel(currency: "IDR", digits: 0, saveSalary: { _, _ in answer }, saveProfile: { _, _ in answer })
    }

    func testAnotherPaydayIsDrawnChosen() {
        let model = step()
        let before = picture(SalaryView(model: model, onBack: {}, onDone: {}))
        model.choose("1")
        XCTAssertNotEqual(before, picture(SalaryView(model: model, onBack: {}, onDone: {})))
    }

    func testANoConnectionSaveIsDrawnInTheCard() async {
        let model = step(.offline)
        model.amount = "150000000"
        let before = picture(SalaryView(model: model, onBack: {}, onDone: {}))
        _ = await model.skip()
        XCTAssertTrue(model.offline)
        XCTAssertNotEqual(before, picture(SalaryView(model: model, onBack: {}, onDone: {})))
    }
}
