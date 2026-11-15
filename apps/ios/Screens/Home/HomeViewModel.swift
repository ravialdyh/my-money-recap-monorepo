import Connect
import Foundation
import MoneyRecapContract
import SwiftData

@Observable
@MainActor
final class HomeViewModel {
    private(set) var home: MoneyRecap_V1_HomeView?
    private let fetch: () async throws -> MoneyRecap_V1_HomeView
    private let screens: HomeScreens
    private let onboardingNeeded: () -> Void
    private let dates: Date.FormatStyle
    private let dismissals: WarningDismissals
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

    func logSheet() -> LogSpendViewModel? {
        home.map { screens.logSpend(for: $0) }
    }

    func editSheet(_ spend: MoneyRecap_V1_Spend) -> LogSpendViewModel? {
        home.map { screens.editSpend(spend, for: $0) }
    }

    func spendList(switching: Bool) -> SpendListViewModel? {
        home.map { screens.spendList(for: $0, switching: switching) }
    }

    func settings() -> SettingsViewModel {
        screens.settings(onHome: { [weak self] in self?.replace(with: $0) })
    }

    func replace(with fresh: MoneyRecap_V1_HomeView) {
        try? HomeSnapshot.replace(with: fresh, in: screens.context)
        home = fresh
    }

    func logged(_ fresh: MoneyRecap_V1_HomeView) {
        replace(with: fresh)
        loggedInSheet = true
    }

    func logSheetClosed() async {
        guard loggedInSheet else { return }
        loggedInSheet = false
        await screens.push?.askOnce()
    }

    func launched() async {
        await screens.push?.launched()
    }

    var offline: Bool {
        !screens.connection.online
    }

    var greeting: String? {
        home.map { "Hi, \($0.displayName)" }
    }

    var heroEyebrow: String {
        guard let home, home.hasSalary, home.salary.daysUntilPayday > 0 else { return "Spent so far" }
        let days = home.salary.daysUntilPayday
        return "Spent so far · salary lands in \(days) \(days == 1 ? "day" : "days")"
    }

    func money(_ amount: MoneyRecap_V1_Money) -> String {
        Money.format(amount.amount, currency: home?.currencyCode ?? amount.currencyCode, locale: screens.locale)
    }

    var showsSavedChip: Bool {
        guard let home, home.hasSaved else { return false }
        return (Decimal(string: home.moneyIn.amount) ?? 0) > 0
    }

    var targetProgress: String? {
        guard let home, home.hasSavingTarget else { return nil }
        let saved = Money.format(
            home.hasSaved ? home.saved.amount : "0",
            currency: home.currencyCode,
            locale: screens.locale
        )
        return "\(saved) of \(money(home.savingTarget.amount))"
    }

    var targetFraction: Double {
        guard let home, home.hasSavingTarget, home.hasSaved,
              let saved = Double(home.saved.amount), let whole = Double(home.savingTarget.amount.amount), whole > 0
        else { return 0 }
        return min(max(saved / whole, 0), 1)
    }

    var warning: String? {
        guard let home, home.hasTargetWarning, !dismissals.months.contains(home.month) else { return nil }
        return home.targetWarning.message
    }

    func dismissWarning() {
        if let month = home?.month {
            dismissals.months.insert(month)
        }
    }

    var monthLabel: String? {
        home.flatMap { ISODate.month($0.month) }.map { $0.formatted(dates.month(.wide)) }
    }

    var previewCopy: String? {
        home.flatMap { ISODate.day($0.firstRecapOn) }
            .map { "Your first recap arrives on \($0.formatted(dates.day().month(.wide)))." }
    }

    func recap(_ month: String) -> RecapViewModel {
        screens.recap(month, onHome: { [weak self] in self?.replace(with: $0) })
    }

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

@Observable
@MainActor
final class WarningDismissals {
    static let session = WarningDismissals()
    var months: Set<String> = []
}
