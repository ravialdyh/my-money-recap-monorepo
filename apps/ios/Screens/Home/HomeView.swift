import MoneyRecapContract
import SwiftData
import SwiftUI

/// PRD §6.8 / SCREEN-STATES B1 at zero: the hero, the target card filled or its prompt, and the salary prompt while no
/// Salary is set (D54, D69). The prompts open steps 3 and 4 (`open`); the chips and Log a spend navigate nowhere yet.
struct HomeView: View {
    @State private var model: HomeViewModel
    private var open: (Int) -> Void = { _ in }

    /// The launch sends the zone (`LaunchView`); Home asks only for itself, once more after the interceptor refreshes
    /// the token on `unauthenticated` (IOS I3), and hands `onboardingNeeded` back to the launch when refused.
    init(
        session: AuthSession,
        context: ModelContext,
        open: @escaping (Int) -> Void,
        onboardingNeeded: @escaping () -> Void
    ) {
        let client = BffClient.make(session: session)
        self.init(model: HomeViewModel(fetch: {
            try await BffClient.retried(session) { await client.getHome(request: .init()) }.result.get().home
        }, context: context, onboardingNeeded: onboardingNeeded))
        self.open = open
    }

    /// Home over a given model, as the render tests draw it.
    init(model: HomeViewModel) {
        _model = State(initialValue: model)
    }

    var body: some View {
        Screen {
            ScrollView {
                VStack(alignment: .leading, spacing: Tokens.Space._5) {
                    HStack(spacing: 8) {
                        if let month = model.monthLabel {
                            Chip(text: month, tappable: true)
                        }
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
        Card(minHeight: Tokens.Size.homeHeroCard) {
            StatHero(eyebrow: model.heroEyebrow, value: Money.format(home.spent.amount, currency: home.currencyCode))
            HStack(spacing: 8) {
                StatChip(
                    tone: .earned,
                    text: "Money in \(Money.format(home.moneyIn.amount, currency: home.currencyCode))"
                )
                if model.showsSavedChip {
                    StatChip(
                        tone: .saved,
                        text: "Saved \(Money.format(home.saved.amount, currency: home.currencyCode))"
                    )
                }
            }
        }
        if !home.hasSalary {
            PromptCard(
                eyebrow: "Salary",
                title: "Set your monthly salary",
                helper: "We'll show what you kept once we know what came in.",
                height: Tokens.Size.homeHeroCard
            ) { open(3) }
        }
        if home.hasSavingTarget {
            target(home)
        } else {
            PromptCard(
                title: "Set a saving target",
                helper: "Pick a share of your salary to keep each month.",
                height: Tokens.Size.homeTargetCard
            ) { open(4) }
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

    /// PRD §6.8 target card: "{Saved} of {target amount}", the bar filled by Saved, and ledger's §8.7 sentence.
    private func target(_ home: MoneyRecap_V1_HomeView) -> some View {
        Card(minHeight: Tokens.Size.homeTargetCard) {
            HStack(alignment: .firstTextBaseline) {
                CardTitle(text: "Saving target")
                Spacer()
                Text(model.targetProgress ?? "")
                    .textStyle(Tokens.TypeRamp.labelM)
                    .monospacedDigit()
                    .foregroundStyle(Tokens.Text.secondary)
            }
            ProgressBar(fraction: model.targetFraction, tone: .green)
            Text(home.savingTarget.progressSentence)
                .textStyle(Tokens.TypeRamp.bodyS)
                .foregroundStyle(Tokens.Text.secondary)
        }
    }

    /// PRD §6.19 first load: both cards at their filled heights with `ink.100` blocks in place of numbers. No spinner.
    private var skeleton: some View {
        Group {
            Card(minHeight: Tokens.Size.homeHeroCard) {
                block(width: 96, height: 16)
                block(width: 180, height: 64)
                block(width: 120, height: 34)
            }
            Card(minHeight: Tokens.Size.homeTargetCard) {
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
