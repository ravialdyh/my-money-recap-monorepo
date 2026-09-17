import SwiftUI

/// PRD §6.1 / SCREEN-STATES A1. Apple and Google are drawn only until slice 3.
struct SignInView: View {
    @State private var model: SignInViewModel

    init(session: AuthSession) {
        _model = State(initialValue: SignInViewModel(
            signIn: { try await session.signIn() },
            initialError: session.signInMessage
        ))
    }

    var body: some View {
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
                MMRButton(title: "Continue with email", loading: model.isLoading) {
                    Task { await model.continueWithEmail() }
                }
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
