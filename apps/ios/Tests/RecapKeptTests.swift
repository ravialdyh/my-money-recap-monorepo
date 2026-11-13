import XCTest
@testable import MyMoneyRecap

/// Page 2, C2 and P2 (#154): Maya's September from ledger's answer key (the stack cannot close September before
/// 1 October), #155's no-target May, and Budi's recorded Augusts. Pictures are evidence.
@MainActor
final class RecapKeptTests: XCTestCase {
    func testMayasSquirrelMonthKeptAgainstHerTargetWithTheRunLit() async throws {
        let model = try await RecapViewModelTests.loaded("recap-maya-september")
        XCTAssertEqual(model.pages, [.cover, .kept, .goal, .archetype])
        XCTAssertEqual(model.monthName, "September")
        XCTAssertEqual(model.keptHeadline, "You kept £640 of £2,400. That's 133% of your target.")
        XCTAssertEqual(model.keptRows?.map(\.label), ["Kept", "Target"])
        XCTAssertEqual(model.keptRows?.map(\.display), ["£640", "£480"])
        XCTAssertEqual(model.streakDays, 9)
        XCTAssertEqual(model.streakTitle, "Longest no-spend streak")
        XCTAssertEqual(model.streakLine, "12 to 20 September. New badge unlocked.")
        XCTAssertEqual(model.windowDays, 30)
        XCTAssertEqual(model.runDays, 12 ... 20)
    }

    func testAOneDayRunPrintsItsOneDate() async throws {
        let model = try await RecapViewModelTests.loaded("recap-one-day-run")
        XCTAssertEqual(model.streakDays, 1)
        XCTAssertEqual(model.streakLine, "12 August. Beat it next month.")
        XCTAssertEqual(model.runDays, 12 ... 12)
    }

    func testWithNoSpendFreeDayTheBadgeCopyStandsAlone() async throws {
        let model = try await RecapViewModelTests.loaded("recap-no-run")
        XCTAssertEqual(model.streakDays, 0)
        XCTAssertEqual(model.streakLine, "Beat it next month.")
        XCTAssertNil(model.runDays)
        XCTAssertEqual(model.windowDays, 31)
    }

    func testWithNoSavingTargetTheHeadlineStopsAndTheTargetCardStands() async throws {
        let model = try await RecapViewModelTests.loaded("recap-no-target")
        XCTAssertEqual(model.keptHeadline, "You kept £640 of £2,400.")
        XCTAssertNil(model.keptRows)
    }

    func testAnOverspentMonthPrintsItsNegativeSavedAsItIs() async throws {
        let model = try await RecapViewModelTests.loaded("recap-overspent")
        XCTAssertEqual(
            model.keptHeadline,
            "You kept -IDR\u{00A0}2,000,000 of IDR\u{00A0}150,000,000. That's 0% of your target."
        )
        XCTAssertEqual(model.keptRows?.first?.display, "-IDR\u{00A0}2,000,000")
    }

    /// P2: the dates as the title and the badge copy under it; in the current month the dots stop at today (the
    /// window ends on the recording's 26 September), and a sign-up month's window starts on its sign-up day.
    func testP2InTheCurrentMonthAndInAPartialMonth() async throws {
        let cases = [
            ("recap-preview", "1 to 26 September", "New badge unlocked.", 26, 1 ... 26),
            ("recap-sign-up", "30 to 31 August", "Beat it next month.", 5, 4 ... 5),
        ]
        for (name, title, line, window, run) in cases {
            let model = try await RecapViewModelTests.loaded(name)
            XCTAssertEqual(model.pages, [.partialCover, .streak, .partialGoal, .closing], name)
            XCTAssertEqual(model.keptHeadline, "Your longest no-spend streak", name)
            XCTAssertNil(model.keptRows, name)
            XCTAssertEqual(model.streakTitle, title, name)
            XCTAssertEqual(model.streakLine, line, name)
            XCTAssertEqual(model.windowDays, window, name)
            XCTAssertEqual(model.runDays, run, name)
        }
    }

    /// The run follows the locale's day and month order; a one-day run is the single date.
    func testTheRunDatesFollowTheLocale() async throws {
        let cases = [
            ("recap-maya-september", "en_GB", "12 to 20 September"),
            ("recap-maya-september", "en_US", "September 12 to 20"),
            ("recap-one-day-run", "en_GB", "12 August"),
            ("recap-one-day-run", "en_US", "August 12"),
        ]
        for (name, locale, dates) in cases {
            let answer = try RecapViewModelTests.recorded(name)
            let model = RecapViewModel(
                month: answer.recap.month,
                fetch: { _ in answer },
                locale: Locale(identifier: locale)
            )
            await model.load()
            XCTAssertEqual(model.runDates, dates, "\(name) \(locale)")
        }
    }

    func testCompareBarsScaleToTheLargerOrTheWholeAndDrawANegativeEmpty() {
        XCTAssertEqual(CompareBars.fractions([640, 480], scale: .max), [1, 0.75])
        XCTAssertEqual(CompareBars.fractions([640, 1760], scale: .total), [0.26666666666666666, 0.7333333333333333])
        XCTAssertEqual(CompareBars.fractions([-200, 480], scale: .max), [0, 1])
        XCTAssertEqual(CompareBars.fractions([0, 0], scale: .max), [0, 0])
    }

    private func page(_ name: String) async throws -> RecapPagerView {
        let model = try await RecapViewModelTests.loaded(name)
        model.page = 1
        return RecapPagerView(model: model)
    }

    private func keep(_ png: Data?, _ name: String) throws {
        let attachment = try XCTAttachment(data: XCTUnwrap(png), uniformTypeIdentifier: "public.png")
        attachment.name = name
        attachment.lifetime = .keepAlways
        add(attachment)
    }

    func testC2AndItsVariantsAndP2DrawApart() async throws {
        let names = [
            "recap-maya-september", "recap-no-target", "recap-one-day-run", "recap-no-run", "recap-overspent",
            "recap-preview", "recap-sign-up",
        ]
        var pictures: [Data?] = []
        for name in names {
            try await pictures.append(picture(page(name)))
        }
        XCTAssertEqual(Set(pictures).count, names.count, "each case draws its own page")
        let cover = try await picture(RecapPagerView(model: RecapViewModelTests.loaded("recap-maya-september")))
        XCTAssertFalse(pictures.contains(cover), "page 2 is on screen, not the cover")
        for (name, png) in zip(names, pictures) {
            try keep(png, "page-2-" + name.replacingOccurrences(of: "recap-", with: ""))
        }
    }

    func testPage2HeldSidewaysScrollsRatherThanClips() async throws {
        let sideways = CGSize(width: 852, height: 393)
        let page = try await page("recap-maya-september")
        XCTAssertGreaterThan(overflow(page, size: sideways), 0, "the page is taller than the screen and scrolls")
        XCTAssertEqual(overflow(page, size: CGSize(width: 393, height: 852)), 0, "upright it fits")
        try keep(picture(page, size: sideways), "page-2-sideways")
    }
}
