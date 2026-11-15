import Connect
import Foundation
import MoneyRecapContract
import SwiftData

@Observable
@MainActor
final class SettingsViewModel: Identifiable {
    typealias Fetch = () async throws -> MoneyRecap_V1_SettingsView

    static let saveFailure = "We couldn't save this. Try again."

    private(set) var settings: MoneyRecap_V1_SettingsView?
    var profile: ProfileViewModel?
    var salaryEdit: SalaryEditViewModel?
    var targetEdit: TargetEditViewModel?
    var currencyEdit: CurrencyEditViewModel?
    var notifications: NotificationsViewModel?
    private(set) var pushDenied = false
    var export: ExportViewModel?
    var deleteAccount: DeleteAccountViewModel?
    private let calls: SettingsCalls
    private let push: PushRegistration?
    private let context: ModelContext
    private let locale: Locale
    private let connection: Connection

    init(
        calls: SettingsCalls,
        push: PushRegistration? = nil,
        context: ModelContext,
        locale: Locale = .autoupdatingCurrent,
        connection: Connection? = nil
    ) {
        self.calls = calls
        self.push = push
        self.context = context
        self.locale = locale
        self.connection = connection ?? .shared
        settings = SettingsSnapshot.load(from: context)
    }

    func refresh() async {
        pushDenied = await push?.status() == .denied
        guard let fresh = try? await calls.fetch() else { return }
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

    var offline: Bool {
        !connection.online
    }

    var name: String? {
        settings?.profile.displayName
    }

    var currency: String? {
        settings?.profile.currencyCode
    }

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

    var target: String? {
        guard let ledger = settings?.ledgerSettings else { return nil }
        return ledger.hasSavingTarget ? "\(ledger.savingTarget.percent)%" : "Not set"
    }

    var notificationsValue: String? {
        if pushDenied {
            return "Off in iOS Settings"
        }
        guard let prefs = settings?.notificationPrefs, settings?.hasNotificationPrefs == true else { return nil }
        let on = [prefs.recapReady, prefs.dailyReminder, prefs.targetWarning].filter(\.self).count
        return on == 0 ? "All off" : "\(on) on"
    }

    func openNotifications() {
        guard let push, let saved = settings, saved.hasNotificationPrefs else { return }
        notifications = NotificationsViewModel(
            prefs: saved.notificationPrefs, push: push, connection: connection, save: calls.saveNotifications
        )
    }

    func openProfile() {
        guard let saved = settings?.profile else { return }
        profile = ProfileViewModel(name: saved.displayName, avatar: saved.avatarKey, save: calls.saveProfile)
    }

    func openSalary() {
        guard let saved = settings else { return }
        let save = calls.saveSalary
        salaryEdit = SalaryEditViewModel(settings: saved, locale: locale) { [weak self] request, headers in
            let result = await save(request, headers)
            if case let .success(answer) = result {
                self?.replace(with: answer.settings)
            }
            return result
        }
    }

    func openTarget() {
        guard let saved = settings else { return }
        let save = calls.saveTarget
        targetEdit = TargetEditViewModel(settings: saved, locale: locale) { [weak self] request, headers in
            let result = await save(request, headers)
            if case let .success(answer) = result {
                self?.replace(with: answer.settings)
            }
            return result
        }
    }

    func openCurrency() {
        guard let saved = settings else { return }
        let change = calls.changeCurrency
        currencyEdit = CurrencyEditViewModel(
            current: saved.profile.currencyCode, locale: locale, preview: calls.previewCurrency
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

    func openExport() {
        guard let saved = settings?.profile else { return }
        export = ExportViewModel(email: saved.email, send: calls.requestExport)
    }

    func openDeleteAccount() {
        deleteAccount = DeleteAccountViewModel(request: calls.requestDeletion, signOut: calls.logOut)
    }

    func logOutNow() async {
        await calls.logOut()
    }

    func setSalaryFirst() {
        targetEdit = nil
        openSalary()
    }

    static func payday(_ stored: String) -> String {
        guard let day = Int(stored) else { return "Last day" }
        let ordinal = NumberFormatter()
        ordinal.numberStyle = .ordinal
        ordinal.locale = Locale(identifier: "en")
        return ordinal.string(from: day as NSNumber) ?? stored
    }
}
