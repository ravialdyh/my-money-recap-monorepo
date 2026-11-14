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
    private let screens: HomeScreens
    private let onboardingNeeded: () -> Void
    private let dates: Date.FormatStyle
    private let dismissals: WarningDismissals
    /// A Spend was logged through the sheet still closing, so B5 may ask once it has closed.
    private var loggedInSheet = false

    init(
        fetch: @escaping () async throws -> MoneyRecap_V1_HomeView,
        screens: HomeScreens,
        dismissals: WarningDismissals? = nil,
        onboardingNeeded: @escaping () -> Void = {}
    ) {
        self.fetch = fetch
        self.screens = screens
        self.onboardingNeeded = onboardingNeeded
        dates = ISODate.style(locale: screens.locale)
        self.dismissals = dismissals ?? .session
        if let snapshot = HomeSnapshot.load(from: screens.context) {
            home = try? MoneyRecap_V1_HomeView(serializedBytes: snapshot.bytes)
        }
    }

    /// Nil `home` means no snapshot yet: the skeleton, no spinner (PRD §6.19). A failed refresh keeps the last one,
    /// offline or unauthenticated (D58, D61); failed_precondition, step 1 or 2 unanswered, deletes it (spec #92).
    func refresh() async {
        do {
            try await replace(with: fetch())
        } catch let error as ConnectError where error.code == .failedPrecondition {
            HomeSnapshot.erase(from: screens.context)
            onboardingNeeded()
        } catch {
            return
        }
    }

    /// The Log a spend sheet over Home's Currency, or nil before Home has an answer.
    func logSheet() -> LogSpendViewModel? {
        home.map { screens.logSpend(for: $0) }
    }

    /// B7 Edit spend for one of the list's Spends; nil before Home has an answer.
    func editSheet(_ spend: MoneyRecap_V1_Spend) -> LogSpendViewModel? {
        home.map { screens.editSpend(spend, for: $0) }
    }

    /// The Spend list of Home's month, its switcher open when the month chip opened it; nil before Home has an answer.
    func spendList(switching: Bool) -> SpendListViewModel? {
        home.map { screens.spendList(for: $0, switching: switching) }
    }

    /// D1 Settings; a saved Target's Home, and a changed Currency's, replaces this one.
    func settings() -> SettingsViewModel {
        screens.settings(onHome: { [weak self] in self?.replace(with: $0) })
    }

    /// A write's answer replaces the stored Home whole, as a refresh does (IOS I4).
    func replace(with fresh: MoneyRecap_V1_HomeView) {
        try? HomeSnapshot.replace(with: fresh, in: screens.context)
        home = fresh
    }

    /// The Log a spend sheet's answer: its Home, and B5 once the sheet has closed.
    func logged(_ fresh: MoneyRecap_V1_HomeView) {
        replace(with: fresh)
        loggedInSheet = true
    }

    /// The Log a spend sheet closed: after a save, the push permission alert when iOS never asked (B5).
    func logSheetClosed() async {
        guard loggedInSheet else { return }
        loggedInSheet = false
        await screens.push?.askOnce()
    }

    /// Each launch registers an allowed phone for pushes again (IOS I8).
    func launched() async {
        await screens.push?.launched()
    }

    /// No network path: the grey Banner at the top and Log a spend at 40% (D58), over the stored numbers.
    var offline: Bool {
        !screens.connection.online
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
        Money.format(amount.amount, currency: home?.currencyCode ?? amount.currencyCode, locale: screens.locale)
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
        let saved = Money.format(
            home.hasSaved ? home.saved.amount : "0",
            currency: home.currencyCode,
            locale: screens.locale
        )
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

    /// One month's Recap pager: a past-Recap row's month, last month, or Home's own month for the preview (D47).
    func recap(_ month: String) -> RecapViewModel {
        screens.recap(month, onHome: { [weak self] in self?.replace(with: $0) })
    }

    /// The month before Home's when it is in the past-Recap list, what "See last month" opens; nil disables it
    /// (PRD §6.8).
    var lastMonth: String? {
        guard let home else { return nil }
        let parts = home.month.split(separator: "-").compactMap { Int($0) }
        guard parts.count == 2 else { return nil }
        let index = parts[0] * 12 + parts[1] - 2
        let month = String(format: "%04d-%02d", index / 12, index % 12 + 1)
        return home.recapMonths.contains { $0.month == month } ? month : nil
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
