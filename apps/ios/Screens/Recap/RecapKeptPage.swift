import SwiftUI

/// Page 2, C2 or P2 (PRD §6.12, §6.17): on the yellow ground, the chips, the headline, the jar, C2's bars card or its
/// target card, and the streak card with the run lit. Scrolls held sideways.
struct RecapKeptPage: View {
    @Bindable var model: RecapViewModel
    /// The page is on screen: the bars reveal (`Reveal`).
    var active = true

    var body: some View {
        Screen(ground: Tokens.Ground.earning, scrolls: true) {
            StatChipRow {
                Chip(text: model.partial ? "Streak" : "Kept", asLabel: true, inverse: true)
                Chip(text: model.monthName, asLabel: true)
            }
            .padding(.leading, RecapLayout.chipShift)
            .padding(.top, Tokens.Space._3)
            Headline(text: model.keptHeadline, size: .compact)
                .padding(.top, 10)
            Illustration(name: "ill-jar", width: RecapLayout.jar, hang: RecapLayout.jarHang, grows: false)
            if !model.partial {
                bars.padding(.bottom, Tokens.Space._4)
            }
            streak
            PageDots(count: model.pages.count, active: $model.page, onLight: true)
        }
    }

    @ViewBuilder private var bars: some View {
        if let rows = model.keptRows {
            Card {
                Eyebrow(text: "Kept against your target")
                CompareBars(rows: rows, scale: .max, shown: active)
            }
        } else {
            PromptCard(
                eyebrow: "Saving target",
                title: "Set a target to see how you did",
                helper: "Pick a share of your salary to keep each month, and next month's recap will score it.",
                action: { Task { await model.editTarget() } }
            )
        }
    }

    private var streak: some View {
        Card(accent: true) {
            HStack(spacing: Tokens.Space._5) {
                StreakBadge(value: model.streakDays, unit: model.streakDays == 1 ? "day" : "days")
                VStack(alignment: .leading, spacing: 2) {
                    if let title = model.streakTitle {
                        CardTitle(text: title)
                    }
                    Text(model.streakLine)
                        .textStyle(Tokens.TypeRamp.labelS)
                        .opacity(0.8)
                        .fixedSize(horizontal: false, vertical: true)
                }
            }
            StreakDots(total: model.windowDays, active: model.runDays)
        }
    }
}

extension RecapLayout {
    /// The jar's width on screen 2 and how far it hangs into the card below (`jarStyle` in `lib.mjs`).
    static let jar: CGFloat = 190
    static let jarHang: CGFloat = 10
}

/// Page 2's words (spec #151 copy table). Every figure is ledger's; the phone prints the run's dates, never finds them.
extension RecapViewModel {
    /// "September", the second chip.
    var monthName: String {
        date.map { $0.formatted(dates.month(.wide)) } ?? month
    }

    /// C2's "You kept {Saved} of {Money In}. That's {kept%} of your target.", its first sentence alone with no Saving
    /// Target and a negative Saved printed as it is; P2's "Your longest no-spend streak".
    var keptHeadline: String {
        guard let recap = answer?.recap, !partial else { return "Your longest no-spend streak" }
        let kept = "You kept \(money(recap.saved)) of \(money(recap.moneyIn))."
        return recap.hasSavingTarget ? "\(kept) That's \(recap.keptPercent)% of your target." : kept
    }

    /// C2's Kept against the Target, scaled to the larger; nil with no Saving Target, where the target card stands, and
    /// on P2, which has no bars card.
    var keptRows: [CompareBars.Row]? {
        guard let recap = answer?.recap, recap.hasSavingTarget, !partial else { return nil }
        let target = recap.savingTarget.amount
        return [
            .init(label: "Kept", value: Double(recap.saved.amount) ?? 0, display: money(recap.saved), tone: .green),
            .init(label: "Target", value: Double(target.amount) ?? 0, display: money(target), tone: .violet),
        ]
    }

    var streakDays: Int {
        Int(answer?.recap.streak.days ?? 0)
    }

    /// C2's "Longest no-spend streak"; P2's run dates in its place, nil with no run.
    var streakTitle: String? {
        partial ? runDates : "Longest no-spend streak"
    }

    /// The line under the title: on C2 "{start} to {end}. {badge copy}", on P2 the badge copy alone; with no run the
    /// dates are left out.
    var streakLine: String {
        let badge = answer?.recap.streak.badgeUnlocked == true ? "New badge unlocked." : "Beat it next month."
        guard !partial, let dates = runDates else { return badge }
        return "\(dates). \(badge)"
    }

    /// The run in the locale's day and month order, "12 to 20 September" or "September 12 to 20" (a run never leaves
    /// its month), one date for a one-day run; nil when the Streak is 0 and ledger sends no run.
    var runDates: String? {
        guard let streak = answer?.recap.streak, let start = ISODate.day(streak.runStart),
              let end = ISODate.day(streak.runEnd) else { return nil }
        let last = end.formatted(dates.day().month(.wide))
        guard start != end, let day = last.range(of: end.formatted(dates.day())) else { return last }
        return last.replacingCharacters(in: day, with: "\(start.formatted(dates.day())) to \(last[day])")
    }

    /// The run's days as 1-based places in the Streak window, which the dots light: date arithmetic on ledger's dates.
    var runDays: ClosedRange<Int>? {
        guard let streak = answer?.recap.streak, let window = ISODate.day(streak.windowStart),
              let start = ISODate.day(streak.runStart), let end = ISODate.day(streak.runEnd) else { return nil }
        let place = { (day: Date) in Int((day.timeIntervalSince(window) / 86400).rounded()) + 1 }
        return place(start) ... place(end)
    }
}
