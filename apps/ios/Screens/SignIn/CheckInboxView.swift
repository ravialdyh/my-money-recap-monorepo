import SwiftUI

/// PRD §6.3 / SCREEN-STATES A3. Nothing here polls: the link itself brings the session in, through the
/// incoming-link handler (spec #61).
struct CheckInboxView: View {
    @Binding var path: [EmailStep]
    @State private var model: CheckInboxViewModel

    init(session: AuthSession, email: String, path: Binding<[EmailStep]>) {
        _path = path
        _model = State(initialValue: CheckInboxViewModel(email: email, send: MagicLink.sender(session)))
    }

    var body: some View {
        Screen {
            BackRow(label: "Email", back: useAnotherEmail)
            Image("ill-envelope")
                .resizable()
                .scaledToFit()
                .frame(width: 240)
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .accessibilityHidden(true)
            Headline(text: CheckInboxViewModel.title, size: .m)
            SubCopy(text: model.subCopy).padding(.top, Tokens.Space._4)
            VStack(alignment: .leading, spacing: Tokens.Space._1) {
                Button(model.resendTitle) {
                    Task { await model.resend() }
                }
                .disabled(!model.canResend)
                Button("Use a different email", action: useAnotherEmail)
                    .underline()
            }
            .textStyle(Tokens.TypeRamp.bodyM)
            .foregroundStyle(Tokens.Text.onGround)
            .frame(minHeight: Tokens.Size.touchTargetMin, alignment: .leading)
            .buttonStyle(PressableStyle())
            .padding(.top, Tokens.Space._8)
        }
        .toolbar(.hidden, for: .navigationBar)
        .task { await model.countDown() }
    }

    /// A3's own exit: back to A2, address still in the field.
    private func useAnotherEmail() {
        path = [.enterEmail]
    }
}
