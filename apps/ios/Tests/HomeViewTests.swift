import MoneyRecapContract
import SwiftData
import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// Home as drawn: with no stored answer yet it is the §6.19 first-load skeleton, the two cards at their filled
/// heights, not a blank ground and not the zero Home it becomes; a set Target fills its card in place of the prompt.
@MainActor
final class HomeViewTests: XCTestCase {
    func testTheFirstLoadIsTheSkeleton() throws {
        let context = try ModelContext(ModelContainer(
            for: HomeSnapshot.self, SpendListSnapshot.self,
            configurations: ModelConfiguration(isStoredInMemoryOnly: true)
        ))
        let waiting = HomeViewModel(fetch: {
            try await Task.sleep(for: .seconds(3600))
            return MoneyRecap_V1_HomeView()
        }, context: context)
        XCTAssertNil(waiting.home)
        let skeleton = picture(HomeView(model: waiting))
        let zero = MoneyRecap_V1_HomeView.with {
            $0.displayName = "Test"
            $0.currencyCode = "IDR"
            $0.spent.amount = "0"
            $0.moneyIn.amount = "0"
        }
        try HomeSnapshot.replace(with: zero, in: context)
        let loaded = picture(HomeView(model: HomeViewModel(fetch: { zero }, context: context)))
        XCTAssertNotNil(skeleton)
        XCTAssertNotEqual(skeleton, loaded, "the skeleton is not the zero Home")
        XCTAssertNotEqual(skeleton, picture(Screen {}), "the skeleton draws its two cards")
    }

    func testASetTargetFillsItsCardInPlaceOfThePrompt() throws {
        let context = try ModelContext(ModelContainer(
            for: HomeSnapshot.self, SpendListSnapshot.self,
            configurations: ModelConfiguration(isStoredInMemoryOnly: true)
        ))
        let unset = MoneyRecap_V1_HomeView.with {
            $0.displayName = "Test"
            $0.currencyCode = "IDR"
            $0.spent.amount = "0"
            $0.moneyIn.amount = "0"
            $0.salary = .with { $0.amount.amount = "15000000"; $0.payday = "25"; $0.daysUntilPayday = 5 }
        }
        let set = {
            var home = unset
            home.savingTarget = .with {
                $0.percent = 20
                $0.amount.amount = "3000000"
                $0.progressSentence = "Nothing to keep yet. Your salary lands on the 25th."
            }
            return home
        }()
        try HomeSnapshot.replace(with: unset, in: context)
        let prompt = picture(HomeView(model: HomeViewModel(fetch: { unset }, context: context)))
        try HomeSnapshot.replace(with: set, in: context)
        let filled = picture(HomeView(model: HomeViewModel(fetch: { set }, context: context)))
        XCTAssertNotEqual(prompt, filled)
    }

    /// X3 (PRD D45): Budi's rupiah Home on an id-ID phone. The hero steps down to 40pt, and every other line fits its
    /// row whole, the stat chips wrapping and "{Saved} of {target}" under its title. The picture is kept as evidence.
    func testTheRupiahHomeFitsWithoutClipping() throws {
        let context = try ModelContext(ModelContainer(
            for: HomeSnapshot.self, SpendListSnapshot.self,
            configurations: ModelConfiguration(isStoredInMemoryOnly: true)
        ))
        let rupiah = { (amount: String) in MoneyRecap_V1_Money.with { $0.amount = amount; $0.currencyCode = "IDR" } }
        let budi = MoneyRecap_V1_HomeView.with {
            $0.displayName = "Budi"
            $0.avatarKey = "owl"
            $0.currencyCode = "IDR"
            $0.month = "2026-09"
            $0.spent = rupiah("9400000")
            $0.moneyIn = rupiah("15000000")
            $0.saved = rupiah("5600000")
            $0.salary = .with { $0.amount = rupiah("15000000"); $0.payday = "25" }
            $0.savingTarget = .with {
                $0.percent = 20
                $0.amount = rupiah("3000000")
                $0.progressSentence = "Keep spending under Rp520.000 a day and you'll make it."
            }
            $0.firstRecapOn = "2026-10-01"
        }
        try HomeSnapshot.replace(with: budi, in: context)
        let model = HomeViewModel(fetch: { budi }, context: context, locale: Locale(identifier: "id_ID"))
        XCTAssertEqual(model.money(budi.spent), "Rp9.400.000")

        let card = 393 - 2 * 24 - 2 * Tokens.Space._7
        let width = { (line: String, font: UIFont) in (line as NSString).size(withAttributes: [.font: font]).width }
        XCTAssertEqual(StatHero.size(for: "Rp9.400.000", in: card), Tokens.Size.statHeroS, "80 and 56pt do not fit")
        XCTAssertLessThanOrEqual(width("Rp9.400.000", AmountField.font(of: Tokens.Size.statHeroS)), card)
        let nunito = { (size: CGFloat) in
            UIFont(
                descriptor: UIFontDescriptor(fontAttributes: [.family: Tokens.FontFamily.text])
                    .addingAttributes([.traits: [UIFontDescriptor.TraitKey.weight: UIFont.Weight.heavy]]),
                size: size
            )
        }
        let label = nunito(Tokens.TypeRamp.labelS.size)
        for chip in ["Money in Rp15.000.000", "Saved Rp5.600.000"] {
            XCTAssertLessThanOrEqual(width(chip, label) + 2 * Tokens.Space._5 + 14, card, "\(chip) fits a line alone")
        }
        XCTAssertLessThanOrEqual(try width(XCTUnwrap(model.targetProgress), nunito(Tokens.TypeRamp.labelM.size)), card)

        let drawn = try XCTUnwrap(picture(HomeView(model: model)))
        let attachment = XCTAttachment(data: drawn, uniformTypeIdentifier: "public.png")
        attachment.name = "x3-rupiah-home"
        attachment.lifetime = .keepAlways
        add(attachment)
    }

    /// B1's coral Banner above the target card, in each of D53's two wordings, and gone when dismissed: the dismissed
    /// Home draws as the one with no warning. The pictures are kept as evidence.
    func testTheWarningBannerDrawsAboveTheTargetUntilDismissed() throws {
        let context = try ModelContext(ModelContainer(
            for: HomeSnapshot.self,
            configurations: ModelConfiguration(isStoredInMemoryOnly: true)
        ))
        let calm = MoneyRecap_V1_HomeView.with {
            $0.displayName = "Maya"
            $0.avatarKey = "fox"
            $0.currencyCode = "GBP"
            $0.month = "2026-09"
            $0.spent = .with { $0.amount = "2000"; $0.currencyCode = "GBP" }
            $0.moneyIn = .with { $0.amount = "0"; $0.currencyCode = "GBP" }
            $0.salary = .with { $0.amount.amount = "2400"; $0.payday = "25"; $0.daysUntilPayday = 5 }
            $0.savingTarget = .with {
                $0.percent = 20
                $0.amount.amount = "480"
                $0.progressSentence = "Nothing to keep yet. Your salary lands on the 25th."
            }
        }
        let draw = { (home: MoneyRecap_V1_HomeView, dismissed: Bool) -> Data? in
            try? HomeSnapshot.replace(with: home, in: context)
            let model = HomeViewModel(
                fetch: { home }, context: context, locale: Locale(identifier: "en_GB"), dismissals: WarningDismissals()
            )
            if dismissed {
                model.dismissWarning()
            }
            return self.picture(HomeView(model: model))
        }
        var beforePayday = calm
        beforePayday.targetWarning.message = "You've spent £2,000 of the £1,920 you meant to spend this month."
        var afterPayday = calm
        afterPayday.salary.daysUntilPayday = 0
        afterPayday.moneyIn.amount = "2400"
        afterPayday.targetWarning.message = "You've passed your saving line. £80 over."

        let none = draw(calm, false)
        let before = draw(beforePayday, false)
        let after = draw(afterPayday, false)
        XCTAssertNotEqual(before, none, "the banner draws")
        XCTAssertNotEqual(before, after)
        XCTAssertEqual(draw(beforePayday, true), none, "dismissed, Home draws as with no warning")
        for (name, png) in [("b1-warning-before-payday", before), ("b1-warning-after-payday", after)] {
            let attachment = try XCTAttachment(data: XCTUnwrap(png), uniformTypeIdentifier: "public.png")
            attachment.name = name
            attachment.lifetime = .keepAlways
            add(attachment)
        }
    }
}
