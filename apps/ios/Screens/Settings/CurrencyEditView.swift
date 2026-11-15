import SwiftUI

struct CurrencyEditView: View {
    @State private var model: CurrencyEditViewModel
    @FocusState private var searching: Bool
    @Environment(\.dismiss) private var dismiss

    init(model: CurrencyEditViewModel) {
        _model = State(initialValue: model)
    }

    var body: some View {
        Sheet(title: "Currency", back: { dismiss() }, content: {
            Card {
                if let refusal = model.target == nil ? model.refusal : nil {
                    InlineError(text: refusal)
                }
                CurrencyPicker(
                    query: $model.query,
                    visible: model.visible,
                    noMatch: model.noMatch,
                    selected: model.current,
                    value: model.changedValue,
                    listHeight: 336,
                    searching: $searching,
                    choose: { code in Task { await model.choose(code) } },
                    clear: { model.query = "" }
                )
            }
        })
        .toolbar(.hidden, for: .navigationBar)
        .sheet(isPresented: $model.sheetOpen) { ConvertSheet(model: model) }
    }
}

struct ConvertSheet: View {
    let model: CurrencyEditViewModel
    @State private var typed = ""

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: Tokens.Space._6) {
                Text(model.title)
                    .textStyle(Tokens.TypeRamp.titleM)
                    .accessibilityAddTraits(.isHeader)
                switch model.step {
                case .failed:
                    InlineError(text: "We couldn't get today's rate.", action: ("Try again", {
                        Task { await model.tryAgain() }
                    }))
                    paragraph("All your amounts will be converted once, at the rate we use.")
                    MMRButton(title: "Enter it myself", loading: model.isBusy, action: model.enterItMyself)
                case .confirm:
                    paragraph([model.rateLine, model.salaryLine].compactMap(\.self).joined(separator: " "))
                    rows
                    convert
                case .manual:
                    rateField
                    paragraph(model.preview == nil ? "All your amounts will be converted at the rate you enter."
                        : model.salaryLine ?? "")
                    rows
                    convert
                }
                MMRButton(title: "Cancel", variant: .secondary, action: model.cancel)
            }
            .padding(.horizontal, Tokens.Space._8)
            .padding(.top, Tokens.Space._9)
            .padding(.bottom, Tokens.Space._9)
        }
        .scrollDismissesKeyboard(.immediately)
        .foregroundStyle(Tokens.Text.primary)
        .presentationDetents([.medium, .large])
        .presentationDragIndicator(.visible)
        .presentationCornerRadius(Tokens.Radius.card)
        .presentationBackground(Tokens.Surface.sheet)
    }

    private func paragraph(_ line: String) -> some View {
        let lead = line.isEmpty ? Text("") : Text("\(line) ").foregroundStyle(Tokens.Text.secondary)
        return (lead + Text("This can't be undone.").foregroundStyle(Tokens.Text.primary))
            .textStyle(Tokens.TypeRamp.bodyM).fontWeight(.heavy)
            .fixedSize(horizontal: false, vertical: true)
    }

    @ViewBuilder private var rows: some View {
        if !model.rows.isEmpty {
            VStack(spacing: 0) {
                ForEach(model.rows, id: \.label) { row in
                    ListRow(label: row.label, value: row.value, trailing: .none)
                }
            }
            .padding(.horizontal, 18)
            .padding(.vertical, 4)
            .background(
                Tokens.Surface.field,
                in: RoundedRectangle(cornerRadius: Tokens.Radius.card, style: .continuous)
            )
        }
    }

    @ViewBuilder private var convert: some View {
        if let refusal = model.refusal {
            InlineError(text: refusal)
        }
        MMRButton(title: "Convert", loading: model.isBusy) { Task { await model.convert() } }
            .disabled(!model.canConvert)
    }

    private var rateField: some View {
        MMRTextField(
            label: "Rate: 1 \(model.current) =",
            text: $typed,
            placeholder: "0",
            helper: "We couldn't fetch this rate. Enter it yourself to continue.",
            suffix: model.target,
            keyboard: .decimalPad
        )
        .onChange(of: typed) { _, now in
            model.type(now)
            if typed != model.rate {
                typed = model.rate
            }
        }
    }
}
