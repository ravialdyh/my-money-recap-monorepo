import Connect
import Foundation
import MoneyRecapContract
import XCTest
@testable import MyMoneyRecap

/// Step 2 (PRD §6.5, spec #92): the platform's common currencies, each one the server accepts, sorted by British name
/// and searched by name or code; the region's currency starts chosen; Next saves it with step 2, one key per choice.
@MainActor
final class CurrencyViewModelTests: XCTestCase {
    private let britain = Locale(identifier: "en_GB")

    private func model(region: Locale = Locale(identifier: "id_ID"), save: @escaping CurrencyViewModel.Save = { _, _ in
        true
    }) -> CurrencyViewModel {
        CurrencyViewModel(region: region, save: save)
    }

    func testTheListIsThePlatformsCommonCurrenciesSortedByTheirBritishNames() {
        let options = CurrencyViewModel.options()
        XCTAssertEqual(Set(options.map(\.code)), Set(Locale.commonISOCurrencyCodes), "each common currency once")
        for (earlier, later) in zip(options, options.dropFirst()) {
            XCTAssertNotEqual(later.name.compare(earlier.name, locale: britain), .orderedAscending, "\(later.name)")
        }
        XCTAssertTrue(options.contains { $0.label == "IDR · Indonesian Rupiah" })
        XCTAssertTrue(options.contains { $0.label == "GBP · British Pound" })
        XCTAssertFalse(options.contains { $0.code == "DEM" || $0.code == "XAU" }, "no retired money and no gold")
    }

    /// The server accepts every row (ARCHITECTURE §7.2): each listed code is one some territory takes as tender in
    /// the CLDR file ledger rounds with, read here from the repository, so a new iOS that lists more fails here first.
    func testEveryListedCurrencyIsOneTheServerAccepts() throws {
        let path = URL(fileURLWithPath: #filePath).deletingLastPathComponent().deletingLastPathComponent()
            .deletingLastPathComponent().appendingPathComponent("backend/internal/cldr/currencyData.json")
        let file = try JSONSerialization.jsonObject(with: Data(contentsOf: path)) as? [String: Any]
        let data = (file?["supplemental"] as? [String: Any])?["currencyData"] as? [String: Any]
        let regions = try XCTUnwrap(data?["region"] as? [String: [[String: [String: String]]]])
        let tender = regions.values.joined().flatMap { entry in
            entry.filter { $0.value["_tender"] != "false" }.keys
        }
        XCTAssertGreaterThan(tender.count, 200, "the file read is CLDR's whole currency list")
        XCTAssertEqual(Set(CurrencyViewModel.options().map(\.code)).subtracting(tender), [])
        XCTAssertFalse(Set(tender).contains("XAU"), "the same rule refuses gold")
    }

    func testSearchMatchesACodeOrANameAndSaysWhenNothingDoes() {
        let step = model()
        step.query = "idr"
        XCTAssertEqual(step.visible.map(\.code), ["IDR"])
        step.query = " rupiah "
        XCTAssertTrue(step.visible.map(\.code).contains("IDR"))
        XCTAssertNil(step.noMatch)
        step.query = "ducat"
        XCTAssertEqual(step.visible, [])
        XCTAssertEqual(step.noMatch, "No currency matches \"ducat\".", "SCREEN-STATES A5 empty search")
        XCTAssertTrue(step.canContinue, "the chosen row stays chosen underneath")
        step.clearSearch()
        XCTAssertEqual(step.visible.count, CurrencyViewModel.options().count)
        XCTAssertNil(step.noMatch)
    }

    func testTheRegionsOwnCurrencyStartsChosen() {
        XCTAssertEqual(model(region: Locale(identifier: "id_ID")).selected, "IDR")
        XCTAssertEqual(model(region: britain).selected, "GBP")
        XCTAssertTrue(model(region: britain).canContinue, "so most Members just tap Next")
    }

    func testARegionThatNamesNoCurrencyChoosesNothingAndNextWaits() {
        let step = model(region: Locale(identifier: "en_001"))
        XCTAssertNil(step.selected, "the app never quietly picks a currency")
        XCTAssertFalse(step.canContinue)
        step.choose("EUR")
        XCTAssertEqual(step.selected, "EUR")
        XCTAssertTrue(step.canContinue)
    }

    func testNextSavesTheCurrencyWithStepTwoUnderOneKeyPerChoice() async {
        var sent: [(MoneyRecap_V1_BffServiceSaveProfileRequest, String)] = []
        var answer = false
        let step = model { request, headers in
            sent.append((request, headers["Idempotency-Key"]?.first ?? ""))
            return answer
        }
        let failed = await step.next()
        XCTAssertFalse(failed)
        XCTAssertTrue(step.canContinue, "a failed save leaves Next on")
        answer = true
        let retried = await step.next()
        XCTAssertTrue(retried)
        step.choose("USD")
        _ = await step.next()
        XCTAssertEqual(sent.map(\.0.currencyCode), ["IDR", "IDR", "USD"])
        XCTAssertEqual(Set(sent.map(\.0.onboardingStep)), [2])
        XCTAssertFalse(sent[0].0.hasDisplayName || sent[0].0.hasAvatarKey, "step 2 sends only its answer")
        XCTAssertEqual(sent[0].1, sent[1].1, "a retry of the same choice reuses its key")
        XCTAssertNotEqual(sent[1].1, sent[2].1, "another choice gets a new key")
        XCTAssertFalse(sent[0].1.isEmpty)
    }
}
