import SwiftUI

/// PRD §6.5 / SCREEN-STATES A5, step 2 of onboarding: the back chevron to step 1 and the step pill, the question, the
/// search and the list in one card, then Next.
struct CurrencyView: View {
    @State private var model: CurrencyViewModel
    private let onBack: () -> Void
    private let onSaved: () -> Void
    @FocusState private var searching: Bool

    init(model: CurrencyViewModel, onBack: @escaping () -> Void, onSaved: @escaping () -> Void) {
        _model = State(initialValue: model)
        self.onBack = onBack
        self.onSaved = onSaved
    }

    var body: some View {
        Screen {
            BackRow(label: "Step 2 of 4", back: onBack)
            Headline(text: "Which currency?").padding(.top, 18)
            Card {
                if model.offline {
                    InlineError(text: NoConnectionView.copy)
                }
                search
                if let noMatch = model.noMatch {
                    empty(noMatch)
                } else {
                    list
                }
            }
            .padding(.vertical, 20)
            if model.noMatch != nil {
                Spacer(minLength: 0)
            }
            MMRButton(title: "Next", loading: model.isSaving) {
                searching = false
                Task {
                    if await model.next() {
                        onSaved()
                    }
                }
            }
            .disabled(!model.canContinue)
        }
    }

    /// The canvas's search pill: a magnifier and the field on the field surface, ringed violet while typing.
    private var search: some View {
        HStack(spacing: 10) {
            Image(systemName: "magnifyingglass")
                .font(.system(size: 16, weight: .bold))
                .foregroundStyle(Tokens.Text.secondary)
                .accessibilityHidden(true)
            TextField("Search by name or code", text: $model.query)
                .textStyle(Tokens.TypeRamp.bodyM)
                .foregroundStyle(Tokens.Text.primary)
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
                .submitLabel(.search)
                .focused($searching)
                .accessibilityLabel("Search by name or code")
        }
        .padding(.horizontal, 18)
        .frame(maxWidth: .infinity, minHeight: Tokens.Size.fieldHeight)
        .background(Tokens.Surface.field, in: Capsule())
        .overlay(Capsule().strokeBorder(searching ? Tokens.State.focusRing : Tokens.Palette.ink100, lineWidth: 2))
    }

    /// Every matching row, scrolled so the chosen one shows; a tap chooses and puts the keyboard away, uncovering Next.
    /// VoiceOver reads the name before the code.
    private var list: some View {
        ScrollViewReader { reader in
            ScrollView {
                VStack(spacing: 0) {
                    ForEach(model.visible) { option in
                        ListRow(label: option.label, trailing: .check, selected: model.selected == option.code) {
                            model.choose(option.code)
                            searching = false
                        }
                        .accessibilityLabel("\(option.name), \(option.code)")
                        .id(option.code)
                    }
                }
            }
            .scrollIndicators(.hidden)
            .scrollDismissesKeyboard(.immediately)
            .onAppear {
                if let selected = model.selected {
                    reader.scrollTo(selected, anchor: .center)
                }
            }
        }
    }

    /// A5 empty search: the line and the link that clears the field; the chosen row stays chosen underneath.
    private func empty(_ line: String) -> some View {
        VStack(spacing: 10) {
            Text(line)
                .textStyle(Tokens.TypeRamp.bodyS)
                .foregroundStyle(Tokens.Text.secondary)
                .multilineTextAlignment(.center)
            TextLink(title: "Clear search") { model.clearSearch() }
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 48)
    }
}
