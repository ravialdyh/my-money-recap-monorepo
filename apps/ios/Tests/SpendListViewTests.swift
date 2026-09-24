import MoneyRecapContract
import SwiftData
import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// The Spend list and Category detail as drawn (SCREEN-STATES B6, B6b): the rows, the switcher open, both empty months
/// and the detail each differ. The magic-link UI test turns them sideways.
@MainActor
final class SpendListViewTests: XCTestCase {
    func testEveryStateIsDrawnApart() async throws {
        let url = try XCTUnwrap(Bundle(for: Self.self).url(forResource: "list-spends", withExtension: "json"))
        let list = try MoneyRecap_V1_BffServiceListSpendsResponse(jsonString: String(contentsOf: url, encoding: .utf8))
            .spends
        var empty = list
        empty.categories = []
        empty.spends = []
        empty.months = ["2026-09", "2026-08"]
        var past = empty
        past.month = "2026-08"
        let context = try ModelContext(ModelContainer(
            for: HomeSnapshot.self, SpendListSnapshot.self,
            configurations: ModelConfiguration(isStoredInMemoryOnly: true)
        ))
        let drawn = { (answer: MoneyRecap_V1_SpendListView, switching: Bool) async -> Data? in
            let model = SpendListViewModel(
                month: answer.month,
                switching: switching,
                fetch: { _ in answer },
                context: context
            )
            await model.refresh()
            return self.picture(SpendListView(model: model))
        }
        let rows = await drawn(list, false)
        let switcher = await drawn(list, true)
        let thisMonth = await drawn(empty, false)
        let lastMonth = await drawn(past, false)
        let model = SpendListViewModel(month: "2026-09", fetch: { _ in list }, context: context)
        await model.refresh()
        let detail = picture(CategoryView(model: model, category: .snacks))
        XCTAssertNotNil(rows)
        XCTAssertEqual(Set([rows, switcher, thisMonth, lastMonth, detail]).count, 5)
    }
}
