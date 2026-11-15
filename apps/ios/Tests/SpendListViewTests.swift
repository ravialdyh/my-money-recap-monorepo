import MoneyRecapContract
import SwiftData
import SwiftUI
import XCTest
@testable import MyMoneyRecap

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
        var otherSpent = list
        otherSpent.spent.amount = "1"
        let otherLine = await drawn(otherSpent, false)
        XCTAssertNotEqual(rows, otherLine, "the month line under the switcher is drawn")
    }

    func testThePlusButtonDimsOffline() async throws {
        let url = try XCTUnwrap(Bundle(for: Self.self).url(forResource: "list-spends", withExtension: "json"))
        var empty = try MoneyRecap_V1_BffServiceListSpendsResponse(
            jsonString: String(contentsOf: url, encoding: .utf8)
        ).spends
        empty.categories = []
        empty.spends = []
        let context = try ModelContext(ModelContainer(
            for: HomeSnapshot.self, SpendListSnapshot.self,
            configurations: ModelConfiguration(isStoredInMemoryOnly: true)
        ))
        let path = Connection(online: true)
        let model = SpendListViewModel(month: empty.month, fetch: { _ in empty }, context: context, connection: path)
        await model.refresh()
        let online = picture(SpendListView(model: model))
        path.online = false
        XCTAssertTrue(model.offline)
        XCTAssertNotEqual(picture(SpendListView(model: model)), online)
        path.online = true
        XCTAssertEqual(picture(SpendListView(model: model)), online)
    }
}
