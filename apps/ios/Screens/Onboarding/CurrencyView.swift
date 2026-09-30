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
            if height == .compact {
                HStack(alignment: .top, spacing: Tokens.Space._8) {
                    VStack(alignment: .leading, spacing: 0) {
                        header
                        Spacer(minLength: 20)
                        next
                    }
                    .frame(maxWidth: 280)
                    card
                }
                .padding(.bottom, 4)
            } else {
                header
                card.padding(.vertical, 20)
                if model.noMatch != nil {
                    Spacer(minLength: 0)
                }
                next
            }
        }
    }

    /// Held sideways, the question and Next sit beside the card, so the list keeps the height a stacked column
    /// would spend on them (PRD §11 "Platform").
    @Environment(\.verticalSizeClass) private var height

    private var header: some View {
        VStack(alignment: .leading, spacing: 0) {
            BackRow(label: "Step 2 of 4", back: onBack)
            Headline(text: "Which currency?").padding(.top, 18)
        }
    }

    private var card: some View {
        Card {
            if model.offline {
                InlineError(text: NoConnectionView.copy)
            }
            CurrencyPicker(
                query: $model.query,
                visible: model.visible,
                noMatch: model.noMatch,
                selected: model.selected,
                searching: $searching,
                choose: model.choose,
                clear: model.clearSearch
            )
        }
    }

    private var next: some View {
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

/// A5's search pill over the matching rows, or its empty line and Clear search; D5 reuses it with `value` beside the
/// chosen row, its `Changed ✓` (SCREEN-STATES D5).
struct CurrencyPicker: View {
    @Binding var query: String
    let visible: [CurrencyViewModel.Option]
    let noMatch: String?
    let selected: String?
    var value: String?
    /// D5's fixed list height, the canvas's 336pt, so the sheet around it scrolls; nil fills what the step leaves.
    var listHeight: CGFloat?
    var searching: FocusState<Bool>.Binding
    let choose: (String) -> Void
    let clear: () -> Void

    var body: some View {
        search
        if let noMatch {
            empty(noMatch)
        } else {
            list
        }
    }

    /// The canvas's search pill: a magnifier and the field on the field surface, ringed violet while typing.
    private var search: some View {
        HStack(spacing: 10) {
            Image(systemName: "magnifyingglass")
                .font(.system(size: 16, weight: .bold))
                .foregroundStyle(Tokens.Text.secondary)
                .accessibilityHidden(true)
            TextField("Search by name or code", text: $query)
                .textStyle(Tokens.TypeRamp.bodyM)
                .foregroundStyle(Tokens.Text.primary)
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
                .submitLabel(.search)
                .focused(searching)
                .accessibilityLabel("Search by name or code")
        }
        .padding(.horizontal, 18)
        .frame(maxWidth: .infinity, minHeight: Tokens.Size.fieldHeight)
        .background(Tokens.Surface.field, in: Capsule())
        .overlay(Capsule().strokeBorder(
            searching.wrappedValue ? Tokens.State.focusRing : Tokens.Palette.ink100,
            lineWidth: 2
        ))
    }

    /// Every matching row, scrolled so the chosen one shows; a tap chooses and puts the keyboard away, uncovering Next.
    /// VoiceOver reads the name before the code.
    private var list: some View {
        ScrollViewReader { reader in
            ScrollView {
                VStack(spacing: 0) {
                    ForEach(visible) { option in
                        ListRow(
                            label: option.label,
                            value: selected == option.code ? value : nil,
                            trailing: .check,
                            selected: selected == option.code
                        ) {
                            choose(option.code)
                            searching.wrappedValue = false
                        }
                        .accessibilityLabel("\(option.name), \(option.code)")
                        .id(option.code)
                    }
                }
            }
            .frame(height: listHeight)
            .scrollIndicators(.hidden)
            .scrollDismissesKeyboard(.immediately)
            .onAppear {
                if let selected {
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
            TextLink(title: "Clear search", action: clear)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 48)
    }
}
