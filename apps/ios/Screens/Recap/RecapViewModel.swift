import Foundation
import MoneyRecapContract
import SwiftData

@Observable
@MainActor
final class RecapViewModel: Identifiable {
    typealias Fetch = (String) async throws -> MoneyRecap_V1_RecapView

    enum Page: Equatable {
        case cover, partialCover, kept, streak, goal, partialGoal, archetype, closing
    }

    let month: String
    private(set) var answer: MoneyRecap_V1_RecapView?
    var page = 0
    var shared: SharedPicture?
    var targetEdit: TargetEditViewModel?
    private let fetch: Fetch
    private let openTarget: () async -> TargetEditViewModel?
    private let context: ModelContext?
    let dates: Date.FormatStyle
    private let locale: Locale
    let today: Date

    init(
        month: String,
        fetch: @escaping Fetch,
        context: ModelContext? = nil,
        locale: Locale = .autoupdatingCurrent,
        today: Date = .now,
        openTarget: @escaping () async -> TargetEditViewModel? = { nil }
    ) {
        self.month = month
        self.fetch = fetch
        self.openTarget = openTarget
        self.context = context
        self.locale = locale
        self.today = today
        dates = ISODate.style(locale: locale)
        answer = context.flatMap { RecapSnapshot.load(month, from: $0) }
    }

    func load() async {
        guard let fresh = try? await fetch(month), fresh.recap.month == month else { return }
        if let context {
            try? RecapSnapshot.replace(with: fresh, in: context)
        }
        answer = fresh
    }

    func editTarget() async {
        targetEdit = await openTarget()
    }

    var pages: [Page] {
        guard answer != nil else { return [] }
        return partial ? [.partialCover, .streak, .partialGoal, .closing] : [.cover, .kept, .goal, .archetype]
    }

    var partial: Bool {
        answer.map { $0.recap.situation != .unspecified } ?? false
    }

    var monthChip: String {
        date.map { $0.formatted(dates.month(.wide).year()) } ?? month
    }

    var nameChip: String {
        "\(answer?.displayName ?? "")'s recap"
    }

    var headline: String {
        let name = monthName
        switch answer?.recap.situation {
        case .currentMonth: return "\(name) so far"
        case .signUpMonth: return "Your first \(windowDays) days"
        case .noMoneyIn: return "Your \(name) spending"
        default: return "Your \(name) in money"
        }
    }

    var windowDays: Int {
        guard let streak = answer?.recap.streak, let start = ISODate.day(streak.windowStart),
              let end = ISODate.day(streak.windowEnd) else { return 0 }
        return Int((end.timeIntervalSince(start) / 86400).rounded()) + 1
    }

    var eyebrow: String {
        partial ? "What you've spent" : "Money that came your way"
    }

    var hero: String {
        guard let recap = answer?.recap else { return " " }
        return money(partial ? recap.spent : recap.moneyIn)
    }

    var chips: [(tone: StatChip.Tone, text: String)] {
        guard let recap = answer?.recap else { return [(.spent, " ")] }
        if partial {
            let count = recap.spendCount
            return [(.spent, count == 0 ? "No spends" : count == 1 ? "1 spend" : "\(count) spends")]
        }
        let saved: [(tone: StatChip.Tone, text: String)] = recap
            .hasSaved ? [(.saved, "Saved \(money(recap.saved))")] : []
        return saved + [(.spent, "Spent \(money(recap.spent))")]
    }

    var date: Date? {
        ISODate.month(month)
    }

    func money(_ amount: MoneyRecap_V1_Money) -> String {
        Money.format(amount.amount, currency: answer?.currencyCode ?? amount.currencyCode, locale: locale)
    }
}
