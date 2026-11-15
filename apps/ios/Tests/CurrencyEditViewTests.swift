import Connect
import MoneyRecapContract
import SwiftUI
import XCTest
@testable import MyMoneyRecap

@MainActor
final class CurrencyEditViewTests: XCTestCase {
    private let sideways = CGSize(width: 852, height: 393)

    private func sheet(
        _ preview: @escaping (String) throws -> PreviewAnswer
    ) async -> CurrencyEditViewModel {
        let model = CurrencyEditViewModel(current: "IDR", pause: .zero, locale: Locale(identifier: "id_ID")) {
            let rate = $0.manualRate
            return Result { try preview(rate) }.mapError(CurrencyEditViewModelTests.connectError)
        } change: { _, _ in .failure(ConnectError(code: .unavailable, message: nil)) }
        await model.choose("GBP")
        return model
    }

    func testTheConfirmSheetDrawsInRupiahAndEachStateApart() async {
        let confirm = await sheet { _ in try CurrencyEditViewModelTests.fetched() }
        let failed = await sheet { _ in throw ConnectError(code: .unavailable, message: nil) }
        let manual = await sheet { rate in
            rate.isEmpty ? .with { $0.preview.unsupportedPair = true }
                : try CurrencyEditViewModelTests.recorded("currency-preview")
        }
        let rupiah = picture(ConvertSheet(model: confirm))
        XCTAssertNotNil(rupiah)
        XCTAssertNotEqual(rupiah, picture(ConvertSheet(model: failed)), "Try again and Enter it myself, no rows")
        let typing = picture(ConvertSheet(model: manual))
        XCTAssertNotEqual(rupiah, typing, "the rate field")
        manual.type("0.000046")
        await manual.typing?.value
        XCTAssertNotEqual(
            typing,
            picture(ConvertSheet(model: manual)),
            "the convert sheet's rows once a preview answers"
        )
        for (model, step) in [(confirm, "the convert sheet"), (manual, "the manual rate sheet")] {
            XCTAssertGreaterThan(overflow(ConvertSheet(model: model), size: sideways), 0, "\(step) scrolls sideways")
        }
        XCTAssertLessThan(
            overflow(ConvertSheet(model: failed), size: sideways),
            0,
            "the failed convert sheet fits in its scroll view"
        )
    }

    typealias PreviewAnswer = MoneyRecap_V1_BffServicePreviewCurrencyChangeResponse

    func testTheListShowsChangedAndScrollsSideways() async {
        let plain = await sheet { _ in throw ConnectError(code: .unavailable, message: nil) }
        plain.cancel()
        let changed = CurrencyEditViewModel(current: "IDR", locale: Locale(identifier: "id_ID")) { _ in
            .success(.with { $0.preview.nothingToConvert = true })
        } change: { _, _ in .success(.init()) }
        await changed.choose("GBP")
        XCTAssertNotEqual(picture(CurrencyEditView(model: plain)), picture(CurrencyEditView(model: changed)))
        XCTAssertGreaterThan(
            overflow(CurrencyEditView(model: plain), size: sideways),
            0,
            "the currency list scrolls sideways"
        )
    }
}
