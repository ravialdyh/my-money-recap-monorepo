import Connect
import Foundation
import MoneyRecapContract
import SwiftData

/// D1 Settings (PRD §6.16, IOS I4, I6): the stored `SettingsView` at once, then the BFF's, each answer replacing the
/// stored copy whole. Every value is the server's; this only words it. Offline, every row that writes is disabled.
@Observable
@MainActor
final class SettingsViewModel: Identifiable {
    typealias Fetch = () async throws -> MoneyRecap_V1_SettingsView

    /// The line every Settings save shows when it fails (spec #169).
    static let saveFailure = "We couldn't save this. Try again."

    /// Nil until a stored copy or an answer exists: the rows show without values.
    private(set) var settings: MoneyRecap_V1_SettingsView?
    /// D2 Profile, while it is open.
    var profile: ProfileViewModel?
    /// D3 Salary and payday, while it is open.
    var salaryEdit: SalaryEditViewModel?
    /// D4 Saving target, while it is open.
    var targetEdit: TargetEditViewModel?
    /// D5 Currency, while it is open.
    var currencyEdit: CurrencyEditViewModel?
    /// D8 Notifications, while it is open.
    var notifications: NotificationsViewModel?
    /// iOS said no to notifications, which the Notifications row reads before any count.
    private(set) var pushDenied = false
    /// D9 Export spends, while it is open.
    var export: ExportViewModel?
    /// D10 Delete account, while it is open.
    var deleteAccount: DeleteAccountViewModel?
    private let fetch: Fetch
    private let saveProfile: NameAvatarViewModel.Save
    private let saveSalary: SalaryEditViewModel.Save
    private let saveTarget: TargetEditViewModel.Save
    private let previewCurrency: CurrencyEditViewModel.Preview
    private let changeCurrency: CurrencyEditViewModel.Change
    private let saveNotifications: NotificationsViewModel.Save
    private let push: PushRegistration?
    private let requestExport: ExportViewModel.Send
    private let requestDeletion: DeleteAccountViewModel.Request
    private let logOut: () async -> Void
    private let context: ModelContext
    private let locale: Locale
    private let connection: Connection

    init(
        fetch: @escaping Fetch,
        saveProfile: @escaping NameAvatarViewModel.Save,
        saveSalary: @escaping SalaryEditViewModel.Save = unavailable,
        saveTarget: @escaping TargetEditViewModel.Save = unavailable,
        previewCurrency: @escaping CurrencyEditViewModel.Preview = unavailable,
        changeCurrency: @escaping CurrencyEditViewModel.Change = unavailable,
        saveNotifications: @escaping NotificationsViewModel.Save = { _, _ in nil },
        push: PushRegistration? = nil,
        requestExport: @escaping ExportViewModel.Send = unavailable,
        requestDeletion: @escaping DeleteAccountViewModel.Request = unavailable,
        logOut: @escaping () async -> Void = {},
        context: ModelContext,
        locale: Locale = .autoupdatingCurrent,
        connection: Connection? = nil
    ) {
        self.fetch = fetch
        self.saveProfile = saveProfile
        self.saveSalary = saveSalary
        self.saveTarget = saveTarget
        self.previewCurrency = previewCurrency
        self.changeCurrency = changeCurrency
        self.saveNotifications = saveNotifications
        self.push = push
        self.requestExport = requestExport
        self.requestDeletion = requestDeletion
        self.logOut = logOut
        self.context = context
        self.locale = locale
        self.connection = connection ?? .shared
        settings = SettingsSnapshot.load(from: context)
    }

    /// The answer replaces the stored copy and what is shown; a failure keeps both, offline included. An answer
    /// without notify's preferences keeps the stored ones (spec #169). iOS's permission is asked again each time.
    func refresh() async {
        pushDenied = await push?.status() == .denied
        guard let fresh = try? await fetch() else { return }
        replace(with: fresh)
    }

    private func replace(with answer: MoneyRecap_V1_SettingsView) {
        var fresh = answer
        if !fresh.hasNotificationPrefs, let kept = settings, kept.hasNotificationPrefs {
            fresh.notificationPrefs = kept.notificationPrefs
        }
        try? SettingsSnapshot.replace(with: fresh, in: context)
        settings = fresh
    }

    /// No network path: the grey Banner, and every row but Log out at 40% (D1 offline).
    var offline: Bool {
        !connection.online
    }

    var name: String? {
        settings?.profile.displayName
    }

    var currency: String? {
        settings?.profile.currencyCode
    }

    /// "Rp150.000.000 · 25th", "{amount} · Last day", or "Not set".
    var salary: String? {
        guard let ledger = settings?.ledgerSettings else { return nil }
        guard ledger.hasSalary else { return "Not set" }
        let amount = Money.format(
            ledger.salary.amount.amount,
            currency: ledger.salary.amount.currencyCode,
            locale: locale
        )
        return "\(amount) · \(Self.payday(ledger.salary.payday))"
    }

    /// "20%", or "Not set".
    var target: String? {
        guard let ledger = settings?.ledgerSettings else { return nil }
        return ledger.hasSavingTarget ? "\(ledger.savingTarget.percent)%" : "Not set"
    }

    /// "Off in iOS Settings" when denied, else "{n} on" or "All off"; nil with no preferences stored or answered.
    var notificationsValue: String? {
        if pushDenied {
            return "Off in iOS Settings"
        }
        guard let prefs = settings?.notificationPrefs, settings?.hasNotificationPrefs == true else { return nil }
        let on = [prefs.recapReady, prefs.dailyReminder, prefs.targetWarning].filter(\.self).count
        return on == 0 ? "All off" : "\(on) on"
    }

    /// D8 over the saved preferences; nil while there are none to show.
    func openNotifications() {
        guard let push, let saved = settings, saved.hasNotificationPrefs else { return }
        notifications = NotificationsViewModel(
            prefs: saved.notificationPrefs, push: push, connection: connection, save: saveNotifications
        )
    }

    /// D2 over the saved answers; nil before Settings has any.
    func openProfile() {
        guard let saved = settings?.profile else { return }
        profile = ProfileViewModel(name: saved.displayName, avatar: saved.avatarKey, save: saveProfile)
    }

    /// D3 over the saved answers; nil before Settings has any. A save's Settings replace the stored copy.
    func openSalary() {
        guard let saved = settings else { return }
        let save = saveSalary
        salaryEdit = SalaryEditViewModel(settings: saved, locale: locale) { [weak self] request, headers in
            let result = await save(request, headers)
            if case let .success(answer) = result {
                self?.replace(with: answer.settings)
            }
            return result
        }
    }

    /// D4 over the saved answers; nil before Settings has any. A save's Settings replace the stored copy.
    func openTarget() {
        guard let saved = settings else { return }
        let save = saveTarget
        targetEdit = TargetEditViewModel(settings: saved, locale: locale) { [weak self] request, headers in
            let result = await save(request, headers)
            if case let .success(answer) = result {
                self?.replace(with: answer.settings)
            }
            return result
        }
    }

    /// D5 over the Currency held; nil before Settings has any. A change's Settings replace the stored copy, and every
    /// stored Spend list and Recap is erased, since they hold the old Currency (spec #169).
    func openCurrency() {
        guard let saved = settings else { return }
        let change = changeCurrency
        currencyEdit = CurrencyEditViewModel(
            current: saved.profile.currencyCode, locale: locale, preview: previewCurrency
        ) { [weak self] request, headers in
            let result = await change(request, headers)
            if case let .success(answer) = result, let self {
                replace(with: answer.settings)
                try? context.delete(model: SpendListSnapshot.self)
                try? context.delete(model: RecapSnapshot.self)
                try? context.save()
            }
            return result
        }
    }

    /// D9 naming the address the Member signed in with; nil before Settings has any.
    func openExport() {
        guard let saved = settings?.profile else { return }
        export = ExportViewModel(email: saved.email, send: requestExport)
    }

    /// D10, whose granted request signs out as Log out does.
    func openDeleteAccount() {
        deleteAccount = DeleteAccountViewModel(request: requestDeletion, signOut: logOut)
    }

    /// D12's `Log out`: this phone forgotten, the realm's session ended and the phone's copy erased (PRD §6.16).
    func logOutNow() async {
        await logOut()
    }

    /// D4's `Set a salary first`: D3 in its place (D44).
    func setSalaryFirst() {
        targetEdit = nil
        openSalary()
    }

    /// A stored Payday as the row prints it: "25th", or "Last day" for the token (PRD D51).
    static func payday(_ stored: String) -> String {
        guard let day = Int(stored) else { return "Last day" }
        let ordinal = NumberFormatter()
        ordinal.numberStyle = .ordinal
        ordinal.locale = Locale(identifier: "en")
        return ordinal.string(from: day as NSNumber) ?? stored
    }
}
