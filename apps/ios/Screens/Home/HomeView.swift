import MoneyRecapContract
import SwiftData
import SwiftUI

/// PRD §6.8 / SCREEN-STATES B1, drawn with the design-system components by name. This slice draws the
/// "no salary set" layout at zero (D54, D69); the chips, Log a spend and the cards are not yet wired to screens.
struct HomeView: View {
    @State private var model: HomeViewModel

    init(session: AuthSession, context: ModelContext) {
        let client = BffClient.make(session: session)
        _model = State(initialValue: HomeViewModel(fetch: {
            var response = await client.getHome(request: .init())
            // The interceptor refreshed the token on `unauthenticated`; retry once while still signed in (I3).
            if response.code == .unauthenticated, session.isSignedIn {
                response = await client.getHome(request: .init())
            }
            return try response.result.get().home
        }, context: context))
    }

    var body: some View {
        Screen {
            ScrollView {
                VStack(alignment: .leading, spacing: 14) {
                    HStack(spacing: 8) {
                        if let month = model.monthLabel { Chip(text: month, tappable: true) }
                        Spacer()
                        Chip(text: "Settings", tappable: true)
                    }
                    .padding(.top, 8)
                    if let home = model.home {
                        filled(home)
                    } else {
                        skeleton
                    }
                }
                .padding(.bottom, 24)
            }
            .scrollIndicators(.hidden)
        }
        .task { await model.refresh() }
    }

    @ViewBuilder
    private func filled(_ home: MoneyRecap_V1_HomeView) -> some View {
        if let greeting = model.greeting {
            HStack(spacing: 12) {
                Headline(text: greeting, size: .m)
                Spacer()
                Avatar(key: home.avatarKey)
            }
        }
        Card(minHeight: 191) {
            StatHero(eyebrow: model.heroEyebrow, value: Money.format(home.spent.amount, currency: home.currencyCode))
            HStack(spacing: 8) {
                StatChip(tone: .earned, text: "Money in \(Money.format(home.moneyIn.amount, currency: home.currencyCode))")
                if model.showsSavedChip {
                    StatChip(tone: .saved, text: "Saved \(Money.format(home.saved.amount, currency: home.currencyCode))")
                }
            }
        }
        if !home.hasSalary {
            PromptCard(title: "Set your monthly salary", height: 191)
        }
        if home.hasSavingTarget {
            // The ProgressBar arrives with the Saving Target slice; until then the card carries the server's sentence.
            Card(minHeight: 136) {
                HStack {
                    CardTitle(text: "Saving target")
                    Spacer()
                    Text(Money.format(home.savingTarget.amount.amount, currency: home.currencyCode)).textStyle(Tokens.TypeRamp.labelM).foregroundStyle(Tokens.Text.secondary)
                }
                HelperText(text: home.savingTarget.progressSentence)
            }
        } else {
            PromptCard(title: "Set a saving target", height: 136)
        }
        MMRButton(title: "Log a spend", icon: "plus") {}
        MMRButton(title: "See last month", variant: .secondary) {}
            .disabled(home.recapMonths.isEmpty)
        if home.recapMonths.isEmpty {
            if let copy = model.previewCopy {
                PromptCard(title: copy, dashed: true)
            }
        } else {
            Card {
                ForEach(home.recapMonths, id: \.month) { month in
                    ListRow(label: model.recapLabel(month), value: month.label)
                }
            }
        }
    }

    /// PRD §6.19 first load: both cards at their filled heights with `ink.100` blocks in place of numbers. No spinner.
    private var skeleton: some View {
        Group {
            Card(minHeight: 191) {
                block(width: 96, height: 16)
                block(width: 180, height: 64)
                block(width: 120, height: 34)
            }
            Card(minHeight: 136) {
                block(width: 140, height: 22)
                block(width: 220, height: 16)
            }
        }
        .accessibilityElement(children: .ignore)
        .accessibilityIdentifier("home-skeleton")
    }

    private func block(width: CGFloat, height: CGFloat) -> some View {
        RoundedRectangle(cornerRadius: Tokens.Radius.inset, style: .continuous)
            .fill(Tokens.Palette.ink100)
            .frame(width: width, height: height)
    }
}
