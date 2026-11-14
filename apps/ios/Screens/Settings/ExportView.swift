import SwiftUI

/// D9 Export spends (PRD §6.16): one card naming the address, its failure line inside it, then `Send it`. It scrolls
/// sideways.
struct ExportView: View {
    @State private var model: ExportViewModel
    @Environment(\.dismiss) private var dismiss

    init(model: ExportViewModel) {
        _model = State(initialValue: model)
    }

    var body: some View {
        Sheet(title: "Export spends", back: { dismiss() }, content: {
            Card {
                CardTitle(text: "Email me a CSV of every spend")
                Text(model.helper)
                    .textStyle(Tokens.TypeRamp.bodyS)
                    .foregroundStyle(Tokens.Text.secondary)
                    .fixedSize(horizontal: false, vertical: true)
                if let error = model.error {
                    InlineError(text: error)
                }
            }
            MMRButton(title: model.button, loading: model.state == .sending) {
                Task { await model.commit() }
            }
            .disabled(!model.canSend)
        })
        .toolbar(.hidden, for: .navigationBar)
    }
}
