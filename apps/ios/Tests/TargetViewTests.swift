import MoneyRecapContract
import SwiftUI
import XCTest
@testable import MyMoneyRecap

@MainActor
final class TargetViewTests: XCTestCase {
    private func step(salary: String?, _ answer: SaveResult = .saved) -> TargetViewModel {
        TargetViewModel(
            salary: salary.map { amount in .with { $0.amount = amount; $0.currencyCode = "IDR" } },
            digits: 0,
            saveTarget: { _, _ in answer },
            saveProfile: { _, _ in answer }
        )
    }

    private func drawn(_ model: TargetViewModel) -> Data? {
        picture(TargetView(model: model, onBack: {}, onSetSalary: {}, onDone: {}))
    }

    func testAnotherPercentIsDrawn() {
        let model = step(salary: "150000000")
        let before = drawn(model)
        model.percent = 50
        XCTAssertNotEqual(before, drawn(model))
    }

    func testWithoutASalaryThePromptReplacesTheSlider() {
        XCTAssertNotEqual(drawn(step(salary: nil)), drawn(step(salary: "150000000")))
    }

    func testANoConnectionSaveIsDrawn() async {
        for salary in [nil, "150000000"] {
            let model = step(salary: salary, .offline)
            let before = drawn(model)
            _ = await model.skip()
            XCTAssertTrue(model.offline)
            XCTAssertNotEqual(before, drawn(model), "in its own card, or in the slider's")
        }
    }
}
