import SwiftUI

/// D2 Profile (PRD §6.16): setup's First name field and `Pick your avatar` grid on the white sheet, then Save, which
/// returns to Settings. A failed save keeps what was typed above the Settings save line.
struct ProfileView: View {
    @State private var model: ProfileViewModel
    /// What the field shows; the model keeps only what `DisplayName` allows, as on step 1.
    @State private var typed: String
    @FocusState private var focused: Bool
    @Environment(\.dismiss) private var dismiss
    private let onSaved: () -> Void

    init(model: ProfileViewModel, onSaved: @escaping () -> Void) {
        _model = State(initialValue: model)
        _typed = State(initialValue: model.name)
        self.onSaved = onSaved
    }

    var body: some View {
        Sheet(title: "Profile", back: { dismiss() }, content: {
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
            if model.failed {
                InlineError(text: SettingsViewModel.saveFailure)
            }
            MMRButton(title: "Save", loading: model.isSaving) {
                focused = false
                Task {
                    if await model.commit() {
                        onSaved()
                    }
                }
            }
            .disabled(!model.canSave)
        })
        .toolbar(.hidden, for: .navigationBar)
    }
}
