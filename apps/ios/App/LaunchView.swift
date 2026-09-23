import Connect
import MoneyRecapContract
import SwiftData
import SwiftUI

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
            setZone: {
                try? await BffClient
                    .retried(session) { await client.setTimezone(request: request, headers: zone.headers) }
                    .result.get().profile
            },
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
            case let .restore(asked):
                RestoreAccountView(
                    model: RestoreAccountViewModel(asked: asked, keep: cancelDeletion),
                    onKept: { Task { await model.kept() } },
                    onContinue: { session.signOut() }
                )
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

    /// "Keep my account", retried once after the token refreshes (IOS I3).
    private func cancelDeletion(_ headers: Headers) async -> SaveResult {
        await SaveResult(BffClient.retried(session) { await client.cancelDeletion(request: .init(), headers: headers) }
            .result)
    }
}
