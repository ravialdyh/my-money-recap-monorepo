import MoneyRecapContract
import SwiftData
import XCTest
@testable import MyMoneyRecap

/// The recorded BFF answer for the test user (`Fixtures/home-zero.json`, IOS I9) drives every case.
@MainActor
final class HomeViewModelTests: XCTestCase {
    private func recordedHome() throws -> MoneyRecap_V1_HomeView {
        let url = try XCTUnwrap(Bundle(for: Self.self).url(forResource: "home-zero", withExtension: "json"))
        return try MoneyRecap_V1_GetHomeResponse(jsonString: String(contentsOf: url, encoding: .utf8)).home
    }

    private func freshContext() throws -> ModelContext {
        ModelContext(try ModelContainer(for: HomeSnapshot.self, configurations: ModelConfiguration(isStoredInMemoryOnly: true)))
    }

    func testZeroHomeOmitsGreetingAndSavedChip() async throws {
        let home = try recordedHome()
        let model = HomeViewModel(fetch: { home }, context: try freshContext())
        XCTAssertTrue(model.isFirstLoad, "no snapshot yet: the skeleton")
        await model.refresh()
        XCTAssertFalse(model.isFirstLoad)
        XCTAssertNil(model.greeting, "display_name is empty before onboarding")
        XCTAssertFalse(model.showsSavedChip, "Money In is 0, so no Saved chip (D41)")
        XCTAssertEqual(model.home?.spent.amount, "0")
        XCTAssertEqual(model.heroEyebrow, "Spent so far")
        XCTAssertEqual(model.previewCopy, "Your first recap arrives on 1 October.")
        XCTAssertEqual(model.monthLabel, "September")
    }

    func testGreetingAndSavedChipAppearWithNameAndMoneyIn() async throws {
        var home = try recordedHome()
        home.displayName = "Maya"
        home.moneyIn.amount = "2400"
        home.saved = .with { $0.amount = "640"; $0.currencyCode = "GBP" }
        let model = HomeViewModel(fetch: { home }, context: try freshContext())
        await model.refresh()
        XCTAssertEqual(model.greeting, "Hi, Maya")
        XCTAssertTrue(model.showsSavedChip)
    }

    func testStoredSnapshotRendersBeforeNetworkAndSurvivesAFailedRefresh() async throws {
        let context = try freshContext()
        try HomeSnapshot.replace(with: try recordedHome(), in: context)
        let model = HomeViewModel(fetch: { throw URLError(.cannotConnectToHost) }, context: context)
        XCTAssertEqual(model.home?.month, "2026-09", "the snapshot is on screen before any call completes (I4)")
        await model.refresh()
        XCTAssertEqual(model.home?.month, "2026-09", "a failed refresh keeps the last snapshot")
        XCTAssertFalse(model.isRefreshing)
    }

    func testSuccessfulRefreshReplacesTheSnapshotWhole() async throws {
        let context = try freshContext()
        try HomeSnapshot.replace(with: try recordedHome(), in: context)
        var updated = try recordedHome()
        updated.month = "2026-10"
        await HomeViewModel(fetch: { updated }, context: context).refresh()
        let rows = try context.fetch(FetchDescriptor<HomeSnapshot>())
        XCTAssertEqual(rows.count, 1)
        XCTAssertEqual(try MoneyRecap_V1_HomeView(serializedBytes: rows[0].bytes).month, "2026-10")
    }

    func testMoneyPrintsWholeAmountsWithoutAZeroFraction() {
        XCTAssertEqual(Money.format("0", currency: "GBP"), Decimal(0).formatted(.currency(code: "GBP").precision(.fractionLength(0))))
        XCTAssertEqual(Money.format("12.40", currency: "GBP"), Decimal(string: "12.40")!.formatted(.currency(code: "GBP")))
    }
}
