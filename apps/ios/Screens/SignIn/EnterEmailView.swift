import SwiftUI

/// PRD §6.2 / SCREEN-STATES A2. One send, then Check your inbox; a refused link comes back here with its
/// own copy and the address still in the field.
struct EnterEmailView: View {
    private let session: AuthSession
    @Binding var path: [EmailStep]
    @State private var model: EnterEmailViewModel
    @FocusState private var focused: Bool

    init(session: AuthSession, path: Binding<[EmailStep]>) {
        self.session = session
        _path = path
        _model = State(initialValue: EnterEmailViewModel(
            send: MagicLink.sender(session),
            initialFailure: session.magicLinkFailure
        ))
    }

    var body: some View {
        Screen {
            EmailHeader { path.removeAll() }
            Headline(text: "Where should we send your link?", size: .m).padding(.top, 18)
            Card {
                if let error = model.errorMessage {
                    InlineError(text: error)
                }
                MMRTextField(
                    label: "Email address",
                    text: $model.email,
                    placeholder: "you@example.com",
                    helper: model.invalidHelper,
                    invalid: model.invalidHelper != nil,
                    focused: focused,
                    keyboard: .emailAddress,
                    content: .emailAddress,
                    onSubmit: send
                )
                .focused($focused)
            }
            .padding(.top, Tokens.Space._7)
            Spacer(minLength: Tokens.Space._7)
            MMRButton(title: "Send me a link", loading: model.isSending, action: send)
                .disabled(!model.canSend)
        }
        .toolbar(.hidden, for: .navigationBar)
        .animation(.easeOut(duration: Tokens.Motion.Duration.reduced), value: model.errorMessage)
        .onChange(of: focused) { _, isFocused in
            if !isFocused {
                model.blur()
            }
        }
        .onChange(of: session.magicLinkFailure) { _, failure in
            model.show(failure)
        }
    }

    private func send() {
        focused = false
        Task {
            if let address = await model.sendLink() {
                path.append(.checkInbox(address))
            }
        }
    }
}
