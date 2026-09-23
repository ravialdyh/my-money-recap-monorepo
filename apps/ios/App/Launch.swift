import Connect
import MoneyRecapContract
import SwiftData
import SwiftUI

/// What the app opens on after sign-in (spec #92): the zone first on every launch (A3), then Home at once for a
/// Member the stored Home already greets (I4), else the first unanswered of steps 1 and 2. No answer at all: Home.
@Observable
@MainActor
final class LaunchViewModel {
    enum Route: Equatable {
        case deciding
        case stepOne(name: String, avatar: String)
        case stepTwo
        case home
    }

    private(set) var route: Route
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
        route = await decided(orElse: .home)
    }

    func stepOneSaved() {
        route = .stepTwo
    }

    func stepTwoSaved() {
        route = .home
    }

    /// Home answered failed_precondition: step 1 or 2 is unanswered, so that step opens (spec #92). With no answer
    /// the screen stays blank rather than returning to a Home that would refuse again.
    func homeRefused() async {
        route = .deciding
        route = await decided(orElse: .deciding)
    }

    private func decided(orElse fallback: Route) async -> Route {
        guard let state = await state() else { return fallback }
        switch state.nextStep {
        case 1:
            let saved = state.profile.displayName
            return .stepOne(name: saved.isEmpty ? state.suggestedName : saved, avatar: state.profile.avatarKey)
        case 2:
            return .stepTwo
        default:
            return .home
        }
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
            case let .stepOne(name, avatar):
                NameAvatarView(
                    model: NameAvatarViewModel(name: name, avatar: avatar, save: saveProfile),
                    onSaved: model.stepOneSaved
                )
            case .stepTwo:
                CurrencyView(model: CurrencyViewModel(save: saveProfile), onSaved: model.stepTwoSaved)
            case .home:
                HomeView(session: session, context: context) {
                    Task { await model.homeRefused() }
                }
            }
        }
        .task { await model.start() }
    }

    /// One onboarding save, retried once after the token refreshes (IOS I3); true once the BFF has it.
    private func saveProfile(_ request: MoneyRecap_V1_BffServiceSaveProfileRequest, _ headers: Headers) async -> Bool {
        let saved = await BffClient.retried(session) { await client.saveProfile(request: request, headers: headers) }
        return (try? saved.result.get()) != nil
    }
}
