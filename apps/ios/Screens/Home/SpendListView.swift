import MoneyRecapContract
import SwiftUI

/// B6 Spend list (PRD §6.10, D59), the full-height sheet over Home: the month switcher, one row per Category that opens
/// B6b, or the empty month's line and the plus button that opens Log a spend. It scrolls, so it works sideways.
struct SpendListView: View {
    @State private var model: SpendListViewModel
    @State private var logging: LogSpendViewModel?
    @Environment(\.dismiss) private var dismiss
    private let log: () -> LogSpendViewModel?
    private let onSaved: (MoneyRecap_V1_HomeView) -> Void

    /// `log` is Home's Log a spend sheet; `onSaved` hands its fresh Home back, and the list asks for its month again.
    init(
        model: SpendListViewModel,
        log: @escaping () -> LogSpendViewModel? = { nil },
        onSaved: @escaping (MoneyRecap_V1_HomeView) -> Void = { _ in }
    ) {
        _model = State(initialValue: model)
        self.log = log
        self.onSaved = onSaved
    }

    var body: some View {
        NavigationStack {
            Screen(scrolls: true) {
                BackRow(label: "Spends") { dismiss() }.padding(.top, Tokens.Space._6)
                Headline(text: "Spends")
                    .accessibilityAddTraits(.isHeader)
                    .padding(.top, Tokens.Space._6)
                Chip(text: model.label(model.month), tappable: true, opens: true) { model.switching.toggle() }
                    .padding(.top, Tokens.Space._4)
                if let line = model.monthLine {
                    SubCopy(text: line).padding(.top, Tokens.Space._4)
                }
                content
                    .opacity(model.switching ? Tokens.State.disabledOpacity : 1)
                    .disabled(model.switching)
                    .overlay(alignment: .top) {
                        if model.switching {
                            switcher.padding(.top, Tokens.Space._3)
                        }
                    }
            }
            .toolbar(.hidden, for: .navigationBar)
            .navigationDestination(item: $model.category) { CategoryView(model: model, category: $0) }
        }
        .task { await model.refresh() }
        .sheet(item: $logging) { sheet in
            LogSpendView(model: sheet) { home in
                logging = nil
                onSaved(home)
                Task { await model.refresh() }
            }
        }
        .presentationDetents([.large])
        .presentationDragIndicator(.visible)
        .presentationCornerRadius(Tokens.Radius.card)
        .presentationBackground(Tokens.Ground.recap)
    }

    @ViewBuilder
    private var content: some View {
        if let line = model.emptyLine {
            VStack(spacing: Tokens.Space._7) {
                Text(line)
                    .textStyle(Tokens.TypeRamp.titleM)
                    .multilineTextAlignment(.center)
                    .frame(maxWidth: 260)
                    .fixedSize(horizontal: false, vertical: true)
                Button { logging = log() } label: {
                    Image(systemName: "plus")
                        .font(.system(size: 20, weight: .bold))
                        .foregroundStyle(Tokens.Text.onInverse)
                        .frame(width: Tokens.Space._10, height: Tokens.Space._10)
                        .background(Tokens.Surface.buttonPrimary, in: Circle())
                }
                .buttonStyle(PressableStyle())
                .accessibilityLabel("Log a spend")
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, Tokens.Space._10 * 2)
        } else if let list = model.list {
            Card {
                VStack(spacing: 0) {
                    ForEach(list.categories, id: \.category) { share in
                        let glyph = SpendListViewModel.glyph(share.category)
                        DataRow(
                            icon: glyph,
                            label: glyph.label,
                            detail: model.count(share),
                            amount: model.money(share.amount)
                        ) {
                            model.category = glyph
                        }
                    }
                }
            }
            .padding(.top, Tokens.Space._7)
        }
    }

    /// The open switcher: every month with data, newest first, the shown one checked (`ListRow trailing="check"`).
    private var switcher: some View {
        Card {
            VStack(spacing: 0) {
                ForEach(model.months, id: \.self) { month in
                    ListRow(label: model.label(month), trailing: .check, selected: month == model.month) {
                        Task { await model.pick(month) }
                    }
                }
            }
        }
    }
}

/// B6b Category detail: that Category's Spends, newest first, each its note or Category, its date and amount, under its
/// share of the month. Back returns to the list.
struct CategoryView: View {
    let model: SpendListViewModel
    let category: Icon.Name
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        Screen(scrolls: true) {
            BackRow(label: model.label(model.month)) { dismiss() }.padding(.top, Tokens.Space._6)
            HStack(spacing: Tokens.Space._4) {
                IconWell(tone: category)
                Headline(text: category.label).accessibilityAddTraits(.isHeader)
            }
            .padding(.top, Tokens.Space._6)
            if let line = model.shareLine(category) {
                SubCopy(text: line).padding(.top, Tokens.Space._4)
            }
            Card {
                VStack(spacing: 0) {
                    ForEach(model.spends(category), id: \.id) { spend in
                        DataRow(
                            icon: category,
                            label: model.title(spend),
                            detail: model.day(spend),
                            amount: model.money(spend.amount)
                        )
                    }
                }
            }
            .padding(.top, Tokens.Space._7)
        }
        .toolbar(.hidden, for: .navigationBar)
    }
}
