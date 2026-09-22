import MoneyRecapContract
import XCTest
@testable import MyMoneyRecap

/// The launch after sign-in (spec #92): the zone first, then Home or step 1.
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
        XCTAssertEqual(model.route, .home)
    }

    func testASavedNameWinsOverTheProvidersAndLaterStepsOpenHome() async {
        let resumed = LaunchViewModel(
            greetedBefore: false,
            setZone: {},
            state: { self.state(1, name: "Maya", suggested: "Test") }
        )
        await resumed.start()
        XCTAssertEqual(resumed.route, .stepOne(name: "Maya", avatar: ""))
        let later = LaunchViewModel(greetedBefore: false, setZone: {}, state: { self.state(2) })
        await later.start()
        XCTAssertEqual(later.route, .home, "step 2 is not built yet, so Home")
        let offline = LaunchViewModel(greetedBefore: false, setZone: {}, state: { nil })
        await offline.start()
        XCTAssertEqual(offline.route, .home, "no answer at all: Home and its skeleton")
    }
}
