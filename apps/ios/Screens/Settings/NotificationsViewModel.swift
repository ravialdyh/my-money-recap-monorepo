import Connect
import Foundation
import MoneyRecapContract
import UIKit

@Observable
@MainActor
final class NotificationsViewModel: Identifiable {
    typealias Save = (MoneyRecap_V1_NotificationPrefs, Headers) async -> MoneyRecap_V1_NotificationPrefs?

    enum Switch {
        case recapReady, dailyReminder, targetWarning
    }

    private(set) var prefs: MoneyRecap_V1_NotificationPrefs
    private(set) var denied = false
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

    static func openSystemSettings() {
        if let url = URL(string: UIApplication.openNotificationSettingsURLString) {
            UIApplication.shared.open(url)
        }
    }

    func readPermission() async {
        denied = await push.status() == .denied
    }

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

    var reminderTime: Date {
        let parts = prefs.reminderTime.split(separator: ":").compactMap { Int($0) }
        let calendar = Calendar.current
        return calendar.date(
            bySettingHour: parts.first ?? 20, minute: parts.last ?? 0, second: 0, of: .now
        ) ?? .now
    }

    func pick(_ time: Date) async {
        let parts = Calendar.current.dateComponents([.hour, .minute], from: time)
        var next = prefs
        next.reminderTime = String(format: "%02d:%02d", parts.hour ?? 20, parts.minute ?? 0)
        guard next != prefs else { return }
        await send(next)
    }

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
