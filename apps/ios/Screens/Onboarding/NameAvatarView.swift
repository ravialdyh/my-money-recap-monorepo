import SwiftUI

/// PRD §6.4 / SCREEN-STATES A4, step 1 of onboarding: no back chevron, the step label as the white pill every canvas
/// draws, the First name field and the Avatar grid, then Next.
struct NameAvatarView: View {
    @State private var model: NameAvatarViewModel
    /// What the field shows. The model keeps only what `DisplayName` allows, and writing that back here redraws
    /// the field even when a refused key leaves the model's name unchanged.
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
            Chip(text: "Step 1 of 4", asLabel: true).padding(.top, 8)
            Headline(text: "What should we call you?", size: .m).padding(.top, 18)
            ScrollView {
                VStack(spacing: Tokens.Space._5) {
                    Card {
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
