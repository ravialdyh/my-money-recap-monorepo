import MoneyRecapContract
import SwiftData
import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// Home as drawn: with no stored answer yet it is the §6.19 first-load skeleton, the two cards at their filled
/// heights, not a blank ground and not the zero Home it becomes.
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
}
