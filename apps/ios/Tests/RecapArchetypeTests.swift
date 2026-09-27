import XCTest
@testable import MyMoneyRecap

/// Page 4, C4 and P4 (#156), from recorded `GetRecap` answers: Budi's rupiah Augusts (Squirrel, Hustler, overspent
/// Treat-Yourself, Slow Burner), Maya's pound July and no-target May, the partial months. Pictures are evidence.
@MainActor
final class RecapArchetypeTests: XCTestCase {
    private struct Case {
        let name: String
        let archetype: ArchetypePill.Kind
        let share: String
        let figures: [String]
    }

    func testEachArchetypeNamesItselfWithItsMascotSavedAgainstSpentAndShare() async throws {
        let cases = [
            Case(
                name: "recap-squirrel", archetype: .squirrel, share: "Share your Squirrel",
                figures: ["IDR\u{00A0}5,600,000", "IDR\u{00A0}9,400,000"]
            ),
            Case(
                name: "recap-no-run", archetype: .hustler, share: "Share your Hustler",
                figures: ["IDR\u{00A0}14,690,000", "IDR\u{00A0}310,000"]
            ),
            Case(
                name: "recap-target-missed", archetype: .treatYourself, share: "Share your Treat-Yourself",
                figures: ["£310", "£2,090"]
            ),
            Case(
                name: "recap-slow-burner", archetype: .slowBurner, share: "Share your Slow Burner",
                figures: ["IDR\u{00A0}2,400,000", "IDR\u{00A0}12,600,000"]
            ),
        ]
        for test in cases {
            let model = try await RecapViewModelTests.loaded(test.name)
            XCTAssertEqual(model.pages, [.cover, .kept, .goal, .archetype], test.name)
            XCTAssertEqual(model.archetype, test.archetype, test.name)
            XCTAssertEqual(model.share, test.share, test.name)
            XCTAssertEqual(model.savedAgainstSpent.map(\.label), ["Saved", "Spent"], test.name)
            XCTAssertEqual(model.savedAgainstSpent.map(\.display), test.figures, test.name)
        }
        let model = try await RecapViewModelTests.loaded("recap-squirrel")
        XCTAssertEqual(model.personalityLead, "This month, Budi, you were")
        XCTAssertEqual(
            ArchetypePill.Kind.allCases.map(RecapArchetypePage.mascot),
            ["ill-squirrel", "ill-hustler", "ill-treat", "ill-slowburner"]
        )
    }

    func testAnOverspentMonthPrintsItsNegativeSavedAsItIsAndDrawsItEmpty() async throws {
        let model = try await RecapViewModelTests.loaded("recap-overspent")
        XCTAssertEqual(model.archetype, .treatYourself)
        XCTAssertEqual(model.savedAgainstSpent.first?.display, "-IDR\u{00A0}200,000")
        XCTAssertEqual(CompareBars.fractions(model.savedAgainstSpent.map(\.value), scale: .total), [0, 1])
    }

    func testWithNoSavingTargetThereIsNoArchetype() async throws {
        let model = try await RecapViewModelTests.loaded("recap-no-target")
        XCTAssertEqual(model.pages, [.cover, .kept, .goal, .archetype])
        XCTAssertNil(model.archetype)
        XCTAssertEqual(model.savedAgainstSpent.map(\.display), ["£640", "£1,760"])
    }

    /// P4 in each situation: ledger's closing line, no Archetype.
    func testP4ClosesEachPartialSituationWithLedgersLine() async throws {
        let cases = [
            ("recap-preview", "Your recap arrives on 1 October."),
            ("recap-sign-up", "Your first full recap arrives on 1 October."),
            ("recap-no-money-in", "Set your monthly salary to get a full recap."),
        ]
        for (name, line) in cases {
            let model = try await RecapViewModelTests.loaded(name)
            XCTAssertEqual(model.pages, [.partialCover, .streak, .partialGoal, .closing], name)
            XCTAssertEqual(model.answer?.recap.closingLine, line, name)
            XCTAssertNil(model.archetype, name)
        }
    }

    private static let archetypes = ["recap-squirrel", "recap-no-run", "recap-target-missed", "recap-slow-burner"]

    private func page(_ name: String) async throws -> RecapPagerView {
        let model = try await RecapViewModelTests.loaded(name)
        model.page = 3
        return RecapPagerView(model: model)
    }

    private func keep(_ png: Data?, _ name: String) throws {
        let attachment = try XCTAttachment(data: XCTUnwrap(png), uniformTypeIdentifier: "public.png")
        attachment.name = name
        attachment.lifetime = .keepAlways
        add(attachment)
    }

    func testC4ForEachArchetypeWithoutOneAndP4DrawApart() async throws {
        let names = Self.archetypes + [
            "recap-overspent", "recap-no-target", "recap-preview", "recap-sign-up", "recap-no-money-in",
        ]
        var pictures: [Data?] = []
        for name in names {
            try await pictures.append(picture(page(name)))
        }
        XCTAssertEqual(Set(pictures).count, names.count, "each case draws its own page")
        for (name, png) in zip(names, pictures) {
            try keep(png, "page-4-" + name.replacingOccurrences(of: "recap-", with: ""))
        }
    }

    /// Every C4 fits upright at 393 by 852 with the pills in full, the tallest mascot, the Hustler's, included.
    func testEachArchetypesPageFitsUpright() async throws {
        for name in Self.archetypes {
            let overflow = try await overflow(page(name), size: CGSize(width: 393, height: 852))
            XCTAssertEqual(overflow, 0, "\(name) fits upright")
        }
    }

    func testPage4HeldSidewaysScrollsRatherThanClips() async throws {
        let sideways = CGSize(width: 852, height: 393)
        for name in ["recap-no-run", "recap-no-target"] {
            let page = try await page(name)
            XCTAssertGreaterThan(overflow(page, size: sideways), 0, "\(name) scrolls sideways")
            try keep(
                picture(page, size: sideways),
                "page-4-sideways-" + name.replacingOccurrences(of: "recap-", with: "")
            )
        }
    }
}
