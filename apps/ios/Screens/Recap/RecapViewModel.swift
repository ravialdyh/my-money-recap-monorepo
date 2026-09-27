import Foundation
import MoneyRecapContract
import SwiftData

/// One month's Recap (SCREEN-STATES C and P, spec #151): its situation's pages and page 1's words, later pages' words
/// beside their views. Figures are the server's (A2); a stored copy draws at once (IOS I4); nil `answer` is the §6.19
/// loading state.
@Observable
@MainActor
final class RecapViewModel: Identifiable {
    typealias Fetch = (String) async throws -> MoneyRecap_V1_RecapView

    /// The pager's pages in their fixed order: C1 to C4 for a full month; P1 to P4 for the month so far and both
    /// partial situations.
    enum Page: Equatable {
        case cover, partialCover, kept, streak, goal, partialGoal, archetype, closing
    }

    /// "YYYY-MM" this Recap covers.
    let month: String
    private(set) var answer: MoneyRecap_V1_RecapView?
    /// The page on screen; a Recap always opens on its first.
    var page = 0
    /// The poster Share drew, presented in the system share sheet until it closes.
    var shared: SharedPicture?
    private let fetch: Fetch
    private let context: ModelContext?
    let dates: Date.FormatStyle
    private let locale: Locale
    /// The Member's today, from which P3 counts ledger's days until Payday (D55).
    let today: Date

    /// `context` holds the stored Recaps; without one nothing is read or kept.
    init(
        month: String,
        fetch: @escaping Fetch,
        context: ModelContext? = nil,
        locale: Locale = .autoupdatingCurrent,
        today: Date = .now
    ) {
        self.month = month
        self.fetch = fetch
        self.context = context
        self.locale = locale
        self.today = today
        dates = ISODate.style(locale: locale)
        answer = context.flatMap { RecapSnapshot.load(month, from: $0) }
    }

    /// Asks the BFF once; the answer replaces the stored copy and what is drawn. A failure keeps whatever is drawn, the
    /// loading state included (§6.19).
    func load() async {
        guard let fresh = try? await fetch(month), fresh.recap.month == month else { return }
        if let context {
            try? RecapSnapshot.replace(with: fresh, in: context)
        }
        answer = fresh
    }

    /// No page before the answer; then the situation chooses (PRD §8.11).
    var pages: [Page] {
        guard answer != nil else { return [] }
        return partial ? [.partialCover, .streak, .partialGoal, .closing] : [.cover, .kept, .goal, .archetype]
    }

    /// Any situation but unspecified is the Partial Recap or the preview (PRD §8.11).
    var partial: Bool {
        answer.map { $0.recap.situation != .unspecified } ?? false
    }

    /// "September 2026".
    var monthChip: String {
        date.map { $0.formatted(dates.month(.wide).year()) } ?? month
    }

    /// "{Display Name}'s recap": the current name, so a new one reaches old Recaps (PRD §7).
    var nameChip: String {
        "\(answer?.displayName ?? "")'s recap"
    }

    /// C1's "Your {Month} in money", or P1's by situation: "{Month} so far", "Your first {n} days" or "Your {Month}
    /// spending".
    var headline: String {
        let name = monthName
        switch answer?.recap.situation {
        case .currentMonth: return "\(name) so far"
        case .signUpMonth: return "Your first \(windowDays) days"
        case .noMoneyIn: return "Your \(name) spending"
        default: return "Your \(name) in money"
        }
    }

    /// The days from the Streak window's first day to its last, both counted.
    var windowDays: Int {
        guard let streak = answer?.recap.streak, let start = ISODate.day(streak.windowStart),
              let end = ISODate.day(streak.windowEnd) else { return 0 }
        return Int((end.timeIntervalSince(start) / 86400).rounded()) + 1
    }

    var eyebrow: String {
        partial ? "What you've spent" : "Money that came your way"
    }

    /// Money In on C1; Spent on P1, which shows Money In nowhere.
    var hero: String {
        guard let recap = answer?.recap else { return " " }
        return money(partial ? recap.spent : recap.moneyIn)
    }

    /// C1's Saved (a negative one printed as it is) and Spent; P1's "{n} spends", "1 spend" or "No spends". Before the
    /// answer, one blank chip holds the card at its height.
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
