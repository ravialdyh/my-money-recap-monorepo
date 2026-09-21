import SwiftUI

/// PRD §6.1 / SCREEN-STATES A1. Apple and Google are drawn only until slice 3.
struct SignInView: View {
    private let session: AuthSession
    @State private var model: SignInViewModel
    @State private var path: [EmailStep] = []

    init(session: AuthSession) {
        self.session = session
        _model = State(initialValue: SignInViewModel(
            signIn: { try await session.signIn() },
            initialError: session.signInMessage
        ))
    }

    var body: some View {
        NavigationStack(path: $path) {
            signIn
                .toolbar(.hidden, for: .navigationBar)
                .navigationDestination(for: EmailStep.self) { step in
                    switch step {
                    case .enterEmail:
                        EnterEmailView(session: session, path: $path)
                    case let .checkInbox(email):
                        CheckInboxView(session: session, email: email, path: $path)
                    }
                }
        }
        .onChange(of: session.magicLinkFailure) { _, failure in
            if failure != nil {
                path = [.enterEmail]
            }
        }
    }

    private var signIn: some View {
        Screen {
            Headline(text: "Your month in money.")
            SubCopy(text: "Sign in to start your recap.").padding(.top, 10)
            Illustration(name: "ill-cover", width: 260)
            Card(underIllustration: true) {
                if let error = model.errorMessage {
                    InlineError(text: error)
                }
                appleButton
                MMRButton(title: "Continue with Google", variant: .secondary) {}
                MMRButton(title: "Continue with email", loading: model.isLoading, action: continueWithEmail)
            }
            .disabled(model.isLoading)
            Text("By continuing you agree to the Terms and Privacy Policy.")
                .textStyle(Tokens.TypeRamp.bodyS)
                .foregroundStyle(Tokens.Text.onGroundMuted)
                .frame(maxWidth: .infinity)
                .multilineTextAlignment(.center)
                .padding(.top, Tokens.Space._5)
        }
        .animation(.easeOut(duration: Tokens.Motion.Duration.reduced), value: model.errorMessage)
    }

    /// Enter email, or the old web sheet when the debug launch argument keeps the authorization-code flow
    /// under test (spec #61 seam 6).
    private func continueWithEmail() {
        if AuthConfig.usesAuthorizationCodeFlow {
            Task { await model.continueWithEmail() }
        } else {
            path.append(.enterEmail)
        }
    }

    /// Apple's own button style (black pill, system type, Apple glyph), listed first. Wired up in slice 3.
    private var appleButton: some View {
        Button {} label: {
            Label("Sign in with Apple", systemImage: "apple.logo")
                .font(.system(size: 17, weight: .semibold))
                .frame(maxWidth: .infinity, minHeight: 52)
                .foregroundStyle(.white)
                .background(.black, in: Capsule())
        }
        .buttonStyle(PressableStyle())
    }
}

/// The email path's two screens, pushed on Sign in's own stack (PRD §6.2, §6.3).
enum EmailStep: Hashable {
    case enterEmail
    case checkInbox(String)
}

/// The 44pt back chevron and the "Email" chip both screens carry, with no header title
/// (`docs/design/design_handoff_my_money_recap/README.md`, A2 and A3).
struct EmailHeader: View {
    let back: () -> Void

    var body: some View {
        HStack(spacing: Tokens.Space._1) {
            Button(action: back) {
                Image(systemName: "chevron.left")
                    .font(.system(size: 20, weight: .bold))
                    .frame(width: Tokens.Size.touchTargetMin, height: Tokens.Size.touchTargetMin)
                    .contentShape(Rectangle())
            }
            .buttonStyle(PressableStyle())
            .accessibilityLabel("Back")
            Chip(text: "Email")
        }
        .padding(.leading, -Tokens.Space._5)
    }
}
