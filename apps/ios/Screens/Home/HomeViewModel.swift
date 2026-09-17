import Foundation
import MoneyRecapContract
import SwiftData

/// B1 Home (SCREEN-STATES), IOS I6: renders from the stored snapshot at once, refreshes from the BFF, keeps the
/// last snapshot when the refresh fails, and shows the §6.19 skeleton only when no snapshot exists yet.
/// It holds no arithmetic: every figure is the server's (ARCHITECTURE A2).
@Observable
@MainActor
final class HomeViewModel {
    private(set) var home: MoneyRecap_V1_HomeView?
    private(set) var updatedAt: Date?
    private(set) var isRefreshing = false
    private let fetch: () async throws -> MoneyRecap_V1_HomeView
    private let context: ModelContext
    /// Dates print in the device locale (PRD §11); the tests pass a fixed one so the runner's locale cannot change the copy.
    private let dates: Date.FormatStyle

    init(fetch: @escaping () async throws -> MoneyRecap_V1_HomeView, context: ModelContext, locale: Locale = .autoupdatingCurrent) {
        self.fetch = fetch
        self.context = context
        self.dates = ISODate.style(locale: locale)
        if let snapshot = HomeSnapshot.load(from: context) {
            home = try? MoneyRecap_V1_HomeView(serializedBytes: snapshot.bytes)
            updatedAt = snapshot.updatedAt
        }
    }

    /// No snapshot yet: the two cards at their filled heights with ink blocks, no spinner (PRD §6.19).
    var isFirstLoad: Bool { home == nil }

    func refresh() async {
        isRefreshing = true
        defer { isRefreshing = false }
        do {
            let fresh = try await fetch()
            try HomeSnapshot.replace(with: fresh, in: context)
            home = fresh
            updatedAt = .now
        } catch {
            // Offline or unauthenticated: the last snapshot stays on screen (PRD D58, D61).
        }
    }

    // MARK: - B1 display rules

    /// "Hi, {Display Name}", omitted while the name is empty (before onboarding).
    var greeting: String? {
        guard let home, !home.displayName.isEmpty else { return nil }
        return "Hi, \(home.displayName)"
    }

    /// "Spent so far", extended with the countdown before the effective Payday (PRD §6.8).
    var heroEyebrow: String {
        guard let home, home.hasSalary, home.salary.daysUntilPayday > 0 else { return "Spent so far" }
        return "Spent so far · salary lands in \(home.salary.daysUntilPayday) days"
    }

    /// The Saved chip exists only while Money In > 0 (PRD D41): the server omits `saved` then, and the phone checks the sign too.
    var showsSavedChip: Bool {
        guard let home, home.hasSaved else { return false }
        return (Decimal(string: home.moneyIn.amount) ?? 0) > 0
    }

    /// The month chip: the Recap Month's name.
    var monthLabel: String? {
        home.flatMap { ISODate.month($0.month) }.map { $0.formatted(dates.month(.wide)) }
    }

    /// "Your first recap arrives on 1 October." (PRD D47).
    var previewCopy: String? {
        home.flatMap { ISODate.day($0.firstRecapOn) }.map { "Your first recap arrives on \($0.formatted(dates.day().month(.wide)))." }
    }

    func recapLabel(_ month: MoneyRecap_V1_RecapMonth) -> String {
        ISODate.month(month.month).map { $0.formatted(dates.month(.wide).year()) } ?? month.month
    }
}

/// The server's "YYYY-MM-DD" and "YYYY-MM" strings are calendar dates, not instants, so they are read and printed
/// in one fixed zone; otherwise a phone west of Greenwich would show 30 September for `2026-10-01` (IOS I7).
enum ISODate {
    static func style(locale: Locale) -> Date.FormatStyle { Date.FormatStyle(locale: locale, timeZone: .gmt) }

    static func day(_ string: String) -> Date? {
        try? Date(string, strategy: .iso8601.year().month().day())
    }

    static func month(_ string: String) -> Date? {
        day(string + "-01")
    }
}
