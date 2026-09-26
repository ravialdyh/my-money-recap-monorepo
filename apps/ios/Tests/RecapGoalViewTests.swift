import MoneyRecapContract
import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// Screen 3 as drawn, opened on its page: C3 met, missed, with no Spend and with no Saving Target, and P3 before and
/// after Payday, each unlike the others; held sideways it scrolls rather than clips. The pictures are kept as evidence.
@MainActor
final class RecapGoalViewTests: XCTestCase {
    private func page(_ name: String) async throws -> RecapPagerView {
        let model = try await RecapGoalViewModelTests.loaded(RecapViewModelTests.recorded(name))
        model.page = try XCTUnwrap(model.pages.firstIndex { $0 == .goal || $0 == .partialGoal })
        return RecapPagerView(model: model)
    }

    private func keep(_ png: Data?, _ name: String) throws {
        let attachment = try XCTAttachment(data: XCTUnwrap(png), uniformTypeIdentifier: "public.png")
        attachment.name = name
        attachment.lifetime = .keepAlways
        add(attachment)
    }

    func testEachStateOfScreenThreeDrawsApart() async throws {
        let states = [
            ("recap-target-met", "c3-target-met"),
            ("recap-target-missed", "c3-target-missed"),
            ("recap-no-spends", "c3-no-spends"),
            ("recap-no-target", "c3-no-target"),
            ("recap-before-payday", "p3-before-payday"),
            ("recap-after-payday", "p3-after-payday"),
            ("recap-no-money-in", "p3-past-partial-month"),
        ]
        var pictures: [Data] = []
        for (fixture, name) in states {
            let view = try await page(fixture)
            let png = try XCTUnwrap(picture(view))
            XCTAssertFalse(pictures.contains(png), "\(name) draws unlike the states before it")
            pictures.append(png)
            try keep(png, name)
        }
        let cover = try await pager(RecapViewModelTests.recorded("recap-target-met"))
        XCTAssertFalse(try pictures.contains(XCTUnwrap(picture(cover))), "the Recap opens on its cover, not screen 3")
    }

    func testScreenThreeFitsUprightAndScrollsHeldSideways() async throws {
        let met = try await page("recap-target-met")
        XCTAssertEqual(overflow(met, size: CGSize(width: 393, height: 852)), 0, "upright it fits")
        let sideways = CGSize(width: 852, height: 393)
        XCTAssertGreaterThan(overflow(met, size: sideways), 0, "the page is taller than the screen and scrolls")
        try keep(picture(met, size: sideways), "c3-sideways")
    }

    private func pager(_ answer: MoneyRecap_V1_RecapView) async -> RecapPagerView {
        await RecapPagerView(model: RecapGoalViewModelTests.loaded(answer))
    }
}
