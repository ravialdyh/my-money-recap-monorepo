import Connect
import MoneyRecapContract
import SwiftData
import SwiftUI

/// The launch after sign-in (spec #92): the zone (A3), then Home at once for a Member the stored Home greets (I4),
/// else the first unanswered step with its answers; with no answer, no connection. Home's prompts reopen steps 3 and 4.
@Observable
@MainActor
final class LaunchViewModel {
    enum Route: Equatable {
        case deciding
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
    private let setZone: () async -> Void
    private let state: () async -> MoneyRecap_V1_OnboardingState?

    init(
        greetedBefore: Bool,
        setZone: @escaping () async -> Void,
        state: @escaping () async -> MoneyRecap_V1_OnboardingState?
    ) {
        route = greetedBefore ? .home : .deciding
        self.setZone = setZone
        self.state = state
    }

    func start() async {
        await setZone()
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
        await setZone()
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

/// Routes between the steps and Home on one client, so the zone, the state and the saves share its interceptor.
struct LaunchView: View {
    private let session: AuthSession
    private let context: ModelContext
    private let client: MoneyRecap_V1_BffServiceClient
    @State private var model: LaunchViewModel

    init(session: AuthSession, context: ModelContext) {
        self.session = session
        self.context = context
        let client = BffClient.make(session: session)
        self.client = client
        let zone = SaveAttempt(TimeZone.current.identifier)
        let request = MoneyRecap_V1_BffServiceSetTimezoneRequest.with { $0.timezone = zone.values }
        let greeted = HomeSnapshot.load(from: context)
            .flatMap { try? MoneyRecap_V1_HomeView(serializedBytes: $0.bytes) }
            .map { !$0.displayName.isEmpty } ?? false
        _model = State(initialValue: LaunchViewModel(
            greetedBefore: greeted,
            setZone: { _ = await BffClient.retried(session) { await client.setTimezone(
                request: request,
                headers: zone.headers
            ) } },
            state: {
                try? await BffClient.retried(session) { await client.getOnboardingState(request: .init()) }.result.get()
                    .state
            }
        ))
    }

    var body: some View {
        Group {
            switch model.route {
            case .deciding:
                Screen {}
            case .noConnection:
                NoConnectionView(retrying: model.isRetrying) {
                    Task { await model.retry() }
                }
            case let .stepOne(name, avatar):
                NameAvatarView(
                    model: NameAvatarViewModel(name: name, avatar: avatar, save: saveProfile),
                    onSaved: model.stepOneSaved
                )
            case let .stepTwo(saved):
                CurrencyView(
                    model: CurrencyViewModel(saved: saved, save: saveProfile),
                    onBack: model.back,
                    onSaved: model.stepTwoSaved
                )
            case let .stepThree(currency, digits, saved):
                SalaryView(
                    model: SalaryViewModel(
                        currency: currency,
                        digits: digits,
                        saved: saved,
                        saveSalary: saveSalary,
                        saveProfile: saveProfile
                    ),
                    onBack: model.back,
                    onDone: model.stepThreeDone
                )
            case let .stepFour(salary, digits, saved):
                TargetView(
                    model: TargetViewModel(
                        salary: salary,
                        digits: digits,
                        saved: saved,
                        saveTarget: saveTarget,
                        saveProfile: saveProfile
                    ),
                    onBack: model.back,
                    onSetSalary: model.setSalary,
                    onDone: model.stepFourDone
                )
            case .home:
                HomeView(session: session, context: context, open: { step in
                    Task { await model.open(step: step) }
                }, onboardingNeeded: {
                    Task { await model.homeRefused() }
                })
            }
        }
        .task { await model.start() }
    }

    /// One onboarding save, retried once after the token refreshes (IOS I3). The profile the BFF answers is what the
    /// steps re-select when the Member goes back.
    private func saveProfile(
        _ request: MoneyRecap_V1_BffServiceSaveProfileRequest,
        _ headers: Headers
    ) async -> SaveResult {
        let saved = await BffClient.retried(session) { await client.saveProfile(request: request, headers: headers) }
        if case let .success(answer) = saved.result {
            model.saved(answer)
        }
        return SaveResult(saved.result)
    }

    /// Step 3's Salary, retried once after the token refreshes (IOS I3); step 4 takes a share of what ledger saved.
    private func saveSalary(
        _ request: MoneyRecap_V1_BffServiceSaveSalaryRequest,
        _ headers: Headers
    ) async -> SaveResult {
        let saved = await BffClient.retried(session) { await client.saveSalary(request: request, headers: headers) }
        if case let .success(answer) = saved.result {
            model.salarySaved(answer.settings.ledgerSettings)
        }
        return SaveResult(saved.result)
    }

    /// Step 4's Saving Target, retried once after the token refreshes (IOS I3).
    private func saveTarget(
        _ request: MoneyRecap_V1_BffServiceSaveTargetRequest,
        _ headers: Headers
    ) async -> SaveResult {
        await SaveResult(BffClient.retried(session) { await client.saveTarget(request: request, headers: headers) }
            .result)
    }
}
