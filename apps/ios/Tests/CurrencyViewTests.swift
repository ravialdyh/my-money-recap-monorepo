import SwiftUI
import XCTest
@testable import MyMoneyRecap

@MainActor
final class CurrencyViewTests: XCTestCase {
    private func step(region: String, query: String = "") -> some View {
        let model = CurrencyViewModel(region: Locale(identifier: region)) { _, _ in .saved }
        model.query = query
        return CurrencyView(model: model, onBack: {}, onSaved: {})
    }

    func testTheChosenCurrencyIsCheckedAndANoMatchSearchDrawsItsLine() {
        let unchosen = picture(step(region: "en_001", query: "rupiah"))
        let chosen = picture(step(region: "id_ID", query: "rupiah"))
        let nothing = picture(step(region: "id_ID", query: "ducat"))
        XCTAssertNotNil(unchosen)
        XCTAssertNotEqual(unchosen, chosen, "the region's currency is drawn checked")
        XCTAssertNotEqual(chosen, nothing, "a search matching nothing draws its line and Clear search")
    }

    func testANoConnectionSaveIsDrawnInTheCard() async {
        let model = CurrencyViewModel(region: Locale(identifier: "id_ID")) { _, _ in .offline }
        let before = picture(CurrencyView(model: model, onBack: {}, onSaved: {}))
        _ = await model.next()
        XCTAssertTrue(model.offline)
        XCTAssertNotEqual(before, picture(CurrencyView(model: model, onBack: {}, onSaved: {})))
    }

    func testSidewaysPutsTheQuestionBesideTheCard() {
        let upright = picture(step(region: "id_ID"))
        let sideways = picture(step(region: "id_ID").environment(\.verticalSizeClass, .compact))
        XCTAssertNotNil(sideways)
        XCTAssertNotEqual(upright, sideways)
    }

    func testAListRowChecksOnlyWhenChosen() {
        let row = { (trailing: ListRow.Trailing, selected: Bool) in
            ListRow(label: "IDR · Indonesian Rupiah", trailing: trailing, selected: selected)
                .padding()
                .background(.white)
        }
        XCTAssertNotEqual(picture(row(.check, false)), picture(row(.check, true)))
        XCTAssertNotEqual(picture(row(.check, false)), picture(row(.chevron, false)), "a choice row has no chevron")
    }
}
