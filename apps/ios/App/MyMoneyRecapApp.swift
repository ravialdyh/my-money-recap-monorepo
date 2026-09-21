import SwiftData
import SwiftUI

@main
struct MyMoneyRecapApp: App {
    @State private var session = AuthSession()

    var body: some Scene {
        WindowGroup {
            RootView().environment(session)
        }
        .modelContainer(HomeSnapshot.container)
    }
}

/// A saved session skips Sign in (PRD §9); the Face ID gate is a later slice.
struct RootView: View {
    @Environment(AuthSession.self) private var session
    @Environment(\.modelContext) private var context

    var body: some View {
        Group {
            if session.isSignedIn {
                HomeView(session: session, context: context)
            } else {
                SignInView(session: session)
            }
        }
        .onChange(of: session.isSignedIn) { _, signedIn in
            if !signedIn {
                try? context.delete(model: HomeSnapshot.self)
                try? context.save()
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
