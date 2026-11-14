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
    private let getRecap: RecapViewModel.Fetch
    private let getSettings: SettingsViewModel.Fetch
    private let saveProfile: NameAvatarViewModel.Save
    private let saveSalary: SalaryEditViewModel.Save
    private let saveTarget: TargetEditViewModel.Save
    private let previewCurrency: CurrencyEditViewModel.Preview
    private let changeCurrency: CurrencyEditViewModel.Change
    private let saveNotifications: NotificationsViewModel.Save
    private let requestExport: ExportViewModel.Send
    private let requestDeletion: DeleteAccountViewModel.Request
    private let logOut: () async -> Void
    private let context: ModelContext
    private let onboardingNeeded: () -> Void
    /// Dates and money print in the device locale (PRD §11); the tests pass a fixed one so the runner's locale cannot
    /// change the copy.
    private let dates: Date.FormatStyle
    private let locale: Locale
    private let dismissals: WarningDismissals
    private let connection: Connection
    private let push: PushRegistration?
    /// A Spend was logged through the sheet still closing, so B5 may ask once it has closed.
    private var loggedInSheet = false

    init(
        fetch: @escaping () async throws -> MoneyRecap_V1_HomeView,
        context: ModelContext,
        locale: Locale = .autoupdatingCurrent,
        dismissals: WarningDismissals? = nil,
        connection: Connection? = nil,
        push: PushRegistration? = nil,
        onboardingNeeded: @escaping () -> Void = {},
        logSpend: @escaping LogSpendViewModel.Send = unavailable,
        listSpends: @escaping SpendListViewModel.Fetch = { _ in throw ConnectError(code: .unavailable, message: nil) },
        editSpend: @escaping LogSpendViewModel.Edit = unavailable,
        deleteSpend: @escaping LogSpendViewModel.Delete = unavailable,
        getRecap: @escaping RecapViewModel.Fetch = { _ in throw ConnectError(code: .unavailable, message: nil) },
        getSettings: @escaping SettingsViewModel.Fetch = { throw ConnectError(code: .unavailable, message: nil) },
        saveProfile: @escaping NameAvatarViewModel.Save = { _, _ in .offline },
        saveSalary: @escaping SalaryEditViewModel.Save = unavailable,
        saveTarget: @escaping TargetEditViewModel.Save = unavailable,
        previewCurrency: @escaping CurrencyEditViewModel.Preview = unavailable,
        changeCurrency: @escaping CurrencyEditViewModel.Change = unavailable,
        saveNotifications: @escaping NotificationsViewModel.Save = { _, _ in nil },
        requestExport: @escaping ExportViewModel.Send = unavailable,
        requestDeletion: @escaping DeleteAccountViewModel.Request = unavailable,
        logOut: @escaping () async -> Void = {}
    ) {
        self.fetch = fetch
        self.logSpend = logSpend
        self.editSpend = editSpend
        self.deleteSpend = deleteSpend
        self.listSpends = listSpends
        self.getRecap = getRecap
        self.getSettings = getSettings
        self.saveProfile = saveProfile
        self.saveSalary = saveSalary
        self.saveTarget = saveTarget
        self.previewCurrency = previewCurrency
        self.changeCurrency = changeCurrency
        self.saveNotifications = saveNotifications
        self.requestExport = requestExport
        self.requestDeletion = requestDeletion
        self.logOut = logOut
        self.context = context
        self.onboardingNeeded = onboardingNeeded
        dates = ISODate.style(locale: locale)
        self.locale = locale
        self.dismissals = dismissals ?? .session
        self.connection = connection ?? .shared
        self.push = push
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
                earliest: $0.earliestSpendDate, locale: locale, connection: connection, edit: editSpend,
                delete: deleteSpend
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
            locale: locale,
            connection: connection
        ) }
    }

    /// D1 Settings over its stored copy (IOS I4), in the locale Home was given. A saved Target's Home, and a changed
    /// Currency's, replaces this one.
    func settings() -> SettingsViewModel {
        let save = saveTarget, change = changeCurrency
        return SettingsViewModel(
            fetch: getSettings, saveProfile: saveProfile, saveSalary: saveSalary, saveTarget: { [weak self] in
                let result = await save($0, $1)
                return self?.keep(result, \.home) ?? result
            }, previewCurrency: previewCurrency, changeCurrency: { [weak self] in
                let result = await change($0, $1)
                return self?.keep(result, \.home) ?? result
            }, saveNotifications: saveNotifications, push: push, requestExport: requestExport,
            requestDeletion: requestDeletion, logOut: logOut, context: context, locale: locale, connection: connection
        )
    }

    /// A write's result, its Home replacing this one when it succeeded.
    private func keep<Answer>(
        _ result: Result<Answer, ConnectError>,
        _ home: (Answer) -> MoneyRecap_V1_HomeView
    ) -> Result<Answer, ConnectError> {
        if case let .success(answer) = result {
            replace(with: home(answer))
        }
        return result
    }

    /// D4 over a Recap: Settings read afresh, or its stored copy when that fails; nil with neither.
    func targetEdit() async -> TargetEditViewModel? {
        let settings = settings()
        await settings.refresh()
        settings.openTarget()
        return settings.targetEdit
    }

    /// A write's answer replaces the stored Home whole, as a refresh does (IOS I4).
    func replace(with fresh: MoneyRecap_V1_HomeView) {
        try? HomeSnapshot.replace(with: fresh, in: context)
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
        await push?.askOnce()
    }

    /// Each launch registers an allowed phone for pushes again (IOS I8).
    func launched() async {
        await push?.launched()
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

    /// One month's Recap pager: a past-Recap row's month, last month, or Home's own month for the preview (D47).
    func recap(_ month: String) -> RecapViewModel {
        RecapViewModel(month: month, fetch: getRecap, context: context, locale: locale, openTarget: targetEdit)
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
