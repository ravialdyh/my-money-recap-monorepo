import Connect
import MoneyRecapContract
import SwiftData
import XCTest
@testable import MyMoneyRecap

/// A Recap opened once opens again with no signal (PRD D61, IOS I4): the stored copy is drawn at once, the answer
/// replaces it, and a failure keeps it. The erase at sign-out and on a refused Home is `HomeViewModelTests`'.
@MainActor
final class RecapSnapshotTests: XCTestCase {
    private func freshContext() throws -> ModelContext {
        try ModelContext(ModelContainer(
            for: HomeSnapshot.self, SpendListSnapshot.self, RecapSnapshot.self,
            configurations: ModelConfiguration(isStoredInMemoryOnly: true)
        ))
    }

    private let failing: RecapViewModel.Fetch = { _ in throw ConnectError(code: .unavailable, message: nil) }

    func testTheAnswerIsStoredAndDrawnAtOnceNextTimeWithNoSignal() async throws {
        let context = try freshContext()
        let preview = try RecapViewModelTests.recorded("recap-preview")
        let online = RecapViewModel(month: "2026-09", fetch: { _ in preview }, context: context)
        await online.load()
        let offline = RecapViewModel(month: "2026-09", fetch: failing, context: context)
        XCTAssertEqual(offline.answer, preview, "drawn before any answer")
        XCTAssertEqual(offline.pages, online.pages, "the same pages as the answer drew")
        await offline.load()
        XCTAssertEqual(offline.answer, preview, "a failure keeps what is shown")
        XCTAssertNil(RecapSnapshot.load("2026-08", from: context), "only the month opened is stored")
    }

    func testTheAnswerReplacesTheStoredCopyWhole() async throws {
        let context = try freshContext()
        var stale = try RecapViewModelTests.recorded("recap-squirrel")
        stale.displayName = "Old name"
        try RecapSnapshot.replace(with: stale, in: context)
        let fresh = try RecapViewModelTests.recorded("recap-squirrel")
        let model = RecapViewModel(month: "2026-08", fetch: { _ in fresh }, context: context)
        XCTAssertEqual(model.nameChip, "Old name's recap", "the stored copy at once")
        await model.load()
        XCTAssertEqual(model.nameChip, "Budi's recap")
        XCTAssertEqual(RecapSnapshot.load("2026-08", from: context), fresh)
        XCTAssertEqual(try context.fetch(FetchDescriptor<RecapSnapshot>()).count, 1, "one row per month")
    }

    func testAnAnswerForAnotherMonthIsNeitherStoredNorDrawn() async throws {
        let context = try freshContext()
        let august = try RecapViewModelTests.recorded("recap-squirrel")
        let model = RecapViewModel(month: "2026-09", fetch: { _ in august }, context: context)
        await model.load()
        XCTAssertNil(model.answer)
        XCTAssertNil(RecapSnapshot.load("2026-08", from: context))
    }
}
