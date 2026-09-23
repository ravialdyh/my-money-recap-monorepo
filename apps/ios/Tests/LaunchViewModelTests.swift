import MoneyRecapContract
import XCTest
@testable import MyMoneyRecap

/// The launch after sign-in (spec #92): the zone first, then Home or the first unanswered of steps 1 to 4 with the
/// answers already saved, going back without losing any, Home's prompts reopening steps 3 and 4, and the no-connection
/// state when nothing answers.
@MainActor
final class LaunchViewModelTests: XCTestCase {
    private func state(
        _ next: Int32,
        name: String = "",
        avatar: String = "",
        suggested: String = "",
        salary: MoneyRecap_V1_SalarySetting? = nil,
        target: MoneyRecap_V1_SavingTargetSetting? = nil
    ) -> MoneyRecap_V1_OnboardingState {
        .with {
            $0.nextStep = next
            $0.profile = profile(name, avatar, next > 2 || next == 0 ? "IDR" : "")
            $0.suggestedName = suggested
            if let salary {
                $0.salary = salary
            }
            if let target {
                $0.savingTarget = target
            }
        }
    }

    private func profile(_ name: String, _ avatar: String, _ currency: String = "") -> MoneyRecap_V1_Profile {
        .with {
            $0.displayName = name
            $0.avatarKey = avatar
            $0.currencyCode = currency
            $0.currencyDigits = currency == "EUR" ? 2 : 0
        }
    }

    private func answer(
        _ profile: MoneyRecap_V1_Profile,
        salary: MoneyRecap_V1_SalarySetting? = nil
    ) -> MoneyRecap_V1_BffServiceSaveProfileResponse {
        .with {
            $0.profile = profile
            if let salary {
                $0.salary = salary
            }
        }
    }

    private func salary(_ amount: String, _ currency: String) -> MoneyRecap_V1_SalarySetting {
        .with {
            $0.amount.amount = amount
            $0.amount.currencyCode = currency
            $0.payday = "last"
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
        model.saved(answer(profile("Test", "owl")))
        model.stepOneSaved()
        XCTAssertEqual(model.route, .stepTwo(saved: ""), "step 1 leads to the Currency, the region's chosen")
        model.saved(answer(profile("Test", "owl", "EUR")))
        model.stepTwoSaved()
        XCTAssertEqual(model.route, .stepThree(currency: "EUR", digits: 2, saved: nil), "typed in the saved Currency")
        model.stepThreeDone()
        XCTAssertEqual(model.route, .stepFour(salary: nil, digits: 2, saved: nil), "skipped, so blocked (D44)")
        model.setSalary()
        XCTAssertEqual(model.route, .stepThree(currency: "EUR", digits: 2, saved: nil), "the prompt back to step 3")
        let typed = salary("2400.00", "EUR")
        model.salarySaved(.with { $0.salary = typed })
        model.stepThreeDone()
        XCTAssertEqual(model.route, .stepFour(salary: typed.amount, digits: 2, saved: nil), "a share of what was saved")
        model.back()
        XCTAssertEqual(model.route, .stepThree(currency: "EUR", digits: 2, saved: typed), "step 3 with its answer")
        model.stepThreeDone()
        model.stepFourDone()
        XCTAssertEqual(model.route, .home)
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
        let typed = salary("15000000", "IDR")
        let salaried = LaunchViewModel(greetedBefore: false, setZone: {}, state: { self.state(3, salary: typed) })
        await salaried.start()
        XCTAssertEqual(salaried.route, .stepThree(currency: "IDR", digits: 0, saved: typed), "a crash after ledger")
        let kept = MoneyRecap_V1_SavingTargetSetting.with { $0.percent = 35 }
        let later = LaunchViewModel(
            greetedBefore: false,
            setZone: {},
            state: { self.state(4, salary: typed, target: kept) }
        )
        await later.start()
        XCTAssertEqual(later.route, .stepFour(salary: typed.amount, digits: 0, saved: kept), "a crash after ledger")
        let done = LaunchViewModel(greetedBefore: false, setZone: {}, state: { self.state(0) })
        await done.start()
        XCTAssertEqual(done.route, .home, "setup over")
    }

    /// Home's prompts open step 3 or 4 after setup with the saved answers; step 3 leads on to step 4 as in setup, back
    /// returns to Home, never to step 2, and with no answer Try again reopens the same step (owner 2026-09-23).
    func testHomesPromptsOpenStepsThreeAndFourAndBackReturnsHome() async {
        var next: MoneyRecap_V1_OnboardingState? = state(0)
        let model = LaunchViewModel(greetedBefore: true, setZone: {}, state: { next })
        await model.start()
        await model.open(step: 3)
        XCTAssertEqual(model.route, .stepThree(currency: "IDR", digits: 0, saved: nil))
        let typed = salary("15000000", "IDR")
        model.salarySaved(.with { $0.salary = typed })
        model.stepThreeDone()
        XCTAssertEqual(model.route, .stepFour(salary: typed.amount, digits: 0, saved: nil))
        model.back()
        XCTAssertEqual(model.route, .home, "back from a step Home opened is Home")
        next = nil
        await model.open(step: 4)
        XCTAssertEqual(model.route, .noConnection)
        next = state(0, salary: typed)
        await model.retry()
        XCTAssertEqual(model.route, .stepFour(salary: typed.amount, digits: 0, saved: nil), "the step Home opened")
        model.stepFourDone()
        XCTAssertEqual(model.route, .home)
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
        model.saved(answer(profile("Mia", "fox", "IDR")))
        model.stepOneSaved()
        XCTAssertEqual(model.route, .stepTwo(saved: "IDR"), "the saved Currency, not the region's")
        model.back()
        XCTAssertEqual(model.route, .stepOne(name: "Mia", avatar: "fox"), "the answers the last save returned")
    }

    /// Step 3's back chevron re-selects the saved Currency; saving another relabels the typed Salary, which step 3
    /// fills in again in the new Currency with its decimals, and a later save without a Salary keeps it (spec #92).
    func testGoingBackFromStepThreeKeepsTheSalaryRelabelled() async {
        let typed = salary("15000000", "IDR")
        let model = LaunchViewModel(greetedBefore: false, setZone: {}, state: { self.state(3, salary: typed) })
        await model.start()
        model.back()
        XCTAssertEqual(model.route, .stepTwo(saved: "IDR"))
        let relabelled = salary("15000000.00", "EUR")
        model.saved(answer(profile("Maya", "otter", "EUR"), salary: relabelled))
        model.stepTwoSaved()
        XCTAssertEqual(model.route, .stepThree(currency: "EUR", digits: 2, saved: relabelled))
        model.back()
        model.back()
        model.saved(answer(profile("Mia", "otter", "EUR")))
        model.stepOneSaved()
        model.stepTwoSaved()
        XCTAssertEqual(model.route, .stepThree(currency: "EUR", digits: 2, saved: relabelled), "step 1 moved nothing")
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
