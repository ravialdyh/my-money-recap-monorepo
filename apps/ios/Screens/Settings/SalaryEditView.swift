import MoneyRecapContract
import SwiftUI

/// D3 Salary and payday (PRD §6.16): setup's amount card, any waiting Salary's line, the Payday card and its start
/// line, then Save, which returns to Settings or opens "Which salary is this?". It scrolls sideways.
struct SalaryEditView: View {
    @State private var model: SalaryEditViewModel
    @Environment(\.dismiss) private var dismiss
    private let onSaved: () -> Void

    init(model: SalaryEditViewModel, onSaved: @escaping () -> Void) {
        _model = State(initialValue: model)
        self.onSaved = onSaved
    }

    var body: some View {
        Sheet(title: "Salary and payday", back: { dismiss() }, content: {
            Card {
                AmountField(
                    label: "Amount, after tax",
                    currency: model.currency,
                    digits: model.digits,
                    amount: $model.amount
                )
            }
            if let waiting = model.waitingLine {
                helper(waiting)
            }
            Card {
                CardTitle(text: "Payday")
                DayPicker(selected: model.payday, nextPaydays: model.nextPaydays, choose: model.choose)
            }
            if let starts = model.startsLine {
                helper(starts)
            }
            if model.failed {
                InlineError(text: SettingsViewModel.saveFailure)
            }
            MMRButton(title: "Save", loading: model.isSaving) { finish(model.commit) }
                .disabled(!model.canSave)
        })
        .toolbar(.hidden, for: .navigationBar)
        .sheet(isPresented: $model.asking) {
            WhichSalarySheet(model: model) { applies in
                finish { await model.answer(applies) }
            }
        }
    }

    private func helper(_ text: String) -> some View {
        Text(text)
            .textStyle(Tokens.TypeRamp.bodyS)
            .foregroundStyle(Tokens.Text.secondary)
            .fixedSize(horizontal: false, vertical: true)
    }

    private func finish(_ save: @escaping () async -> Bool) {
        UIApplication.shared.sendAction(#selector(UIResponder.resignFirstResponder), to: nil, from: nil, for: nil)
        Task {
            if await save() {
                onSaved()
            }
        }
    }
}

/// The D56 sheet: the body line, a tinted card of two rows, each with its consequence, and Cancel, which saves
/// nothing. It scrolls sideways.
struct WhichSalarySheet: View {
    let model: SalaryEditViewModel
    let choose: (MoneyRecap_V1_SalaryAppliesFrom) -> Void
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: Tokens.Space._6) {
                Text("Which salary is this?")
                    .textStyle(Tokens.TypeRamp.titleM)
                    .accessibilityAddTraits(.isHeader)
                Text(model.sheetBody)
                    .textStyle(Tokens.TypeRamp.bodyM).fontWeight(.heavy)
                    .foregroundStyle(Tokens.Text.secondary)
                    .fixedSize(horizontal: false, vertical: true)
                VStack(spacing: 0) {
                    ListRow(label: model.correctionRow, detail: model.correctionDetail) { choose(.thisMonth) }
                    ListRow(label: model.raiseRow, detail: model.raiseDetail) { choose(.nextMonth) }
                }
                .padding(.horizontal, 18)
                .padding(.vertical, 4)
                .background(
                    Tokens.Surface.field,
                    in: RoundedRectangle(cornerRadius: Tokens.Radius.card, style: .continuous)
                )
                MMRButton(title: "Cancel", variant: .secondary) { dismiss() }
            }
            .padding(.horizontal, Tokens.Space._8)
            .padding(.top, Tokens.Space._9)
            .padding(.bottom, Tokens.Space._9)
        }
        .foregroundStyle(Tokens.Text.primary)
        .presentationDetents([.medium, .large])
        .presentationDragIndicator(.visible)
        .presentationCornerRadius(Tokens.Radius.card)
        .presentationBackground(Tokens.Surface.sheet)
    }
}
