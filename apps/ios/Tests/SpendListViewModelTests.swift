import Connect
import MoneyRecapContract
import SwiftData
import XCTest
@testable import MyMoneyRecap

/// The Spend list (PRD §6.10, D59) from the BFF's answer recorded after three Spends today
/// (`Fixtures/list-spends.json`):
/// Bills Rp150.000 once and Snacks Rp43.000 twice, the second one with no note.
@MainActor
final class SpendListViewModelTests: XCTestCase {
    private func recorded() throws -> MoneyRecap_V1_SpendListView {
        let url = try XCTUnwrap(Bundle(for: Self.self).url(forResource: "list-spends", withExtension: "json"))
        return try MoneyRecap_V1_BffServiceListSpendsResponse(jsonString: String(contentsOf: url, encoding: .utf8))
            .spends
    }

    private func freshContext() throws -> ModelContext {
        try ModelContext(ModelContainer(
            for: HomeSnapshot.self, SpendListSnapshot.self,
            configurations: ModelConfiguration(isStoredInMemoryOnly: true)
        ))
    }

    private let locale = Locale(identifier: "id_ID")

    func testTheMonthsRowsAndTheCategoryDetail() async throws {
        let list = try recorded()
        let model = try SpendListViewModel(
            month: "2026-09",
            fetch: { _ in list },
            context: freshContext(),
            locale: locale
        )
        XCTAssertNil(model.list, "nothing stored yet")
        await model.refresh()
        XCTAssertEqual(model.label(model.month), "September 2026")
        let rows = model.list?.categories.map {
            "\(SpendListViewModel.glyph($0.category).label) \(model.count($0)) \(model.money($0.amount))"
        }
        XCTAssertEqual(rows, ["Bills 1 spend Rp150.000", "Snacks 2 spends Rp43.000"], "largest first (PRD §6.10)")
        XCTAssertNil(model.emptyLine)
        XCTAssertEqual(model.shareLine(.snacks), "Rp43.000 · 22% of September so far")
        let snacks = model.spends(.snacks).map { "\(model.title($0)), \(model.day($0)), \(model.money($0.amount))" }
        XCTAssertEqual(
            snacks,
            ["Snacks, 24 September, Rp18.000", "Kopi susu, 24 September, Rp25.000"],
            "newest first, the note or else the Category"
        )
    }

    /// B6 empty: this month counts the Streak (D43); a past month with Money In and no Spend names itself (ruling
    /// 2026-09-24). Picking the past month from the switcher asks for it and closes the switcher.
    func testBothEmptyMonths() async throws {
        var september = try recorded()
        september.categories = []
        september.spends = []
        september.months = ["2026-09", "2026-08"]
        september.streakDay = 23
        var august = september
        august.month = "2026-08"
        var asked: [String] = []
        let model = try SpendListViewModel(month: "2026-09", switching: true, fetch: { month in
            asked.append(month)
            return month == "2026-08" ? august : september
        }, context: freshContext(), locale: Locale(identifier: "en_GB"))
        await model.refresh()
        XCTAssertEqual(model.emptyLine, "Nothing spent yet. Streak day 23.")
        XCTAssertEqual(model.months.map(model.label), ["September 2026", "August 2026"])
        await model.pick("2026-08")
        XCTAssertFalse(model.switching)
        XCTAssertEqual(asked, ["2026-09", "2026-08"])
        XCTAssertEqual(model.emptyLine, "Nothing spent in August.")
    }

    /// A past month's share has no "so far"; a month opened once reads offline from its snapshot (IOS I4), which
    /// sign-out erases with the stored Home.
    func testAMonthOpenedOnceReadsOfflineUntilSignOut() async throws {
        let context = try freshContext()
        var august = try recorded()
        august.month = "2026-08"
        august.months = ["2026-09", "2026-08"]
        await SpendListViewModel(month: "2026-08", fetch: { _ in august }, context: context, locale: locale).refresh()
        let offline = SpendListViewModel(
            month: "2026-08",
            fetch: { _ in throw ConnectError(code: .unavailable, message: nil) },
            context: context,
            locale: locale
        )
        await offline.refresh()
        XCTAssertEqual(offline.list, august, "the stored month, not a blank list")
        XCTAssertEqual(offline.shareLine(.bills), "Rp150.000 · 78% of Agustus")
        XCTAssertNil(SpendListSnapshot.load("2026-09", from: context), "only the month opened is stored")
        HomeSnapshot.erase(from: context)
        XCTAssertNil(SpendListSnapshot.load("2026-08", from: context))
    }

    /// The month line under the switcher (owner ruling 2026-09-24): Spent and the count, "so far" only this month.
    func testTheMonthLineInItsFourForms() async throws {
        let many = try recorded()
        var one = many
        one.spends = [many.spends[2]]
        one.spent = many.spends[2].amount
        var none = many
        none.spends = []
        none.categories = []
        none.spent.amount = "0"
        var past = many
        past.month = "2026-08"
        past.months = ["2026-09", "2026-08"]
        var lines: [String?] = []
        for list in [many, one, none, past] {
            let model = try SpendListViewModel(
                month: list.month,
                fetch: { _ in list },
                context: freshContext(),
                locale: locale
            )
            await model.refresh()
            lines.append(model.monthLine)
        }
        XCTAssertEqual(lines, [
            "Rp193.000 so far, 3 spends", "Rp25.000 so far, 1 spend", "Rp0 so far, no spends", "Rp193.000, 3 spends",
        ])
    }
}
