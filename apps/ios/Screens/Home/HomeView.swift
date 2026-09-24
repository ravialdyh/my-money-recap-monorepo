import MoneyRecapContract
import SwiftData
import SwiftUI

/// PRD §6.8 / SCREEN-STATES B1: the hero, the Target Warning, the target card or its prompt, the salary prompt (D54,
/// D69). The prompts open steps 3 and 4, Log a spend its sheet, the hero and month chip the Spend list (D59).
struct HomeView: View {
    @State private var model: HomeViewModel
    @State private var logging: LogSpendViewModel?
    @State private var listing: SpendListViewModel?
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
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
        }, context: context, onboardingNeeded: onboardingNeeded, logSpend: { request, headers in
            await BffClient.retried(session) { await client.logSpend(request: request, headers: headers) }.result
        }, listSpends: { month in
            try await BffClient.retried(session) { await client.listSpends(request: .with { $0.month = month }) }
                .result.get().spends
        }))
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
                            Chip(text: month, tappable: true) { listing = model.spendList(switching: true) }
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
        .contentTransition(reduceMotion ? .opacity : .numericText())
        .task { await model.refresh() }
        .sheet(item: $logging) { sheet in
            LogSpendView(model: sheet) { home in
                logging = nil
                withAnimation(reveal) { model.replace(with: home) }
            }
        }
        .sheet(item: $listing) { list in
            SpendListView(model: list, log: { model.logSheet() }, onSaved: { home in
                withAnimation(reveal) { model.replace(with: home) }
            })
        }
    }

    /// `motion.reveal.number` for the numbers a save changed; with Reduce Motion, the 150 ms fade.
    private var reveal: Animation {
        reduceMotion
            ? .easeInOut(duration: Tokens.Motion.Duration.reduced)
            : .timingCurve(Tokens.Motion.Easing.standard, duration: Tokens.Motion.Duration.revealNumber)
            .delay(Tokens.Motion.Duration.revealNumberDelay)
    }

    @ViewBuilder
    private func filled(_ home: MoneyRecap_V1_HomeView) -> some View {
        if let greeting = model.greeting {
            HStack(spacing: 12) {
                Headline(text: greeting)
                Spacer()
                Avatar(key: home.avatarKey)
            }
        }
        Card(minHeight: Tokens.Size.homeHeroCard) {
            StatHero(eyebrow: model.heroEyebrow, value: model.money(home.spent))
            StatChipRow {
                StatChip(tone: .earned, text: "Money in \(model.money(home.moneyIn))")
                if model.showsSavedChip {
                    StatChip(tone: .saved, text: "Saved \(model.money(home.saved))")
                }
            }
        }
        .contentShape(Rectangle())
        .onTapGesture { listing = model.spendList(switching: false) }
        .accessibilityAction(named: "Show spends") { listing = model.spendList(switching: false) }
        if !home.hasSalary {
            PromptCard(
                eyebrow: "Salary",
                title: "Set your monthly salary",
                helper: "We'll show what you kept once we know what came in.",
                height: Tokens.Size.homeHeroCard
            ) { open(3) }
        }
        if let warning = model.warning {
            Banner(text: warning) { withAnimation(reveal) { model.dismissWarning() } }
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
        MMRButton(title: "Log a spend", icon: "plus") { logging = model.logSheet() }
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
            ViewThatFits(in: .horizontal) {
                HStack(alignment: .firstTextBaseline) {
                    CardTitle(text: "Saving target")
                    Spacer()
                    progress
                }
                VStack(alignment: .leading, spacing: Tokens.Space._1) {
                    CardTitle(text: "Saving target")
                    progress
                }
            }
            ProgressBar(fraction: model.targetFraction, tone: .green)
            Text(home.savingTarget.progressSentence)
                .textStyle(Tokens.TypeRamp.bodyS)
                .foregroundStyle(Tokens.Text.secondary)
        }
    }

    /// "{Saved} of {target amount}", one line; beside the title while both fit, under it otherwise (X3).
    private var progress: some View {
        Text(model.targetProgress ?? "")
            .textStyle(Tokens.TypeRamp.labelM)
            .monospacedDigit()
            .lineLimit(1)
            .fixedSize()
            .foregroundStyle(Tokens.Text.secondary)
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
