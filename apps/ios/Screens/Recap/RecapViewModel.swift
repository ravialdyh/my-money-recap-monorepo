import Foundation
import MoneyRecapContract

/// One month's Recap (SCREEN-STATES C1, P1; spec #151): which pages the situation draws and the words they print.
/// Every figure is the server's (A2); nil `answer` is the §6.19 loading state, kept when the BFF cannot answer.
@Observable
@MainActor
final class RecapViewModel: Identifiable {
    typealias Fetch = (String) async throws -> MoneyRecap_V1_RecapView

    /// The pager's pages in their fixed order: C1 for a full month, P1 for the current month and both partials.
    enum Page: Equatable {
        case cover, partialCover
    }

    /// "YYYY-MM" this Recap covers.
    let month: String
    private(set) var answer: MoneyRecap_V1_RecapView?
    /// The page on screen; a Recap always opens on its first.
    var page = 0
    private let fetch: Fetch
    private let dates: Date.FormatStyle
    private let locale: Locale

    init(month: String, fetch: @escaping Fetch, locale: Locale = .autoupdatingCurrent) {
        self.month = month
        self.fetch = fetch
        self.locale = locale
        dates = ISODate.style(locale: locale)
    }

    /// Asks the BFF once; a failure keeps whatever is drawn, the loading state included (§6.19).
    func load() async {
        if let fresh = try? await fetch(month) {
            answer = fresh
        }
    }

    /// No page before the answer; then the situation chooses (PRD §8.11).
    var pages: [Page] {
        guard answer != nil else { return [] }
        return partial ? [.partialCover] : [.cover]
    }

    /// Any situation but unspecified is the Partial Recap or the preview (PRD §8.11).
    private var partial: Bool {
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
        let name = date.map { $0.formatted(dates.month(.wide)) } ?? month
        switch answer?.recap.situation {
        case .currentMonth: return "\(name) so far"
        case .signUpMonth: return "Your first \(windowDays) days"
        case .noMoneyIn: return "Your \(name) spending"
        default: return "Your \(name) in money"
        }
    }

    /// The days from the Streak window's first day to its last, both counted.
    private var windowDays: Int {
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

    private var date: Date? {
        ISODate.month(month)
    }

    private func money(_ amount: MoneyRecap_V1_Money) -> String {
        Money.format(amount.amount, currency: answer?.currencyCode ?? amount.currencyCode, locale: locale)
    }
}
