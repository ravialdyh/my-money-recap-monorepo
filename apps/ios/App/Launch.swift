import MoneyRecapContract
import SwiftUI

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
    private var profile = MoneyRecap_V1_Profile()
    private var salary: MoneyRecap_V1_SalarySetting?
    private var target: MoneyRecap_V1_SavingTargetSetting?
    private var openedFromHome: Int?
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

    func kept() async {
        route = .deciding
        route = await decided()
    }

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

    func salarySaved(_ settings: MoneyRecap_V1_LedgerSettings) {
        salary = settings.hasSalary ? settings.salary : salary
    }

    func stepThreeDone() {
        route = stepFour
    }

    func setSalary() {
        route = stepThree
    }

    func stepFourDone() {
        openedFromHome = nil
        route = .home
    }

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

    func homeRefused() async {
        route = .deciding
        route = await decided()
    }

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
