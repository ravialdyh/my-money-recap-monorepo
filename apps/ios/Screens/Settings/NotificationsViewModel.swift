import Connect
import Foundation
import MoneyRecapContract
import UIKit

/// D8 Notifications (PRD §6.16): three switches and the reminder's time, each change sending the whole message under a
/// new key (A14); a failure puts the switch back. Denied, a switch opens iOS Settings; never asked, turning one on
/// asks.
@Observable
@MainActor
final class NotificationsViewModel: Identifiable {
    /// The saved preferences as notify answered them, or nil when the save failed.
    typealias Save = (MoneyRecap_V1_NotificationPrefs, Headers) async -> MoneyRecap_V1_NotificationPrefs?

    enum Switch {
        case recapReady, dailyReminder, targetWarning
    }

    /// What the switches show: the last change, until a failure puts back `saved`.
    private(set) var prefs: MoneyRecap_V1_NotificationPrefs
    /// iOS said no: the info Banner, the switches at 40%, and every tap opens iOS Settings.
    private(set) var denied = false
    /// The last save failed, over `SettingsViewModel.saveFailure`.
    private(set) var failed = false
    private var saved: MoneyRecap_V1_NotificationPrefs
    private let push: PushRegistration
    private let connection: Connection
    private let save: Save
    private let openSettings: @MainActor () -> Void

    init(
        prefs: MoneyRecap_V1_NotificationPrefs,
        push: PushRegistration,
        connection: Connection? = nil,
        openSettings: @escaping @MainActor () -> Void = NotificationsViewModel.openSystemSettings,
        save: @escaping Save
    ) {
        self.prefs = prefs
        saved = prefs
        self.push = push
        self.connection = connection ?? .shared
        self.openSettings = openSettings
        self.save = save
    }

    /// The app's own page in iOS Settings, where its notifications are switched (the Banner's `Open Settings`).
    static func openSystemSettings() {
        if let url = URL(string: UIApplication.openNotificationSettingsURLString) {
            UIApplication.shared.open(url)
        }
    }

    /// Asked on appearing and whenever the app comes back, since the Member may have changed it in iOS Settings.
    func readPermission() async {
        denied = await push.status() == .denied
    }

    /// No network path: every switch at 40% and inert.
    var offline: Bool {
        !connection.online
    }

    func isOn(_ which: Switch) -> Bool {
        switch which {
        case .recapReady: prefs.recapReady
        case .dailyReminder: prefs.dailyReminder
        case .targetWarning: prefs.targetWarning
        }
    }

    /// A switch moved: iOS Settings when denied; the system alert first when turning one on and iOS never asked, then
    /// nothing unless the Member allowed; otherwise the save.
    func flip(_ which: Switch, to on: Bool) async {
        guard !denied else { return openSettings() }
        if on, await push.status() == .notDetermined {
            await push.askOnce()
            await readPermission()
            guard await push.status() != .notDetermined, !denied else { return }
        }
        var next = prefs
        switch which {
        case .recapReady: next.recapReady = on
        case .dailyReminder: next.dailyReminder = on
        case .targetWarning: next.targetWarning = on
        }
        await send(next)
    }

    /// The reminder's time as the wheel shows it: today at the saved "HH:MM", in the phone's calendar.
    var reminderTime: Date {
        let parts = prefs.reminderTime.split(separator: ":").compactMap { Int($0) }
        let calendar = Calendar.current
        return calendar.date(
            bySettingHour: parts.first ?? 20, minute: parts.last ?? 0, second: 0, of: .now
        ) ?? .now
    }

    /// The wheel settled on a time: saved as "HH:MM" on a 24-hour clock whatever the device's format.
    func pick(_ time: Date) async {
        let parts = Calendar.current.dateComponents([.hour, .minute], from: time)
        var next = prefs
        next.reminderTime = String(format: "%02d:%02d", parts.hour ?? 20, parts.minute ?? 0)
        guard next != prefs else { return }
        await send(next)
    }

    /// Shows `next` at once; its answer becomes the saved state, and a failure puts back the saved state unless a newer
    /// change has already replaced `next`.
    private func send(_ next: MoneyRecap_V1_NotificationPrefs) async {
        prefs = next
        failed = false
        if let answer = await save(next, SaveAttempt(next).headers) {
            saved = answer
            if prefs == next {
                prefs = answer
            }
        } else {
            failed = true
            if prefs == next {
                prefs = saved
            }
        }
    }
}
