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
    private let logSpend: LogSpendViewModel.Send
    private let editSpend: LogSpendViewModel.Edit
    private let deleteSpend: LogSpendViewModel.Delete
    private let listSpends: SpendListViewModel.Fetch
    private let context: ModelContext
    private let onboardingNeeded: () -> Void
    /// Dates and money print in the device locale (PRD §11); the tests pass a fixed one so the runner's locale cannot
    /// change the copy.
    private let dates: Date.FormatStyle
    private let locale: Locale
    private let dismissals: WarningDismissals
    private let connection: Connection

    init(
        fetch: @escaping () async throws -> MoneyRecap_V1_HomeView,
        context: ModelContext,
        locale: Locale = .autoupdatingCurrent,
        dismissals: WarningDismissals? = nil,
        connection: Connection? = nil,
        onboardingNeeded: @escaping () -> Void = {},
        logSpend: @escaping LogSpendViewModel
            .Send = { _, _ in .failure(ConnectError(code: .unavailable, message: nil)) },
        listSpends: @escaping SpendListViewModel.Fetch = { _ in throw ConnectError(code: .unavailable, message: nil) },
        editSpend: @escaping LogSpendViewModel
            .Edit = { _, _ in .failure(ConnectError(code: .unavailable, message: nil)) },
        deleteSpend: @escaping LogSpendViewModel
            .Delete = { _, _ in .failure(ConnectError(code: .unavailable, message: nil)) }
    ) {
        self.fetch = fetch
        self.logSpend = logSpend
        self.editSpend = editSpend
        self.deleteSpend = deleteSpend
        self.listSpends = listSpends
        self.context = context
        self.onboardingNeeded = onboardingNeeded
        dates = ISODate.style(locale: locale)
        self.locale = locale
        self.dismissals = dismissals ?? .session
        self.connection = connection ?? .shared
        if let snapshot = HomeSnapshot.load(from: context) {
            home = try? MoneyRecap_V1_HomeView(serializedBytes: snapshot.bytes)
        }
    }

    /// Nil `home` means no snapshot yet: the skeleton, no spinner (PRD §6.19). A failed refresh keeps the last one,
    /// offline or unauthenticated (D58, D61); failed_precondition, step 1 or 2 unanswered, deletes it (spec #92).
    func refresh() async {
        do {
            try await replace(with: fetch())
        } catch let error as ConnectError where error.code == .failedPrecondition {
            HomeSnapshot.erase(from: context)
            onboardingNeeded()
        } catch {
            return
        }
    }

    /// The Log a spend sheet in the Member's Currency, its decimals and ledger's date range, or nil before Home has an
    /// answer.
    func logSheet() -> LogSpendViewModel? {
        home.map {
            LogSpendViewModel(
                currency: $0.currencyCode, digits: Int($0.currencyDigits), earliest: $0.earliestSpendDate,
                send: logSpend
            )
        }
    }

    /// B7 Edit spend for one of the list's Spends, in the Member's Currency, its decimals and ledger's date range; nil
    /// before Home has an answer.
    func editSheet(_ spend: MoneyRecap_V1_Spend) -> LogSpendViewModel? {
        home.map {
            LogSpendViewModel(
                editing: spend, currency: $0.currencyCode, digits: Int($0.currencyDigits),
                earliest: $0.earliestSpendDate, locale: locale, edit: editSpend, delete: deleteSpend
            )
        }
    }

    /// The Spend list of Home's month (PRD §6.10), its switcher open when the month chip opened it (D59); nil before
    /// Home has an answer.
    func spendList(switching: Bool) -> SpendListViewModel? {
        home.map { SpendListViewModel(
            month: $0.month,
            switching: switching,
            fetch: listSpends,
            context: context,
            locale: locale
        ) }
    }

    /// A write's answer replaces the stored Home whole, as a refresh does (IOS I4).
    func replace(with fresh: MoneyRecap_V1_HomeView) {
        try? HomeSnapshot.replace(with: fresh, in: context)
        home = fresh
    }

    /// No network path: the grey Banner at the top and Log a spend at 40% (D58), over the stored numbers.
    var offline: Bool {
        !connection.online
    }

    /// "Hi, {Display Name}": Home is only served once step 1 has the name (spec #92).
    var greeting: String? {
        home.map { "Hi, \($0.displayName)" }
    }

    /// "Spent so far", extended with the countdown before the effective Payday (PRD §6.8): "in 1 day", "in 15 days".
    var heroEyebrow: String {
        guard let home, home.hasSalary, home.salary.daysUntilPayday > 0 else { return "Spent so far" }
        let days = home.salary.daysUntilPayday
        return "Spent so far · salary lands in \(days) \(days == 1 ? "day" : "days")"
    }

    /// An amount in the Member's Currency, printed in the locale Home was given (`Money`).
    func money(_ amount: MoneyRecap_V1_Money) -> String {
        Money.format(amount.amount, currency: home?.currencyCode ?? amount.currencyCode, locale: locale)
    }

    /// The Saved chip exists only while Money In > 0 (PRD D41): the server omits `saved` then, and the phone checks the
    /// sign too.
    var showsSavedChip: Bool {
        guard let home, home.hasSaved else { return false }
        return (Decimal(string: home.moneyIn.amount) ?? 0) > 0
    }

    /// "Rp0 of Rp3.000.000", Saved against the Target's amount (PRD §6.8); Saved reads 0 while the server omits it
    /// (D41).
    var targetProgress: String? {
        guard let home, home.hasSavingTarget else { return nil }
        let saved = Money.format(home.hasSaved ? home.saved.amount : "0", currency: home.currencyCode, locale: locale)
        return "\(saved) of \(money(home.savingTarget.amount))"
    }

    /// How far the target bar is drawn: Saved over the Target's amount, clamped to the track. Drawing, not a figure.
    var targetFraction: Double {
        guard let home, home.hasSavingTarget, home.hasSaved,
              let saved = Double(home.saved.amount), let whole = Double(home.savingTarget.amount.amount), whole > 0
        else { return 0 }
        return min(max(saved / whole, 0), 1)
    }

    /// The Target Warning's kept copy (PRD §8.8), until the Member dismisses this month's for the session (§6.8).
    var warning: String? {
        guard let home, home.hasTargetWarning, !dismissals.months.contains(home.month) else { return nil }
        return home.targetWarning.message
    }

    func dismissWarning() {
        if let month = home?.month {
            dismissals.months.insert(month)
        }
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

/// The months whose Target Warning the Member dismissed: one set per run of the app, so the banner is back on the next
/// launch while the month is stamped (PRD §6.8).
@Observable
@MainActor
final class WarningDismissals {
    static let session = WarningDismissals()
    var months: Set<String> = []
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
