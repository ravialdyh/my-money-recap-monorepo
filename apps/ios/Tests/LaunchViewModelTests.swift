import MoneyRecapContract
import XCTest
@testable import MyMoneyRecap

/// The launch after sign-in (spec #92): the zone first, then Home or the first unanswered of steps 1 and 2.
@MainActor
final class LaunchViewModelTests: XCTestCase {
    private func state(_ next: Int32, name: String = "", suggested: String = "") -> MoneyRecap_V1_OnboardingState {
        .with {
            $0.nextStep = next
            $0.profile.displayName = name
            $0.suggestedName = suggested
        }
    }

    func testAMemberHomeAlreadyGreetsOpensHomeAtOnceAndStillSendsTheZone() async {
        var zones = 0
        var asked = 0
        let model = LaunchViewModel(greetedBefore: true, setZone: { zones += 1 }, state: {
            asked += 1
            return nil
        })
        XCTAssertEqual(model.route, .home, "drawn from the snapshot before any call (I4)")
        await model.start()
        XCTAssertEqual(zones, 1, "the zone goes on every launch (A3)")
        XCTAssertEqual(asked, 0)
    }

    func testAnUnansweredStepOneOpensWithTheProvidersName() async {
        let model = LaunchViewModel(greetedBefore: false, setZone: {}, state: { self.state(1, suggested: "Test") })
        XCTAssertEqual(model.route, .deciding)
        await model.start()
        XCTAssertEqual(model.route, .stepOne(name: "Test", avatar: ""))
        model.stepOneSaved()
        XCTAssertEqual(model.route, .stepTwo, "step 1 leads to the Currency")
        model.stepTwoSaved()
        XCTAssertEqual(model.route, .home, "steps 3 and 4 are not built yet, so Home")
    }

    func testASavedNameWinsOverTheProvidersAndAMissingCurrencyOpensStepTwo() async {
        let resumed = LaunchViewModel(
            greetedBefore: false,
            setZone: {},
            state: { self.state(1, name: "Maya", suggested: "Test") }
        )
        await resumed.start()
        XCTAssertEqual(resumed.route, .stepOne(name: "Maya", avatar: ""))
        let currency = LaunchViewModel(greetedBefore: false, setZone: {}, state: { self.state(2) })
        await currency.start()
        XCTAssertEqual(currency.route, .stepTwo)
        let later = LaunchViewModel(greetedBefore: false, setZone: {}, state: { self.state(3) })
        await later.start()
        XCTAssertEqual(later.route, .home, "step 3 is not built yet, so Home")
        let offline = LaunchViewModel(greetedBefore: false, setZone: {}, state: { nil })
        await offline.start()
        XCTAssertEqual(offline.route, .home, "no answer at all: Home and its skeleton")
    }

    /// A Home that refuses (failed_precondition, spec #92) reopens the first unanswered step, even for a Member the
    /// stored Home greeted; with no answer the screen waits instead of returning to a Home that would refuse again.
    func testARefusedHomeReopensTheUnansweredStep() async {
        var next: MoneyRecap_V1_OnboardingState? = state(2)
        let model = LaunchViewModel(greetedBefore: true, setZone: {}, state: { next })
        await model.start()
        XCTAssertEqual(model.route, .home)
        await model.homeRefused()
        XCTAssertEqual(model.route, .stepTwo)
        next = nil
        await model.homeRefused()
        XCTAssertEqual(model.route, .deciding)
    }
}
