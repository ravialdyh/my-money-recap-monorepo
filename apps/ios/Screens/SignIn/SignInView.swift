import SwiftUI

/// PRD §6.1 / SCREEN-STATES A1. Apple and Google hand the web sheet to their provider through `kc_idp_hint` (#71).
struct SignInView: View {
    private let session: AuthSession
    @State private var model: SignInViewModel
    @State private var path: [EmailStep] = []

    init(session: AuthSession) {
        self.session = session
        _model = State(initialValue: SignInViewModel(
            signIn: { try await session.signIn(identityProvider: $0.identityProvider) },
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
                    .disabled(model.dims(.apple))
                MMRButton(title: "Continue with Google", variant: .secondary, loading: model.tapped == .google) {
                    tap(.google)
                }
                .disabled(model.dims(.google))
                MMRButton(title: "Continue with email", loading: model.tapped == .email, action: continueWithEmail)
                    .disabled(model.dims(.email))
            }
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
            tap(.email)
        } else {
            path.append(.enterEmail)
        }
    }

    private func tap(_ button: SignInButton) {
        Task { await model.tap(button) }
    }

    /// Apple's own button style (black pill, system type, Apple glyph), listed first as App Store guideline 4.8
    /// expects; the glyph becomes the spinner while its sheet is open.
    private var appleButton: some View {
        Button { tap(.apple) } label: {
            Label {
                Text("Sign in with Apple")
            } icon: {
                if model.tapped == .apple {
                    ProgressView().tint(.white)
                } else {
                    Image(systemName: "apple.logo")
                }
            }
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
