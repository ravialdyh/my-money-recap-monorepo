import Connect
import Foundation
import MoneyRecapContract
import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// Step 4 (PRD §6.7, D44, spec #92): 20% by default, the preview a share of the Salary rounded half up to the
/// Currency's
/// decimals, Done saving the percent under one key per answer, Skip for now recording step 4; no Salary blocks Done.
@MainActor
final class TargetViewModelTests: XCTestCase {
    private func money(_ amount: String, _ code: String) -> MoneyRecap_V1_Money {
        .with {
            $0.amount = amount
            $0.currencyCode = code
        }
    }

    private func model(
        salary: MoneyRecap_V1_Money?,
        digits: Int = 0,
        saved: MoneyRecap_V1_SavingTargetSetting? = nil,
        locale: String = "en_GB",
        saveTarget: @escaping TargetViewModel.SaveTarget = { _, _ in .saved },
        saveProfile: @escaping CurrencyViewModel.Save = { _, _ in .saved }
    ) -> TargetViewModel {
        TargetViewModel(
            salary: salary, digits: digits, saved: saved, locale: Locale(identifier: locale),
            saveTarget: saveTarget, saveProfile: saveProfile
        )
    }

    func testThePreviewIsAShareOfTheSalaryRoundedHalfUpToTheCurrencysDecimals() {
        let step = model(salary: money("2400.00", "GBP"), digits: 2)
        XCTAssertEqual(step.percent, 20, "PRD §6.7 default")
        XCTAssertEqual(step.preview, "£480 of £2,400 each month")
        step.percent = 35
        XCTAssertEqual(step.preview, "£840 of £2,400 each month")
        let odd = model(salary: money("1234.50", "GBP"), digits: 2)
        odd.percent = 15
        XCTAssertEqual(odd.preview, "£185.18 of £1,234.50 each month", "185.175 rounds half up")
        let rupiah = model(salary: money("15000005", "IDR"), locale: "id_ID")
        rupiah.percent = 5
        XCTAssertEqual(rupiah.preview, "Rp750.000 of Rp15.000.005 each month", "750,000.25 to no decimals")
        XCTAssertNil(model(salary: nil).preview, "no Salary, no amount (D44)")
    }

    func testDoneSavesThePercentWithOneKeyPerAnswer() async {
        var sent: [(MoneyRecap_V1_BffServiceSaveTargetRequest, Headers)] = []
        var answer = SaveResult.offline
        let step = model(salary: money("15000000", "IDR"), saved: .with { $0.percent = 35 }) { request, headers in
            sent.append((request, headers))
            return answer
        }
        XCTAssertEqual(step.percent, 35, "a resumed answer is selected again")
        let offline = await step.done()
        XCTAssertFalse(offline)
        XCTAssertTrue(step.offline, "the step says so and Done tries again")
        answer = .saved
        let saved = await step.done()
        XCTAssertTrue(saved)
        XCTAssertFalse(step.offline)
        XCTAssertEqual(sent.map(\.0.percent), [35, 35])
        XCTAssertEqual(sent[0].1, sent[1].1, "a retry keeps its key")
        step.percent = 40
        _ = await step.done()
        XCTAssertNotEqual(sent[2].1, sent[1].1, "a changed answer gets a new key")
    }

    func testWithoutASalaryDoneIsRefusedAndSkipRecordsStepFour() async {
        var saves = 0
        var steps: [Int32] = []
        let step = model(salary: nil, saveTarget: { _, _ in
            saves += 1
            return .saved
        }, saveProfile: { request, _ in
            steps.append(request.onboardingStep)
            return .saved
        })
        XCTAssertTrue(step.isBlocked)
        let done = await step.done()
        XCTAssertFalse(done)
        XCTAssertEqual(saves, 0, "no Target without a Salary (D44)")
        let skipped = await step.skip()
        XCTAssertTrue(skipped)
        XCTAssertEqual(steps, [4], "step 4 alone, which ends setup")
    }

    func testTheSliderSnapsToStepsOfFiveInsideTheRange() {
        let slider = MyMoneyRecap.Slider(label: "Saving target", value: .constant(20))
        XCTAssertEqual(slider.snapped(0.21), 20)
        XCTAssertEqual(slider.snapped(0.23), 25)
        XCTAssertEqual(slider.snapped(-1), 0)
        XCTAssertEqual(slider.snapped(2), 100)
    }
}
