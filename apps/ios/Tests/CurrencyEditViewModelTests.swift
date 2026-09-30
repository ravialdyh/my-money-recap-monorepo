import Connect
import MoneyRecapContract
import SwiftData
import SwiftProtobuf
import XCTest
@testable import MyMoneyRecap

/// D5 to D7 (PRD §8.9) from the BFF's recorded answers for the test Member in rupiah: with no amounts, then with a
/// Rp15.000.000 Salary and a Rp25.000 Spend previewed and converted to pounds at a typed 0.000046.
@MainActor
final class CurrencyEditViewModelTests: XCTestCase {
    nonisolated static func recorded<Answer: SwiftProtobuf.Message>(
        _ name: String,
        as _: Answer.Type = Answer.self
    ) throws -> Answer {
        let url = try XCTUnwrap(Bundle(for: Self.self).url(forResource: name, withExtension: "json"))
        return try Answer(jsonString: String(contentsOf: url, encoding: .utf8))
    }

    typealias PreviewAnswer = MoneyRecap_V1_BffServicePreviewCurrencyChangeResponse
    typealias ChangeAnswer = MoneyRecap_V1_BffServiceChangeCurrencyResponse

    /// The recorded typed-rate preview as a fetched one (#175 adds the fetch).
    nonisolated static func fetched() throws -> PreviewAnswer {
        var answer: PreviewAnswer = try recorded("currency-preview")
        answer.preview.source = .provider
        return answer
    }

    private final class Calls {
        var previews: [MoneyRecap_V1_BffServicePreviewCurrencyChangeRequest] = []
        var changes: [(MoneyRecap_V1_BffServiceChangeCurrencyRequest, Headers)] = []
    }

    private func model(
        preview: @escaping (MoneyRecap_V1_BffServicePreviewCurrencyChangeRequest) throws -> PreviewAnswer,
        change: @escaping () throws -> ChangeAnswer = { try recorded("currency-converted") }
    ) -> (CurrencyEditViewModel, Calls) {
        let calls = Calls()
        let model = CurrencyEditViewModel(
            current: "IDR",
            pause: .zero,
            locale: Locale(identifier: "id_ID")
        ) { request in
            calls.previews.append(request)
            return Result { try preview(request) }.mapError(Self.connectError)
        } change: { request, headers in
            calls.changes.append((request, headers))
            return Result { try change() }.mapError(Self.connectError)
        }
        return (model, calls)
    }

    nonisolated static func connectError(_ error: any Error) -> ConnectError {
        error as? ConnectError ?? ConnectError(code: .unknown, message: nil)
    }

    private static func refused(_ code: Code) -> ConnectError {
        ConnectError(code: code, message: nil)
    }

    func testTheCurrentCodeDoesNothingAndTheListIsSearchable() async {
        let (model, calls) = model { _ in try Self.recorded("currency-nothing-to-convert") }
        await model.choose("IDR")
        XCTAssertTrue(calls.previews.isEmpty, "tapping the checked row does nothing")
        model.query = "pound"
        XCTAssertTrue(model.visible.contains { $0.code == "GBP" })
        model.query = "gbp"
        XCTAssertEqual(model.visible.map(\.code), ["GBP"], "by code too")
        model.query = "zzz"
        XCTAssertEqual(model.noMatch, "No currency matches \"zzz\".")
    }

    /// Nothing to convert: the change goes at once with no rate, and the row reads `Changed ✓` until the list opens
    /// again.
    func testNothingToConvertChangesAtOnce() async throws {
        let (model, calls) = model(
            preview: { _ in try Self.recorded("currency-nothing-to-convert") },
            change: { try Self.recorded("currency-changed") }
        )
        await model.choose("GBP")
        XCTAssertEqual(calls.previews.map(\.toCurrencyCode), ["GBP"])
        XCTAssertEqual(calls.previews.first?.manualRate, "", "no rate asked")
        let sent = try XCTUnwrap(calls.changes.first?.0)
        XCTAssertEqual(sent.toCurrencyCode, "GBP")
        XCTAssertEqual(sent.rate, "")
        XCTAssertEqual(sent.source, .unspecified)
        XCTAssertNotNil(calls.changes.first?.1["Idempotency-Key"])
        XCTAssertFalse(model.sheetOpen, "no sheet")
        XCTAssertEqual(model.current, "GBP")
        XCTAssertEqual(model.changedValue, "Changed ✓")
        XCTAssertNil(self.model { _ in PreviewAnswer() }.0.changedValue, "a new list has no Changed ✓")
    }

    /// A fetched rate opens D6: the rate line, the salary sentence and the rows in the device locale; Convert sends the
    /// preview's rate and source.
    func testAFetchedRateOpensTheConfirmSheet() async throws {
        let (model, calls) = model { _ in try Self.fetched() }
        await model.choose("GBP")
        XCTAssertTrue(model.sheetOpen)
        XCTAssertEqual(model.step, .confirm)
        XCTAssertEqual(model.title, "Convert everything to GBP?")
        XCTAssertEqual(
            model.rateLine,
            "All your amounts will be converted at today's rate, 1 IDR = 0,000046 GBP."
        )
        XCTAssertEqual(model.salaryLine, "Your Rp15.000.000 salary becomes £690.")
        XCTAssertEqual(model.rows.map(\.label), ["Salary", "Spends"])
        XCTAssertEqual(model.rows.map(\.value), ["Rp15.000.000 → £690", "Rp25.000 → £1,15"])
        XCTAssertTrue(model.canConvert)
        await model.convert()
        let sent = try XCTUnwrap(calls.changes.first?.0)
        XCTAssertEqual([sent.toCurrencyCode, sent.rate], ["GBP", "0.000046"])
        XCTAssertEqual(sent.source, .provider)
        XCTAssertFalse(model.sheetOpen, "the sheet closes on the new Currency")
        XCTAssertEqual(model.current, "GBP")
        XCTAssertNil(model.changedValue, "Changed ✓ is the direct change's")
    }

    /// A failed fetch opens D6 failed; Try again asks once more, and Enter it myself opens D7 with nothing typed.
    func testAFailedFetchOffersTryAgainAndEnterItMyself() async {
        let (model, calls) = model { _ in throw Self.refused(.unavailable) }
        await model.choose("GBP")
        XCTAssertEqual(model.step, .failed)
        XCTAssertTrue(model.rows.isEmpty, "no rate, no rows")
        XCTAssertFalse(model.canConvert)
        await model.tryAgain()
        XCTAssertEqual(calls.previews.count, 2)
        XCTAssertEqual(model.step, .failed)
        model.enterItMyself()
        XCTAssertEqual(model.step, .manual)
        XCTAssertEqual(model.rate, "")
        XCTAssertFalse(model.canConvert, "disabled until a rate has a preview")
        model.cancel()
        XCTAssertFalse(model.sheetOpen, "Cancel changes nothing")
        XCTAssertTrue(calls.changes.isEmpty)
    }

    func testAnUnsupportedPairOpensTheTypedRateDirectly() async {
        let (model, _) = model { _ in .with { $0.preview.unsupportedPair = true } }
        await model.choose("GBP")
        XCTAssertEqual(model.step, .manual)
    }

    /// D7 asks again after typing stops, only for a rate above zero, and Convert waits for that answer; it sends the
    /// typed rate as MANUAL.
    func testTheTypedRateIsPreviewedThenConverted() async {
        let (model, calls) = model { request in
            request.manualRate.isEmpty ? .with { $0.preview.unsupportedPair = true } : try Self
                .recorded("currency-preview")
        }
        await model.choose("GBP")
        model.type("0")
        XCTAssertNil(model.typing, "zero asks nothing")
        model.type("0,000046")
        XCTAssertFalse(model.canConvert, "not before the preview answers")
        await model.typing?.value
        XCTAssertEqual(calls.previews.last?.manualRate, "0.000046", "the BFF reads a point")
        XCTAssertEqual(model.salaryLine, "Your Rp15.000.000 salary becomes £690.")
        XCTAssertNil(model.rateLine, "D7 has no today's rate")
        XCTAssertTrue(model.canConvert)
        await model.convert()
        XCTAssertEqual(calls.changes.first?.0.rate, "0.000046")
        XCTAssertEqual(calls.changes.first?.0.source, .manual)
    }

    func testTheTypedRatesLimits() {
        XCTAssertEqual(CurrencyEditViewModel.clamp("123456789012345"), "12345678901", "11 digits before")
        XCTAssertEqual(CurrencyEditViewModel.clamp("1.12345678901234"), "1.123456789012", "12 after")
        XCTAssertEqual(CurrencyEditViewModel.clamp("1.2.3"), "1.23", "one separator")
        XCTAssertEqual(CurrencyEditViewModel.clamp("a1-b,5"), "1,5", "digits only")
        XCTAssertNil(CurrencyEditViewModel.sendable(""))
        XCTAssertNil(CurrencyEditViewModel.sendable("0,000"))
        XCTAssertEqual(CurrencyEditViewModel.sendable(".5"), "0.5")
        XCTAssertEqual(CurrencyEditViewModel.sendable("5."), "5")
    }

    /// Past D57 the sheet says so with Convert off; any other failed Convert asks to try again under the same key.
    func testRefusalsShowInTheSheet() async {
        var refusal = Code.outOfRange
        let (model, calls) = model(preview: { request in
            if request.manualRate == "99999999" {
                throw Self.refused(.outOfRange)
            }
            return try Self.fetched()
        }, change: { throw Self.refused(refusal) })
        await model.choose("GBP")
        await model.convert()
        XCTAssertEqual(model.refusal, "That's more than the app can store.")
        XCTAssertFalse(model.canConvert)
        XCTAssertTrue(model.sheetOpen)
        model.enterItMyself()
        model.type("99999999")
        await model.typing?.value
        XCTAssertEqual(model.refusal, "That's more than the app can store.")
        XCTAssertFalse(model.canConvert)
        model.type("0.000046")
        await model.typing?.value
        refusal = .internalError
        await model.convert()
        await model.convert()
        XCTAssertEqual(model.refusal, "We couldn't convert your amounts. Try again.")
        XCTAssertTrue(model.canConvert, "it may try again")
        XCTAssertEqual(calls.changes[1].1, calls.changes[2].1, "the retry counts once")
    }

    /// Through Home's Settings: the answer replaces Home and Settings, stored and shown, and every stored Spend list
    /// and
    /// Recap is erased, since they hold the old Currency.
    func testAConversionReplacesHomeAndSettingsAndErasesStoredScreens() async throws {
        let context = try SettingsViewModelTests.context()
        let zero: MoneyRecap_V1_GetHomeResponse = try Self.recorded("home-zero")
        try HomeSnapshot.replace(with: zero.home, in: context)
        try SpendListSnapshot.replace(with: .with { $0.month = "2026-09" }, in: context)
        try RecapSnapshot.replace(with: .with { $0.recap.month = "2026-09" }, in: context)
        let converted: ChangeAnswer = try Self.recorded("currency-converted")
        let fetched = try Self.fetched()
        let home = HomeViewModel(
            fetch: { zero.home }, context: context, locale: Locale(identifier: "id_ID"),
            getSettings: { try SettingsViewModelTests.recorded("settings-filled") },
            previewCurrency: { _ in .success(fetched) }, changeCurrency: { _, _ in .success(converted) }
        )
        let settings = home.settings()
        settings.openCurrency()
        XCTAssertNil(settings.currencyEdit, "nothing stored and no answer yet")
        await settings.refresh()
        settings.openCurrency()
        let currency = try XCTUnwrap(settings.currencyEdit)
        XCTAssertEqual(currency.current, "IDR")
        await currency.choose("GBP")
        await currency.convert()
        XCTAssertEqual(home.home, converted.home)
        XCTAssertEqual(HomeSnapshot.load(from: context)?.bytes, try converted.home.serializedBytes() as Data)
        var kept = converted.settings
        kept.notificationPrefs = try SettingsViewModelTests.recorded("settings-filled").notificationPrefs
        XCTAssertEqual(settings.settings, kept, "the answer, with the stored preferences it lacks")
        XCTAssertEqual(SettingsSnapshot.load(from: context), kept)
        XCTAssertEqual(settings.currency, "GBP")
        XCTAssertNil(SpendListSnapshot.load("2026-09", from: context))
        XCTAssertNil(RecapSnapshot.load("2026-09", from: context))
    }
}
