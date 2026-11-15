import SwiftUI

struct RecapKeptPage: View {
    @Bindable var model: RecapViewModel
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
    static let jar: CGFloat = 190
    static let jarHang: CGFloat = 10
}

extension RecapViewModel {
    var monthName: String {
        date.map { $0.formatted(dates.month(.wide)) } ?? month
    }

    var keptHeadline: String {
        guard let recap = answer?.recap, !partial else { return "Your longest no-spend streak" }
        let kept = "You kept \(money(recap.saved)) of \(money(recap.moneyIn))."
        return recap.hasSavingTarget ? "\(kept) That's \(recap.keptPercent)% of your target." : kept
    }

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

    var streakTitle: String? {
        partial ? runDates : "Longest no-spend streak"
    }

    var streakLine: String {
        let badge = answer?.recap.streak.badgeUnlocked == true ? "New badge unlocked." : "Beat it next month."
        guard !partial, let dates = runDates else { return badge }
        return "\(dates). \(badge)"
    }

    var runDates: String? {
        guard let streak = answer?.recap.streak, let start = ISODate.day(streak.runStart),
              let end = ISODate.day(streak.runEnd) else { return nil }
        let last = end.formatted(dates.day().month(.wide))
        guard start != end, let day = last.range(of: end.formatted(dates.day())) else { return last }
        return last.replacingCharacters(in: day, with: "\(start.formatted(dates.day())) to \(last[day])")
    }

    var runDays: ClosedRange<Int>? {
        guard let streak = answer?.recap.streak, let window = ISODate.day(streak.windowStart),
              let start = ISODate.day(streak.runStart), let end = ISODate.day(streak.runEnd) else { return nil }
        let place = { (day: Date) in Int((day.timeIntervalSince(window) / 86400).rounded()) + 1 }
        return place(start) ... place(end)
    }
}
