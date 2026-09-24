import MoneyRecapContract
import SwiftUI

/// PRD §6.9, §6.10 / SCREEN-STATES B2 to B4, B7, B8: the amount, the eight Category chips, the Date row and its wheel
/// picker, the note, the one failure line and Save; editing, Save changes and the Delete row. It scrolls sideways.
struct LogSpendView: View {
    @State private var model: LogSpendViewModel
    @State private var picking = false
    @State private var confirming = false
    private let onSaved: (MoneyRecap_V1_HomeView) -> Void

    init(model: LogSpendViewModel, onSaved: @escaping (MoneyRecap_V1_HomeView) -> Void) {
        _model = State(initialValue: model)
        self.onSaved = onSaved
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: Tokens.Space._6) {
                Text(model.spend == nil ? "Log a spend" : "Edit spend")
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
                dateRow
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
                if model.deleteFailed {
                    InlineError(text: LogSpendViewModel.deleteFailure)
                }
                MMRButton(title: model.spend == nil ? "Save" : "Save changes", loading: model.isSaving) { save() }
                    .disabled(!model.canSave)
                if model.spend != nil {
                    ListRow(label: "Delete this spend", trailing: .none, destructive: true) {
                        endEditing()
                        confirming = true
                    }
                    .disabled(model.isSaving)
                    .confirmationDialog("Delete this spend?", isPresented: $confirming, titleVisibility: .visible) {
                        Button("Delete", role: .destructive) { remove() }
                        Button("Cancel", role: .cancel) {}
                    } message: {
                        Text("You can't undo this.")
                    }
                }
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

    /// The Date row opening the wheel picker, or, for a Spend older than the picker's days, its date as read-only text.
    private var dateRow: some View {
        VStack(alignment: .leading, spacing: 0) {
            if model.dateLocked {
                HStack(spacing: 12) {
                    Text("Date").textStyle(Tokens.TypeRamp.labelM)
                    Spacer(minLength: 0)
                    Text(model.dateLabel).textStyle(Tokens.TypeRamp.labelM).foregroundStyle(Tokens.Text.secondary)
                }
                .padding(.vertical, Tokens.Space._5)
                .accessibilityElement(children: .combine)
                helper(LogSpendViewModel.oldDateHelper)
            } else {
                ListRow(label: "Date", value: model.dateLabel) {
                    endEditing()
                    picking.toggle()
                }
                if picking {
                    DatePicker("Date", selection: $model.date, in: model.dates, displayedComponents: .date)
                        .datePickerStyle(.wheel)
                        .labelsHidden()
                        .frame(maxWidth: .infinity)
                    helper(LogSpendViewModel.dateHelper)
                }
            }
        }
    }

    private func helper(_ text: String) -> some View {
        Text(text)
            .textStyle(Tokens.TypeRamp.bodyS)
            .foregroundStyle(Tokens.Text.secondary)
            .fixedSize(horizontal: false, vertical: true)
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

    private func remove() {
        Task {
            if let home = await model.deleteSpend() {
                onSaved(home)
            }
        }
    }
}
