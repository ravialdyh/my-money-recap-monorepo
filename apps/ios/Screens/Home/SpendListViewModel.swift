import Foundation
import MoneyRecapContract
import SwiftData

/// B6 Spend list and B6b Category detail (PRD §6.10, D59): one month's stored list at once, then the BFF's, each answer
/// replacing that month's snapshot (IOS I4). Every figure and share is the server's (A2); this only words them.
@Observable
@MainActor
final class SpendListViewModel: Identifiable {
    typealias Fetch = (String) async throws -> MoneyRecap_V1_SpendListView

    /// The month shown, "YYYY-MM".
    private(set) var month: String
    /// Nil until a snapshot or an answer for `month` exists.
    private(set) var list: MoneyRecap_V1_SpendListView?
    /// The month switcher's list is open.
    var switching: Bool
    /// The Category whose detail is open.
    var category: Icon.Name?
    private let fetch: Fetch
    private let context: ModelContext
    private let dates: Date.FormatStyle
    private let locale: Locale
    private let connection: Connection

    init(
        month: String,
        switching: Bool = false,
        fetch: @escaping Fetch,
        context: ModelContext,
        locale: Locale = .autoupdatingCurrent,
        connection: Connection? = nil
    ) {
        self.month = month
        self.switching = switching
        self.fetch = fetch
        self.context = context
        self.locale = locale
        self.connection = connection ?? .shared
        dates = ISODate.style(locale: locale)
        list = SpendListSnapshot.load(month, from: context)
    }

    /// Asks for `month`; the answer replaces its snapshot, and the list while that month is shown. A failure keeps what
    /// is shown, offline included.
    func refresh() async {
        await refresh(month)
    }

    /// After a write, asks for the month shown and every month the write changed (`LogSpendViewModel.months`), so each
    /// stored snapshot is the server's (IOS I4). A Category left with no Spend closes its detail.
    func refresh(_ months: Set<String>) async {
        for asked in months.union([month]).sorted() {
            await refresh(asked)
        }
        if let category, share(category) == nil {
            self.category = nil
        }
    }

    private func refresh(_ asked: String) async {
        guard let fresh = try? await fetch(asked), fresh.month == asked else { return }
        try? SpendListSnapshot.replace(with: fresh, in: context)
        if month == asked {
            list = fresh
        }
    }

    /// No network path: the empty month's plus button is disabled at 40%, as Home's Log a spend is (D58).
    var offline: Bool {
        !connection.online
    }

    /// The switcher's pick: that month's snapshot at once, then its answer.
    func pick(_ picked: String) async {
        switching = false
        guard picked != month else { return }
        month = picked
        list = SpendListSnapshot.load(picked, from: context)
        await refresh()
    }

    /// The switcher's months, newest first; the month shown alone until an answer names them.
    var months: [String] {
        list.map(\.months).flatMap { $0.isEmpty ? nil : $0 } ?? [month]
    }

    /// "September 2026", the switcher's chip and rows and the Category detail's label.
    func label(_ month: String) -> String {
        ISODate.month(month).map { $0.formatted(dates.month(.wide).year()) } ?? month
    }

    /// The current month is the newest the switcher lists (D59).
    private var isCurrent: Bool {
        list?.months.first == month
    }

    private var monthName: String {
        ISODate.month(month).map { $0.formatted(dates.month(.wide)) } ?? month
    }

    /// B6 empty: the Streak for this month, the month's name for a past one (owner ruling 2026-09-24); nil with Spends.
    var emptyLine: String? {
        guard let list, list.categories.isEmpty else { return nil }
        return isCurrent ? "Nothing spent yet. Streak day \(list.streakDay)." : "Nothing spent in \(monthName)."
    }

    /// The month line under the switcher: "{spent} so far, {n} spends" ("1 spend", "no spends"); a past month drops
    /// "so far" (owner ruling 2026-09-24).
    var monthLine: String? {
        guard let list else { return nil }
        let count = list.spends.count
        let spends = count == 0 ? "no spends" : count == 1 ? "1 spend" : "\(count) spends"
        return "\(money(list.spent))\(isCurrent ? " so far" : ""), \(spends)"
    }

    func money(_ amount: MoneyRecap_V1_Money) -> String {
        Money.format(amount.amount, currency: list?.currencyCode ?? amount.currencyCode, locale: locale)
    }

    /// A Category row's sub-line: "1 spend", "{n} spends".
    func count(_ share: MoneyRecap_V1_CategoryShare) -> String {
        share.spendCount == 1 ? "1 spend" : "\(share.spendCount) spends"
    }

    /// A Category's row, as ledger answered it.
    func share(_ category: Icon.Name) -> MoneyRecap_V1_CategoryShare? {
        list?.categories.first { Self.glyph($0.category) == category }
    }

    /// B6b's sub-line: "{amount} · {n}% of {Month} so far", without "so far" for a past month.
    func shareLine(_ category: Icon.Name) -> String? {
        share(category).map { "\(money($0.amount)) · \($0.percent)% of \(monthName)\(isCurrent ? " so far" : "")" }
    }

    /// A Category's Spends, newest first as the list holds them.
    func spends(_ category: Icon.Name) -> [MoneyRecap_V1_Spend] {
        list?.spends.filter { Self.glyph($0.category) == category } ?? []
    }

    /// A Spend row's label: its note, or its Category's name when it has none (PRD §6.10).
    func title(_ spend: MoneyRecap_V1_Spend) -> String {
        spend.note.isEmpty ? Self.glyph(spend.category).label : spend.note
    }

    /// A Spend row's secondary line, its date: "24 September".
    func day(_ spend: MoneyRecap_V1_Spend) -> String {
        ISODate.day(spend.date).map { $0.formatted(dates.day().month(.wide)) } ?? spend.date
    }

    /// The glyph of a contract Category: both follow PRD D20's order.
    static func glyph(_ category: MoneyRecap_V1_Category) -> Icon.Name {
        let all = Icon.Name.allCases
        return all.indices.contains(category.rawValue - 1) ? all[category.rawValue - 1] : .other
    }
}
