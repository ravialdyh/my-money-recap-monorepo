import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// Step 2 as drawn (PRD §6.5): the chosen row carries the check, a search that matches nothing draws its line and the
/// link in place of the list, and a list row checks only when chosen.
@MainActor
final class CurrencyViewTests: XCTestCase {
    /// `view` as it draws in a 393 × 852 window, the size of the iPhone the UI tests run on.
    private func picture(_ view: some View) -> Data? {
        let host = UIHostingController(rootView: view)
        let window = UIWindow(frame: CGRect(x: 0, y: 0, width: 393, height: 852))
        window.rootViewController = host
        window.makeKeyAndVisible()
        host.view.layoutIfNeeded()
        return UIGraphicsImageRenderer(bounds: host.view.bounds).image { _ in
            host.view.drawHierarchy(in: host.view.bounds, afterScreenUpdates: true)
        }.pngData()
    }

    private func step(region: String, query: String = "") -> some View {
        let model = CurrencyViewModel(region: Locale(identifier: region)) { _, _ in true }
        model.query = query
        return CurrencyView(model: model) {}
    }

    func testTheChosenCurrencyIsCheckedAndANoMatchSearchDrawsItsLine() {
        let unchosen = picture(step(region: "en_001", query: "rupiah"))
        let chosen = picture(step(region: "id_ID", query: "rupiah"))
        let nothing = picture(step(region: "id_ID", query: "ducat"))
        XCTAssertNotNil(unchosen)
        XCTAssertNotEqual(unchosen, chosen, "the region's currency is drawn checked")
        XCTAssertNotEqual(chosen, nothing, "a search matching nothing draws its line and Clear search")
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
