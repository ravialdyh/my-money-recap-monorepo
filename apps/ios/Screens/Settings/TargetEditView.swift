import SwiftUI

struct TargetEditView: View {
    @State private var model: TargetEditViewModel
    @Environment(\.dismiss) private var dismiss
    private let onSetSalary: () -> Void
    private let onSaved: () -> Void

    init(model: TargetEditViewModel, onSetSalary: @escaping () -> Void = {}, onSaved: @escaping () -> Void) {
        _model = State(initialValue: model)
        self.onSetSalary = onSetSalary
        self.onSaved = onSaved
    }

    var body: some View {
        Sheet(title: "Saving target", back: { dismiss() }, content: {
            if model.isBlocked {
                PromptCard(
                    title: "Set a salary first",
                    helper: "A target is a share of your salary, so we need the salary first.",
                    action: onSetSalary
                )
            } else {
                TargetCard(percent: $model.percent, preview: model.preview)
                if let starts = model.startsLine {
                    Text(starts)
                        .textStyle(Tokens.TypeRamp.bodyS)
                        .foregroundStyle(Tokens.Text.secondary)
                        .fixedSize(horizontal: false, vertical: true)
                }
                if model.failed {
                    InlineError(text: SettingsViewModel.saveFailure)
                }
                MMRButton(title: "Save", loading: model.isSaving) {
                    Task {
                        if await model.commit() {
                            onSaved()
                        }
                    }
                }
                .disabled(!model.canSave)
            }
        })
        .toolbar(.hidden, for: .navigationBar)
    }
}
