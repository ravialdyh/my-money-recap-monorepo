import SwiftData
import SwiftUI

@main
struct MyMoneyRecapApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) private var delegate
    @State private var session = AuthSession()

    var body: some Scene {
        WindowGroup {
            RootView().environment(session)
        }
        .modelContainer(HomeSnapshot.container)
    }
}

struct RootView: View {
    @Environment(AuthSession.self) private var session
    @Environment(\.modelContext) private var context
    @Environment(\.scenePhase) private var phase
    @State private var leftAt: ContinuousClock.Instant?

    var body: some View {
        Group {
            switch session.lock {
            case .unlocked where session.isSignedIn: LaunchView(session: session, context: context)
            case .locked: UnlockView(session: session)
            default: SignInView(session: session)
            }
        }
        .onChange(of: phase) { _, now in
            if now == .background {
                leftAt = .now
            } else if now == .active, let leftAt {
                session.returned(after: TimeInterval((ContinuousClock.now - leftAt).components.seconds))
                self.leftAt = nil
            }
        }
        .onChange(of: session.isSignedIn) { _, signedIn in
            if !signedIn {
                HomeSnapshot.erase(from: context)
                PushRegistration.signedOut()
            }
        }
        .task {
            guard let link = MagicLink.launchedLink else { return }
            await session.openMagicLink(link)
        }
        .onOpenURL { url in
            Task { await session.openMagicLink(url) }
        }
    }
}
