import Connect
import MoneyRecapContract
import SwiftData
import XCTest
@testable import MyMoneyRecap

/// The recorded BFF answer for the test Member after steps 1 and 2 (`Fixtures/home-zero.json`, IOS I9) drives every
/// case.
@MainActor
final class HomeViewModelTests: XCTestCase {
    private func recordedHome() throws -> MoneyRecap_V1_HomeView {
        let url = try XCTUnwrap(Bundle(for: Self.self).url(forResource: "home-zero", withExtension: "json"))
        return try MoneyRecap_V1_GetHomeResponse(jsonString: String(contentsOf: url, encoding: .utf8)).home
    }

    private func freshContext() throws -> ModelContext {
        try ModelContext(ModelContainer(
            for: HomeSnapshot.self,
            configurations: ModelConfiguration(isStoredInMemoryOnly: true)
        ))
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
