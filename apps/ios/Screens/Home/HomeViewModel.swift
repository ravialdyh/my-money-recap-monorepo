import Connect
import Foundation
import MoneyRecapContract
import SwiftData

/// B1 Home (SCREEN-STATES, IOS I6): renders the stored snapshot at once, refreshes from the BFF, and
/// shows the §6.19 skeleton only while no snapshot exists. It holds no arithmetic: every figure is
/// the server's (ARCHITECTURE A2).
@Observable
@MainActor
final class HomeViewModel {
    private(set) var home: MoneyRecap_V1_HomeView?
    private let fetch: () async throws -> MoneyRecap_V1_HomeView
    private let context: ModelContext
    private let onboardingNeeded: () -> Void
    /// Dates print in the device locale (PRD §11); the tests pass a fixed one so the runner's locale cannot change the
    /// copy.
    private let dates: Date.FormatStyle

    init(
        fetch: @escaping () async throws -> MoneyRecap_V1_HomeView,
        context: ModelContext,
        locale: Locale = .autoupdatingCurrent,
        onboardingNeeded: @escaping () -> Void = {}
    ) {
        self.fetch = fetch
        self.context = context
        self.onboardingNeeded = onboardingNeeded
        dates = ISODate.style(locale: locale)
        if let snapshot = HomeSnapshot.load(from: context) {
            home = try? MoneyRecap_V1_HomeView(serializedBytes: snapshot.bytes)
        }
    }

    /// Nil `home` means no snapshot yet: the skeleton, no spinner (PRD §6.19). A failed refresh keeps the last one,
    /// offline or unauthenticated (D58, D61); failed_precondition, step 1 or 2 unanswered, deletes it (spec #92).
    func refresh() async {
        do {
            let fresh = try await fetch()
            try HomeSnapshot.replace(with: fresh, in: context)
            home = fresh
        } catch let error as ConnectError where error.code == .failedPrecondition {
            HomeSnapshot.erase(from: context)
            onboardingNeeded()
        } catch {
            return
        }
    }

    /// "Hi, {Display Name}": Home is only served once step 1 has the name (spec #92).
    var greeting: String? {
        home.map { "Hi, \($0.displayName)" }
    }

    /// "Spent so far", extended with the countdown before the effective Payday (PRD §6.8).
    var heroEyebrow: String {
        guard let home, home.hasSalary, home.salary.daysUntilPayday > 0 else { return "Spent so far" }
        return "Spent so far · salary lands in \(home.salary.daysUntilPayday) days"
    }

    /// The Saved chip exists only while Money In > 0 (PRD D41): the server omits `saved` then, and the phone checks the
    /// sign too.
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
        home.flatMap { ISODate.day($0.firstRecapOn) }
            .map { "Your first recap arrives on \($0.formatted(dates.day().month(.wide)))." }
    }

    func recapLabel(_ month: MoneyRecap_V1_RecapMonth) -> String {
        ISODate.month(month.month).map { $0.formatted(dates.month(.wide).year()) } ?? month.month
    }
}

/// The server's "YYYY-MM-DD" and "YYYY-MM" strings are calendar dates, not instants, so they are read and printed
/// in one fixed zone; otherwise a phone west of Greenwich would show 30 September for `2026-10-01` (IOS I7).
enum ISODate {
    static func style(locale: Locale) -> Date.FormatStyle {
        Date.FormatStyle(locale: locale, timeZone: .gmt)
    }

    static func day(_ string: String) -> Date? {
        try? Date(string, strategy: .iso8601.year().month().day())
    }

    static func month(_ string: String) -> Date? {
        day(string + "-01")
    }
}
