import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// Step 3 as drawn (PRD §6.6): the chosen Payday is the ink pill with its dates under the picker, and a save that
/// could not reach the BFF is drawn in the amount card.
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

    /// A save that could not reach the BFF draws `No connection. We need one to save your answers.` in the card.
    func testANoConnectionSaveIsDrawnInTheCard() async {
        let model = step(.offline)
        model.amount = "15000000"
        let before = picture(SalaryView(model: model, onBack: {}, onDone: {}))
        _ = await model.skip()
        XCTAssertTrue(model.offline)
        XCTAssertNotEqual(before, picture(SalaryView(model: model, onBack: {}, onDone: {})))
    }
}
