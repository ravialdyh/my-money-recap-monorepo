import SwiftUI

/// D10 Delete account (PRD §6.16): the coral card saying what goes and that the Member has 24 hours, and the coral
/// button that pushes D11. Both scroll sideways.
struct DeleteAccountView: View {
    @State private var model: DeleteAccountViewModel
    @Environment(\.dismiss) private var dismiss

    init(model: DeleteAccountViewModel) {
        _model = State(initialValue: model)
    }

    var body: some View {
        Sheet(title: "Delete account", back: { dismiss() }, content: {
            Card(coral: true) {
                CardTitle(text: "This is permanent")
                Text(
                    "Deleting removes your account, every spend you've logged and every recap. " +
                        "Sign in again within 24 hours to keep it; after that we can't get any of it back."
                )
                .textStyle(Tokens.TypeRamp.bodyS)
                .fixedSize(horizontal: false, vertical: true)
            }
            MMRButton(title: "Delete my account", variant: .coral) { model.confirming = true }
        })
        .toolbar(.hidden, for: .navigationBar)
        .navigationDestination(isPresented: $model.confirming) {
            DeleteConfirmView(model: model)
        }
    }
}

/// D11: `Delete my account` stays at 40% until the field holds exactly `DELETE`; a failure shows its line above it.
struct DeleteConfirmView: View {
    @Bindable var model: DeleteAccountViewModel
    @FocusState private var focused: Bool

    var body: some View {
        Sheet(title: "Delete account", back: { model.confirming = false }, content: {
            Card {
                MMRTextField(
                    label: "Type DELETE to confirm",
                    text: $model.typed,
                    focused: focused,
                    capitalization: .characters,
                    submit: .done
                )
                .focused($focused)
                if model.failed {
                    InlineError(text: DeleteAccountViewModel.failedCopy)
                }
            }
            MMRButton(title: "Delete my account", variant: .coral, loading: model.deleting) {
                Task { await model.delete() }
            }
            .disabled(!model.canDelete)
        })
        .toolbar(.hidden, for: .navigationBar)
    }
}
