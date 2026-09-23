import Connect
import MoneyRecapContract
import SwiftData
import XCTest
@testable import MyMoneyRecap

/// The recorded BFF answer for the test Member after steps 1 and 2 (`Fixtures/home-zero.json`, IOS I9) drives every
/// case.
@MainActor
final class HomeViewModelTests: XCTestCase {
    private func recordedHome(_ name: String = "home-zero") throws -> MoneyRecap_V1_HomeView {
        let url = try XCTUnwrap(Bundle(for: Self.self).url(forResource: name, withExtension: "json"))
        return try MoneyRecap_V1_GetHomeResponse(jsonString: String(contentsOf: url, encoding: .utf8)).home
    }

    private func freshContext() throws -> ModelContext {
        try ModelContext(ModelContainer(
            for: HomeSnapshot.self,
            configurations: ModelConfiguration(isStoredInMemoryOnly: true)
        ))
    }

    /// Sign-out erases what the device holds (PRD §9), and the stored Home is the only setup progress it holds: it is
    /// what opens Home with no round trip (spec #92).
    func testEraseEmptiesTheStoreAsSignOutDoes() throws {
        let context = try freshContext()
        try HomeSnapshot.replace(with: recordedHome(), in: context)
        XCTAssertNotNil(HomeSnapshot.load(from: context))
        HomeSnapshot.erase(from: context)
        XCTAssertNil(HomeSnapshot.load(from: context))
    }

    func testZeroHomeGreetsTheMemberWithoutASavedChip() async throws {
        let home = try recordedHome()
        let model = try HomeViewModel(fetch: { home }, context: freshContext(), locale: Locale(identifier: "en_GB"))
        XCTAssertNil(model.home, "no snapshot yet: the skeleton")
        await model.refresh()
        XCTAssertNotNil(model.home)
        XCTAssertEqual(model.greeting, "Hi, Test", "Home is only served once step 1 has the name (spec #92)")
        XCTAssertEqual(model.home?.currencyCode, "IDR", "every figure in the Currency step 2 saved")
        XCTAssertFalse(model.showsSavedChip, "Money In is 0, so no Saved chip (D41)")
        XCTAssertEqual(model.home?.spent.amount, "0")
        XCTAssertEqual(model.heroEyebrow, "Spent so far")
        XCTAssertEqual(model.previewCopy, "Your first recap arrives on 1 October.")
        XCTAssertEqual(model.monthLabel, "September")
    }

    /// Home's past-recap rows print each month and the label ledger answered, its Archetype or "Partial" (PRD §6.8,
    /// §8.11), from the BFF's answer recorded with a Hustler August and a July of Spends with no Salary.
    func testPastRecapRowsShowLedgersLabels() async throws {
        let home = try recordedHome("home-recaps")
        let model = try HomeViewModel(fetch: { home }, context: freshContext(), locale: Locale(identifier: "en_GB"))
        await model.refresh()
        let rows = model.home?.recapMonths.map { "\(model.recapLabel($0)): \($0.label)" }
        XCTAssertEqual(rows, ["August 2026: Hustler", "July 2026: Partial"], "newest first, as ledger lists them")
    }

    /// PRD §6.8's countdown before the effective Payday, in the singular on the day before (§12, criterion 4).
    func testTheEyebrowCountsTheDaysToTheSalary() async throws {
        var home = try recordedHome()
        home.salary.daysUntilPayday = 15
        let model = try HomeViewModel(fetch: { home }, context: freshContext())
        await model.refresh()
        XCTAssertEqual(model.heroEyebrow, "Spent so far · salary lands in 15 days")
        home.salary.daysUntilPayday = 1
        await model.refresh()
        XCTAssertEqual(model.heroEyebrow, "Spent so far · salary lands in 1 day")
    }

    func testGreetingAndSavedChipAppearWithNameAndMoneyIn() async throws {
        var home = try recordedHome()
        home.displayName = "Maya"
        home.moneyIn.amount = "3000000"
        home.saved = .with { $0.amount = "600000"; $0.currencyCode = "IDR" }
        let model = try HomeViewModel(fetch: { home }, context: freshContext())
        await model.refresh()
        XCTAssertEqual(model.greeting, "Hi, Maya")
        XCTAssertTrue(model.showsSavedChip)
    }

    func testStoredSnapshotRendersBeforeNetworkAndSurvivesAFailedRefresh() async throws {
        let context = try freshContext()
        try HomeSnapshot.replace(with: recordedHome(), in: context)
        var sent = 0
        let model = HomeViewModel(
            fetch: { throw URLError(.cannotConnectToHost) },
            context: context,
            onboardingNeeded: { sent += 1 }
        )
        XCTAssertEqual(model.home?.month, "2026-09", "the snapshot is on screen before any call completes (I4)")
        await model.refresh()
        XCTAssertEqual(model.home?.month, "2026-09", "a failed refresh keeps the last snapshot")
        XCTAssertEqual(sent, 0, "no signal is not a reason to leave Home")
    }

    /// failed_precondition means step 1 or 2 is unanswered (spec #92): the stored Home, which may still carry the
    /// currency the backend used to invent, is deleted and onboarding opens.
    func testARefusedHomeDeletesTheStoredHomeAndOpensOnboarding() async throws {
        let context = try freshContext()
        try HomeSnapshot.replace(with: recordedHome(), in: context)
        var sent = 0
        let refused = ConnectError(code: .failedPrecondition, message: "onboarding steps 1 and 2 are unanswered")
        await HomeViewModel(fetch: { throw refused }, context: context, onboardingNeeded: { sent += 1 }).refresh()
        XCTAssertEqual(sent, 1)
        XCTAssertNil(HomeSnapshot.load(from: context), "no stored Home to open on at the next launch")
    }

    func testSuccessfulRefreshReplacesTheSnapshotWhole() async throws {
        let context = try freshContext()
        try HomeSnapshot.replace(with: recordedHome(), in: context)
        var updated = try recordedHome()
        updated.month = "2026-10"
        await HomeViewModel(fetch: { updated }, context: context).refresh()
        let rows = try context.fetch(FetchDescriptor<HomeSnapshot>())
        XCTAssertEqual(rows.count, 1)
        XCTAssertEqual(try MoneyRecap_V1_HomeView(serializedBytes: rows[0].bytes).month, "2026-10")
    }
}
