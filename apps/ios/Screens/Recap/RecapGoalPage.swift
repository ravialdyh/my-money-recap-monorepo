import MoneyRecapContract
import SwiftUI

/// Page 3, C3 or P3 (PRD §6.13, §6.17): the coral ground, the target card by situation, then where the money went.
/// `active` is true while the page is on screen, which reveals its bars. Scrolls held sideways.
struct RecapGoalPage: View {
    @Bindable var model: RecapViewModel
    var active = true

    var body: some View {
        Screen(ground: Tokens.Ground.goal, scrolls: true) {
            StatChipRow {
                Chip(text: "Target", asLabel: true)
                Chip(text: "Spending", asLabel: true, inverse: true)
            }
            .padding(.leading, RecapLayout.chipShift)
            .padding(.top, Tokens.Space._3)
            Headline(text: model.goalHeadline, size: .compact)
                .padding(.top, Tokens.Space._6)
            target
            spending.padding(.top, Tokens.Space._4)
            PageDots(count: model.pages.count, active: $model.page)
                .padding(.top, Tokens.Space._3)
        }
    }

    @ViewBuilder private var target: some View {
        if let target = model.target {
            Card {
                ViewThatFits(in: .horizontal) {
                    HStack(alignment: .firstTextBaseline) {
                        CardTitle(text: "Saving target")
                        Spacer()
                        progress(target.progress)
                    }
                    VStack(alignment: .leading, spacing: Tokens.Space._1) {
                        CardTitle(text: "Saving target")
                        progress(target.progress)
                    }
                }
                ProgressBar(fraction: target.fraction, tone: .coral, piggy: true, shown: active)
                    .padding(.top, Tokens.Space._5)
                helper(model.answer?.recap.savingTarget.progressSentence ?? "")
            }
            .padding(.top, RecapLayout.cardGap)
        } else if let payday = model.paydayLine {
            Card {
                Eyebrow(text: "Saving target")
                CardTitle(text: payday)
            }
            .padding(.top, RecapLayout.cardGap)
        } else if !model.partial {
            PromptCard(
                title: "Set a saving target",
                helper: "Pick a share of your salary to keep each month.",
                height: Tokens.Size.homeTargetCard
            )
            .padding(.top, RecapLayout.cardGap)
        }
    }

    private var spending: some View {
        Card(split: true) {
            VStack(alignment: .leading, spacing: 10) {
                VStack(alignment: .leading, spacing: 0) {
                    Eyebrow(text: model.spendingEyebrow)
                    CardTitle(text: model.spendingTitle)
                }
                VStack(alignment: .leading, spacing: Tokens.Space._2) {
                    ForEach(model.bars, id: \.category) { bar in
                        CategoryBar(category: bar.category, percent: bar.percent, amount: bar.amount, shown: active)
                    }
                    if model.everyPenny {
                        helper("You kept every penny that came in.")
                    }
                    if let rollUp = model.rollUp {
                        helper(rollUp)
                    }
                }
                Spacer(minLength: 0)
                if let splurge = model.splurge {
                    VStack(alignment: .leading, spacing: 2) {
                        Eyebrow(text: "The splurge", coral: true)
                        Text(splurge).textStyle(Tokens.TypeRamp.bodyM).fixedSize(horizontal: false, vertical: true)
                    }
                }
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
            Image("ill-vending")
                .resizable()
                .scaledToFit()
                .frame(width: RecapLayout.vending)
                .padding(.top, -RecapLayout.vendingRise)
                .accessibilityHidden(true)
        }
    }

    private func progress(_ text: String) -> some View {
        Text(text)
            .textStyle(Tokens.TypeRamp.labelM)
            .monospacedDigit()
            .lineLimit(1)
            .fixedSize()
            .foregroundStyle(Tokens.Text.secondary)
    }

    private func helper(_ text: String) -> some View {
        Text(text)
            .textStyle(Tokens.TypeRamp.bodyS)
            .foregroundStyle(Tokens.Text.secondary)
            .fixedSize(horizontal: false, vertical: true)
    }
}

extension RecapViewModel {
    /// C3's "{kept%} of your saving target"; P3's, and C3's with no Saving Target, "Where {Spent} went".
    var goalHeadline: String {
        guard let recap = answer?.recap else { return "" }
        return !partial && recap.hasKeptPercent ? "\(recap.keptPercent)% of your saving target"
            : "Where \(money(recap.spent)) went"
    }

    /// C3's target card: "{Saved} of {target}" and the bar's drawn fill from those two figures; nil without a target.
    var target: (progress: String, fraction: Double)? {
        guard !partial, let recap = answer?.recap, recap.hasSavingTarget, recap.hasSaved else { return nil }
        let whole = Double(recap.savingTarget.amount.amount) ?? 0
        return (
            "\(money(recap.saved)) of \(money(recap.savingTarget.amount))",
            whole > 0 ? (Double(recap.saved.amount) ?? 0) / whole : 0
        )
    }

    /// P3 before Payday (D55): "Your salary lands on the 25th.", the day ledger's days until Payday lands on from
    /// today.
    var paydayLine: String? {
        guard partial, let days = answer?.recap.salary.daysUntilPayday, days > 0,
              let payday = Calendar.current.date(byAdding: .day, value: Int(days), to: today) else { return nil }
        let ordinal = NumberFormatter()
        ordinal.numberStyle = .ordinal
        ordinal.locale = Locale(identifier: "en")
        let day = Calendar.current.component(.day, from: payday) as NSNumber
        return "Your salary lands on the \(ordinal.string(from: day) ?? "\(day)")."
    }

    private var categories: [MoneyRecap_V1_CategoryShare] {
        answer?.recap.categories ?? []
    }

    var spendingEyebrow: String {
        guard let recap = answer?.recap, !categories.isEmpty else { return "Where it went" }
        return "Where \(money(recap.spent)) went"
    }

    /// "Mostly {top Category}" in lower case, or "Nothing spent this month".
    var spendingTitle: String {
        categories.first.map { "Mostly \(SpendListViewModel.glyph($0.category).label.lowercased())" }
            ?? "Nothing spent this month"
    }

    /// The three largest Categories (D71), each with ledger's percentage and its amount.
    var bars: [CategoryBar] {
        categories.prefix(3).map {
            CategoryBar(
                category: SpendListViewModel.glyph($0.category),
                percent: Int($0.percent),
                amount: money($0.amount)
            )
        }
    }

    /// "{names} · {n}%" for the Categories past the third, `n` the sum of ledger's rounded percentages (IOS §1).
    var rollUp: String? {
        let rest = categories.dropFirst(3)
        guard !rest.isEmpty else { return nil }
        let names = rest.map { SpendListViewModel.glyph($0.category).label }.joined(separator: ", ")
        return "\(names) · \(rest.reduce(0) { $0 + Int($1.percent) })%"
    }

    /// No Spend, and Money In above 0.
    var everyPenny: Bool {
        categories.isEmpty && (Decimal(string: answer?.recap.moneyIn.amount ?? "") ?? 0) > 0
    }

    /// "{amount} on {note, or the Category in lower case}. Your biggest single buy."; nil with no Spend.
    var splurge: String? {
        guard let recap = answer?.recap, recap.hasSplurge else { return nil }
        let spend = recap.splurge
        let what = spend.note.isEmpty ? SpendListViewModel.glyph(spend.category).label.lowercased() : spend.note
        return "\(money(spend.amount)) on \(what). Your biggest single buy."
    }
}
