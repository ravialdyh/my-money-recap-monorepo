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
            for: HomeSnapshot.self,
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
            for: HomeSnapshot.self,
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
}
