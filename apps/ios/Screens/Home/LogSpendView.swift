import MoneyRecapContract
import SwiftUI

/// PRD §6.9 / SCREEN-STATES B2 to B4, the sheet over Home: the amount field, the eight Category chips in a wrapping
/// row, the Date row opening the wheel picker, the note, the one failure line and Save. It scrolls, so it works
/// sideways.
struct LogSpendView: View {
    @State private var model: LogSpendViewModel
    @State private var picking = false
    private let onSaved: (MoneyRecap_V1_HomeView) -> Void

    init(model: LogSpendViewModel, onSaved: @escaping (MoneyRecap_V1_HomeView) -> Void) {
        _model = State(initialValue: model)
        self.onSaved = onSaved
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: Tokens.Space._6) {
                Text("Log a spend")
                    .textStyle(Tokens.TypeRamp.titleM)
                    .accessibilityAddTraits(.isHeader)
                AmountField(label: "Amount", currency: model.currency, digits: model.digits, amount: $model.amount)
                StatChipRow {
                    ForEach(Icon.Name.allCases, id: \.self) { category in
                        Chip(
                            text: category.label,
                            tappable: true,
                            tone: category,
                            selected: model.category == category
                        ) { model.choose(category) }
                    }
                }
                VStack(alignment: .leading, spacing: 0) {
                    ListRow(label: "Date", value: model.dateLabel) {
                        endEditing()
                        picking.toggle()
                    }
                    if picking {
                        DatePicker("Date", selection: $model.date, in: model.dates, displayedComponents: .date)
                            .datePickerStyle(.wheel)
                            .labelsHidden()
                            .frame(maxWidth: .infinity)
                        Text(LogSpendViewModel.dateHelper)
                            .textStyle(Tokens.TypeRamp.bodyS)
                            .foregroundStyle(Tokens.Text.secondary)
                            .fixedSize(horizontal: false, vertical: true)
                    }
                }
                MMRTextField(
                    label: "Note",
                    text: Binding(get: { model.note }, set: model.type),
                    placeholder: "What was it? (optional)",
                    capitalization: .sentences,
                    submit: .done
                )
                if model.failed {
                    InlineError(text: LogSpendViewModel.failure)
                }
                MMRButton(title: "Save", loading: model.isSaving) { save() }
                    .disabled(!model.canSave)
            }
            .padding(.horizontal, Tokens.Space._8)
            .padding(.top, Tokens.Space._9)
            .padding(.bottom, Tokens.Space._9)
        }
        .scrollDismissesKeyboard(.immediately)
        .foregroundStyle(Tokens.Text.primary)
        .presentationDetents([.large])
        .presentationDragIndicator(.visible)
        .presentationCornerRadius(Tokens.Radius.card)
        .presentationBackground(Tokens.Surface.sheet)
    }

    /// Puts the keyboard away, so the wheel picker and Save are not under it.
    private func endEditing() {
        UIApplication.shared.sendAction(#selector(UIResponder.resignFirstResponder), to: nil, from: nil, for: nil)
    }

    private func save() {
        endEditing()
        Task {
            if let home = await model.save() {
                onSaved(home)
            }
        }
    }
}
