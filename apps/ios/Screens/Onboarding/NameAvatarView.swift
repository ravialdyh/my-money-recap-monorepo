import SwiftUI

struct NameAvatarView: View {
    @State private var model: NameAvatarViewModel
    @State private var typed: String
    private let onSaved: () -> Void
    @FocusState private var focused: Bool

    init(model: NameAvatarViewModel, onSaved: @escaping () -> Void) {
        _model = State(initialValue: model)
        _typed = State(initialValue: model.name)
        self.onSaved = onSaved
    }

    var body: some View {
        Screen {
            ScrollView {
                VStack(alignment: .leading, spacing: 0) {
                    Chip(text: "Step 1 of 4", asLabel: true).frame(minHeight: Tokens.Size.touchTargetMin)
                    Headline(text: "What should we call you?").padding(.top, 18)
                    VStack(spacing: Tokens.Space._5) {
                        Card {
                            if model.offline {
                                InlineError(text: NoConnectionView.copy)
                            }
                            MMRTextField(
                                label: "First name",
                                text: $typed,
                                focused: focused,
                                content: .givenName,
                                capitalization: .words,
                                submit: .done,
                                onSubmit: { focused = false }
                            )
                            .focused($focused)
                            .onChange(of: typed) { _, now in
                                model.type(now)
                                if typed != model.name {
                                    typed = model.name
                                }
                            }
                        }
                        Card {
                            CardTitle(text: "Pick your avatar")
                            AvatarGrid(selected: model.avatar) { model.choose($0) }
                        }
                    }
                    .padding(.top, Tokens.Space._7)
                }
            }
            .scrollIndicators(.hidden)
            MMRButton(title: "Next", loading: model.isSaving) {
                focused = false
                Task {
                    if await model.next() {
                        onSaved()
                    }
                }
            }
            .disabled(!model.canContinue)
        }
    }
}
