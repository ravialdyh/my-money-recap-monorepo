import Foundation
import MoneyRecapContract
import SwiftData

@Observable
@MainActor
final class SpendListViewModel: Identifiable {
    typealias Fetch = (String) async throws -> MoneyRecap_V1_SpendListView

    private(set) var month: String
    private(set) var list: MoneyRecap_V1_SpendListView?
    var switching: Bool
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

    func refresh() async {
        await refresh(month)
    }

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

    var offline: Bool {
        !connection.online
    }

    func pick(_ picked: String) async {
        switching = false
        guard picked != month else { return }
        month = picked
        list = SpendListSnapshot.load(picked, from: context)
        await refresh()
    }

    var months: [String] {
        list.map(\.months).flatMap { $0.isEmpty ? nil : $0 } ?? [month]
    }

    func label(_ month: String) -> String {
        ISODate.month(month).map { $0.formatted(dates.month(.wide).year()) } ?? month
    }

    private var isCurrent: Bool {
        list?.months.first == month
    }

    private var monthName: String {
        ISODate.month(month).map { $0.formatted(dates.month(.wide)) } ?? month
    }

    var emptyLine: String? {
        guard let list, list.categories.isEmpty else { return nil }
        return isCurrent ? "Nothing spent yet. Streak day \(list.streakDay)." : "Nothing spent in \(monthName)."
    }

    var monthLine: String? {
        guard let list else { return nil }
        let count = list.spends.count
        let spends = count == 0 ? "no spends" : count == 1 ? "1 spend" : "\(count) spends"
        return "\(money(list.spent))\(isCurrent ? " so far" : ""), \(spends)"
    }

    func money(_ amount: MoneyRecap_V1_Money) -> String {
        Money.format(amount.amount, currency: list?.currencyCode ?? amount.currencyCode, locale: locale)
    }

    func count(_ share: MoneyRecap_V1_CategoryShare) -> String {
        share.spendCount == 1 ? "1 spend" : "\(share.spendCount) spends"
    }

    func share(_ category: Icon.Name) -> MoneyRecap_V1_CategoryShare? {
        list?.categories.first { Self.glyph($0.category) == category }
    }

    func shareLine(_ category: Icon.Name) -> String? {
        share(category).map { "\(money($0.amount)) · \($0.percent)% of \(monthName)\(isCurrent ? " so far" : "")" }
    }

    func spends(_ category: Icon.Name) -> [MoneyRecap_V1_Spend] {
        list?.spends.filter { Self.glyph($0.category) == category } ?? []
    }

    func title(_ spend: MoneyRecap_V1_Spend) -> String {
        spend.note.isEmpty ? Self.glyph(spend.category).label : spend.note
    }

    func day(_ spend: MoneyRecap_V1_Spend) -> String {
        ISODate.day(spend.date).map { $0.formatted(dates.day().month(.wide)) } ?? spend.date
    }

    static func glyph(_ category: MoneyRecap_V1_Category) -> Icon.Name {
        let all = Icon.Name.allCases
        return all.indices.contains(category.rawValue - 1) ? all[category.rawValue - 1] : .other
    }
}
