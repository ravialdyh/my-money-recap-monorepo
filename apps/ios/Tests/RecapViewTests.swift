import MoneyRecapContract
import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// Page 1 as drawn: C1 and P1 differ from each other and from the loading state; held sideways the page scrolls
/// rather than clips; X4, Budi's rupiah cover, fits upright with its chips wrapped. The pictures are kept as evidence.
@MainActor
final class RecapViewTests: XCTestCase {
    private let upright = CGSize(width: 393, height: 852)

    private func pager(
        _ answer: MoneyRecap_V1_RecapView?,
        locale: Locale = .init(identifier: "en_GB")
    ) async -> RecapPagerView {
        let model = RecapViewModel(month: answer?.recap.month ?? "2026-08", fetch: { _ in
            guard let answer else { throw CancellationError() }
            return answer
        }, locale: locale)
        await model.load()
        return RecapPagerView(model: model)
    }

    private func keep(_ png: Data?, _ name: String) throws {
        let attachment = try XCTAttachment(data: XCTUnwrap(png), uniformTypeIdentifier: "public.png")
        attachment.name = name
        attachment.lifetime = .keepAlways
        add(attachment)
    }

    func testTheCoverAndThePartialCoverDrawApartFromTheLoadingState() async throws {
        let loading = await picture(pager(nil))
        let cover = try await picture(pager(RecapViewModelTests.recorded("recap-squirrel")))
        let partial = try await picture(pager(RecapViewModelTests.recorded("recap-preview")))
        XCTAssertNotEqual(loading, cover)
        XCTAssertNotEqual(loading, partial)
        XCTAssertNotEqual(cover, partial)
        XCTAssertNotEqual(loading, picture(Screen {}), "the loading state draws the cover illustration")
        try keep(loading, "recap-loading")
        try keep(cover, "c1-cover")
        try keep(partial, "p1-month-so-far")
    }

    func testAPageHeldSidewaysScrollsRatherThanClips() async throws {
        let sideways = CGSize(width: 852, height: 393)
        let page = try await pager(RecapViewModelTests.recorded("recap-squirrel"))
        XCTAssertGreaterThan(overflow(page, size: sideways), 0, "the page is taller than the screen and scrolls")
        XCTAssertEqual(overflow(page, size: upright), 0, "upright it fits")
        try keep(picture(page, size: sideways), "c1-sideways")
    }

    /// X4 (PRD D45): Budi's recorded month in rupiah on an id-ID phone. Money In steps down to 40pt, the Saved and
    /// Spent
    /// chips wrap, and the illustration narrows from 310 to 260pt so the page dots stay on screen.
    func testTheRupiahCoverFitsWithItsChipsWrapped() async throws {
        let page = try await pager(RecapViewModelTests.recorded("recap-squirrel"), locale: Locale(identifier: "id_ID"))
        XCTAssertEqual(page.model.hero, "Rp15.000.000")
        XCTAssertEqual(page.model.chips.map(\.text), ["Saved Rp5.600.000", "Spent Rp9.400.000"])
        let card = upright.width - 2 * 24 - 2 * Tokens.Space._7
        XCTAssertEqual(StatHero.size(for: page.model.hero, in: card), Tokens.Size.statHeroS)
        XCTAssertEqual(overflow(page, size: upright), 0, "the chips wrap and the page still fits")
        try keep(picture(page), "x4-rupiah-cover")
    }
}
