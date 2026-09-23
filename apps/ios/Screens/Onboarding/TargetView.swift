import SwiftUI

/// PRD §6.7 / SCREEN-STATES A7, step 4 of onboarding: the back chevron and the step pill, the question, then the slider
/// card with Done and Skip for now, or, without a Salary (D44), the prompt back to step 3 and Skip for now alone.
struct TargetView: View {
    @State private var model: TargetViewModel
    private let onBack: () -> Void
    private let onSetSalary: () -> Void
    private let onDone: () -> Void

    init(
        model: TargetViewModel,
        onBack: @escaping () -> Void,
        onSetSalary: @escaping () -> Void,
        onDone: @escaping () -> Void
    ) {
        _model = State(initialValue: model)
        self.onBack = onBack
        self.onSetSalary = onSetSalary
        self.onDone = onDone
    }

    var body: some View {
        Screen(scrolls: true) {
            BackRow(label: "Step 4 of 4", back: onBack)
            Headline(text: "How much do you want to keep?").padding(.top, 18)
            VStack(spacing: Tokens.Space._4) {
                if model.offline, model.isBlocked {
                    Card { InlineError(text: NoConnectionView.copy) }
                }
                if model.isBlocked {
                    PromptCard(
                        title: "Set a salary first",
                        helper: "A target is a share of your salary, so we need the salary first.",
                        action: onSetSalary
                    )
                } else {
                    slider
                }
            }
            .padding(.top, 20)
            Spacer(minLength: 0)
            if model.isBlocked {
                MMRButton(title: "Skip for now", loading: model.isSaving) { finish(model.skip) }
                    .disabled(model.isSaving)
            } else {
                VStack(spacing: Tokens.Space._1) {
                    MMRButton(title: "Done", loading: model.isSaving) { finish(model.done) }
                        .disabled(model.isSaving)
                    GroundLink(title: "Skip for now") { finish(model.skip) }
                        .disabled(model.isSaving)
                }
            }
        }
    }

    private var slider: some View {
        Card {
            if model.offline {
                InlineError(text: NoConnectionView.copy)
            }
            Text("Share of salary to keep")
                .textStyle(Tokens.TypeRamp.labelS)
                .foregroundStyle(Tokens.Text.secondary)
            Text("\(model.percent)%")
                .textStyle(Tokens.TypeRamp.displayXl)
                .monospacedDigit()
                .accessibilityHidden(true)
            Slider(label: "Saving target", value: $model.percent)
            HStack {
                Text("0%")
                Spacer()
                Text("100%")
            }
            .textStyle(Tokens.TypeRamp.labelS)
            .foregroundStyle(Tokens.Text.secondary)
            .accessibilityHidden(true)
            if let preview = model.preview {
                Text(preview)
                    .textStyle(Tokens.TypeRamp.bodyS)
                    .monospacedDigit()
                    .foregroundStyle(Tokens.Text.secondary)
            }
        }
    }

    private func finish(_ save: @escaping () async -> Bool) {
        Task {
            if await save() {
                onDone()
            }
        }
    }
}
