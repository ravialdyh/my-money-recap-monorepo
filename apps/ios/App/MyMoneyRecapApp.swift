import SwiftUI

@main
struct MyMoneyRecapApp: App {
    @State private var session = AuthSession()

    var body: some Scene {
        WindowGroup {
            RootView().environment(session)
        }
    }
}

/// A saved session skips Sign in (PRD §9); the Face ID gate is a later slice.
struct RootView: View {
    @Environment(AuthSession.self) private var session

    var body: some View {
        if session.isSignedIn {
            HomePlaceholderView()
        } else {
            SignInView(session: session)
        }
    }
}
