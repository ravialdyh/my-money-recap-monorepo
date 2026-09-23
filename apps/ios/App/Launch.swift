import MoneyRecapContract
import SwiftUI

/// The launch after sign-in (spec #92): the zone (A3), whose Profile opens the restore sheet for a recent deletion
/// (D60); then Home for a Member the stored Home greets (I4), else the first unanswered step; with no answer, no
/// connection.
@Observable
@MainActor
final class LaunchViewModel {
    enum Route: Equatable {
        case deciding
        case restore(asked: Date)
        case noConnection
        case stepOne(name: String, avatar: String)
        case stepTwo(saved: String)
        case stepThree(currency: String, digits: Int, saved: MoneyRecap_V1_SalarySetting?)
        case stepFour(salary: MoneyRecap_V1_Money?, digits: Int, saved: MoneyRecap_V1_SavingTargetSetting?)
        case home
    }

    private(set) var route: Route
    private(set) var isRetrying = false
    /// The answers saved so far, the server's copy: from the onboarding state, then from each save's answer.
    private var profile = MoneyRecap_V1_Profile()
    /// Step 3's saved Salary, relabelled whenever step 2 is saved again (spec #92).
    private var salary: MoneyRecap_V1_SalarySetting?
    private var target: MoneyRecap_V1_SavingTargetSetting?
    /// The step a Home prompt opened, which Try again opens again; nil during setup.
    private var openedFromHome: Int?
    /// The launch call's Profile, nil when the BFF could not answer.
    private let setZone: () async -> MoneyRecap_V1_Profile?
    private let state: () async -> MoneyRecap_V1_OnboardingState?
    private let now: () -> Date

    init(
        greetedBefore: Bool,
        setZone: @escaping () async -> MoneyRecap_V1_Profile?,
        state: @escaping () async -> MoneyRecap_V1_OnboardingState?,
        now: @escaping () -> Date = Date.init
    ) {
        route = greetedBefore ? .home : .deciding
        self.setZone = setZone
        self.state = state
        self.now = now
    }

    func start() async {
        if await restoring() {
            return
        }
        guard route == .deciding else { return }
        route = await decided()
    }

    /// Try again on the no-connection state: the whole launch once more, zone included, or the step Home opened.
    func retry() async {
        isRetrying = true
        defer { isRetrying = false }
        if let step = openedFromHome {
            await open(step: step)
            return
        }
        if await restoring() {
            return
        }
        route = await decided()
    }

    /// "Keep my account" cleared the request: the setup check, then Home or the first unanswered step.
    func kept() async {
        route = .deciding
        route = await decided()
    }

    /// A Home prompt: step 3 or 4 with the saved answers, which only the onboarding state carries (spec #92).
    func open(step: Int) async {
        openedFromHome = step
        route = .deciding
        guard let state = await state() else {
            route = .noConnection
            return
        }
        remember(state)
        route = step == 3 ? stepThree : stepFour
    }

    /// A save the BFF answered, whose profile is what the steps re-select when the Member goes back. A Currency save
    /// during setup also answers the Salary it relabelled; any other save leaves the Salary as it is.
    func saved(_ answer: MoneyRecap_V1_BffServiceSaveProfileResponse) {
        profile = answer.profile
        if answer.hasSalary {
            salary = answer.salary
        }
    }

    func stepOneSaved() {
        route = .stepTwo(saved: profile.currencyCode)
    }

    func stepTwoSaved() {
        route = stepThree
    }

    /// Step 3's Salary as ledger answered it, the one step 4 takes a share of.
    func salarySaved(_ settings: MoneyRecap_V1_LedgerSettings) {
        salary = settings.hasSalary ? settings.salary : salary
    }

    /// Step 3 saved or skipped: step 4, as during setup, blocked when no Salary was ever saved (D44).
    func stepThreeDone() {
        route = stepFour
    }

    /// Step 4's prompt back to step 3 (D44).
    func setSalary() {
        route = stepThree
    }

    /// Step 4 saved or skipped: setup is over.
    func stepFourDone() {
        openedFromHome = nil
        route = .home
    }

    /// The back chevron: the step before, with its saved answers. Saving it again leaves later answers and the saved
    /// progress as they are (members keeps the highest step), and Next returns to the step after it.
    func back() {
        if openedFromHome != nil {
            openedFromHome = nil
            route = .home
            return
        }
        switch route {
        case .stepFour:
            route = stepThree
        case .stepThree:
            route = .stepTwo(saved: profile.currencyCode)
        default:
            route = .stepOne(name: profile.displayName, avatar: profile.avatarKey)
        }
    }

    /// Home answered failed_precondition: step 1 or 2 is unanswered, so that step opens (spec #92).
    func homeRefused() async {
        route = .deciding
        route = await decided()
    }

    /// The launch call, and the restore sheet when its Profile carries a deletion under 24 hours old.
    private func restoring() async -> Bool {
        let zone = await setZone()
        let asked = zone.flatMap { $0.hasDeletionRequestedAt ? $0.deletionRequestedAt.date : nil }
        guard let asked, RestoreAccountViewModel.shows(asked: asked, now: now()) else { return false }
        route = .restore(asked: asked)
        return true
    }

    private func decided() async -> Route {
        guard let state = await state() else { return .noConnection }
        remember(state)
        switch state.nextStep {
        case 1:
            let saved = profile.displayName
            return .stepOne(name: saved.isEmpty ? state.suggestedName : saved, avatar: profile.avatarKey)
        case 2:
            return .stepTwo(saved: profile.currencyCode)
        case 3:
            return stepThree
        case 4:
            return stepFour
        default:
            return .home
        }
    }

    private func remember(_ state: MoneyRecap_V1_OnboardingState) {
        profile = state.profile
        salary = state.hasSalary ? state.salary : nil
        target = state.hasSavingTarget ? state.savingTarget : nil
    }

    private var stepThree: Route {
        .stepThree(currency: profile.currencyCode, digits: Int(profile.currencyDigits), saved: salary)
    }

    private var stepFour: Route {
        .stepFour(salary: salary?.amount, digits: Int(profile.currencyDigits), saved: target)
    }
}
