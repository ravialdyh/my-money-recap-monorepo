import SwiftUI

/// PRD §6.6 / SCREEN-STATES A6, step 3 of onboarding: the back chevron and the step pill, the question, the
/// amount card and the Payday card, then Next and Skip for now.
struct SalaryView: View {
    @State private var model: SalaryViewModel
    private let onBack: () -> Void
    private let onDone: () -> Void

    init(model: SalaryViewModel, onBack: @escaping () -> Void, onDone: @escaping () -> Void) {
        _model = State(initialValue: model)
        self.onBack = onBack
        self.onDone = onDone
    }

    var body: some View {
        Screen {
            ScrollView {
                VStack(alignment: .leading, spacing: 0) {
                    BackRow(label: "Step 3 of 4", back: onBack)
                    Headline(text: "Your monthly salary").padding(.top, 18)
                    VStack(spacing: Tokens.Space._4) {
                        Card {
                            if model.offline {
                                InlineError(text: NoConnectionView.copy)
                            }
                            AmountField(
                                label: "Amount, after tax",
                                currency: model.currency,
                                digits: model.digits,
                                amount: $model.amount
                            )
                        }
                        Card {
                            CardTitle(text: "Payday")
                            DayPicker(selected: model.payday, nextPaydays: model.nextPaydays, choose: model.choose)
                        }
                    }
                    .padding(.top, 20)
                }
            }
            .scrollIndicators(.hidden)
            .scrollDismissesKeyboard(.immediately)
            VStack(spacing: Tokens.Space._1) {
                MMRButton(title: "Next", loading: model.isSaving) {
                    finish(model.next)
                }
                .disabled(!model.canContinue)
                GroundLink(title: "Skip for now") { finish(model.skip) }
                    .disabled(model.isSaving)
            }
        }
    }

    private func finish(_ save: @escaping () async -> Bool) {
        UIApplication.shared.sendAction(#selector(UIResponder.resignFirstResponder), to: nil, from: nil, for: nil)
        Task {
            if await save() {
                onDone()
            }
        }
    }
}
