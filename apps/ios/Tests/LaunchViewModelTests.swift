import MoneyRecapContract
import XCTest
@testable import MyMoneyRecap

/// The launch after sign-in (spec #92): the zone first, then Home or the first unanswered of steps 1 and 2 with the
/// answers already saved, going back without losing any, and the no-connection state when nothing answers.
@MainActor
final class LaunchViewModelTests: XCTestCase {
    private func state(
        _ next: Int32,
        name: String = "",
        avatar: String = "",
        suggested: String = ""
    ) -> MoneyRecap_V1_OnboardingState {
        .with {
            $0.nextStep = next
            $0.profile.displayName = name
            $0.profile.avatarKey = avatar
            $0.suggestedName = suggested
        }
    }

    private func profile(_ name: String, _ avatar: String, _ currency: String = "") -> MoneyRecap_V1_Profile {
        .with {
            $0.displayName = name
            $0.avatarKey = avatar
            $0.currencyCode = currency
        }
    }

    func testAMemberHomeAlreadyGreetsOpensHomeAtOnceAndStillSendsTheZone() async {
        var zones = 0
        var asked = 0
        let model = LaunchViewModel(greetedBefore: true, setZone: { zones += 1 }, state: {
            asked += 1
            return nil
        })
        XCTAssertEqual(model.route, .home, "drawn from the snapshot before any call (I4), so no signal is needed")
        await model.start()
        XCTAssertEqual(zones, 1, "the zone goes on every launch (A3)")
        XCTAssertEqual(asked, 0)
    }

    func testAnUnansweredStepOneOpensWithTheProvidersName() async {
        let model = LaunchViewModel(greetedBefore: false, setZone: {}, state: { self.state(1, suggested: "Test") })
        XCTAssertEqual(model.route, .deciding)
        await model.start()
        XCTAssertEqual(model.route, .stepOne(name: "Test", avatar: ""))
        model.saved(profile("Test", "owl"))
        model.stepOneSaved()
        XCTAssertEqual(model.route, .stepTwo(saved: ""), "step 1 leads to the Currency, the region's chosen")
        model.stepTwoSaved()
        XCTAssertEqual(model.route, .home, "steps 3 and 4 are not built yet, so Home")
    }

    func testAResumeOpensTheFirstUnansweredStepWithTheSavedAnswers() async {
        let resumed = LaunchViewModel(
            greetedBefore: false,
            setZone: {},
            state: { self.state(1, name: "Maya", avatar: "otter", suggested: "Test") }
        )
        await resumed.start()
        XCTAssertEqual(resumed.route, .stepOne(name: "Maya", avatar: "otter"), "a saved name wins over the provider's")
        let currency = LaunchViewModel(greetedBefore: false, setZone: {}, state: { self.state(2) })
        await currency.start()
        XCTAssertEqual(currency.route, .stepTwo(saved: ""))
        let later = LaunchViewModel(greetedBefore: false, setZone: {}, state: { self.state(3) })
        await later.start()
        XCTAssertEqual(later.route, .home, "step 3 is not built yet, so Home")
    }

    /// Step 2's back chevron re-selects what step 1 saved; saving step 1 again returns to step 2 with its own saved
    /// answer chosen, so going back never costs a later answer (spec #92).
    func testGoingBackReselectsTheSavedAnswersAndNextReturnsToStepTwo() async {
        let model = LaunchViewModel(
            greetedBefore: false,
            setZone: {},
            state: { self.state(2, name: "Maya", avatar: "otter") }
        )
        await model.start()
        XCTAssertEqual(model.route, .stepTwo(saved: ""))
        model.back()
        XCTAssertEqual(model.route, .stepOne(name: "Maya", avatar: "otter"), "resumed answers, from the server")
        model.saved(profile("Mia", "fox", "IDR"))
        model.stepOneSaved()
        XCTAssertEqual(model.route, .stepTwo(saved: "IDR"), "the saved Currency, not the region's")
        model.back()
        XCTAssertEqual(model.route, .stepOne(name: "Mia", avatar: "fox"), "the answers the last save returned")
    }

    /// With no answer the Member cannot be told apart from one mid-setup, and nothing can be saved: the named state,
    /// whose Try again runs the launch again, zone included (spec #92).
    func testNoAnswerAtLaunchIsTheNoConnectionStateAndTryAgainResumes() async {
        var zones = 0
        var next: MoneyRecap_V1_OnboardingState?
        var retryingWhileAsked = false
        var model: LaunchViewModel?
        model = LaunchViewModel(greetedBefore: false, setZone: { zones += 1 }, state: {
            retryingWhileAsked = model?.isRetrying ?? false
            return next
        })
        await model?.start()
        XCTAssertEqual(model?.route, .noConnection, "never Home: this Member may be part-way through setup")
        next = state(2)
        await model?.retry()
        XCTAssertTrue(retryingWhileAsked, "Try again shows it is working")
        XCTAssertEqual(model?.isRetrying, false)
        XCTAssertEqual(zones, 2)
        XCTAssertEqual(model?.route, .stepTwo(saved: ""))
    }

    /// A Home that refuses (failed_precondition, spec #92) reopens the first unanswered step, even for a Member the
    /// stored Home greeted; with no answer, the no-connection state rather than a Home that would refuse again.
    func testARefusedHomeReopensTheUnansweredStep() async {
        var next: MoneyRecap_V1_OnboardingState? = state(2)
        let model = LaunchViewModel(greetedBefore: true, setZone: {}, state: { next })
        await model.start()
        XCTAssertEqual(model.route, .home)
        await model.homeRefused()
        XCTAssertEqual(model.route, .stepTwo(saved: ""))
        next = nil
        await model.homeRefused()
        XCTAssertEqual(model.route, .noConnection)
    }
}
