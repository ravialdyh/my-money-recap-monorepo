import Connect
import Foundation
import UIKit
import UserNotifications

/// iOS's side of push permission and registration (IOS I8): the seam the unit tests fake, so their host never shows the
/// system alert.
struct PushSystem {
    var status: () async -> UNAuthorizationStatus
    /// The system alert; true when the Member allowed.
    var ask: () async -> Bool
    /// APNs' device token, or nil when registration failed, the simulator's normal case until #87.
    var register: @MainActor () async -> Data?

    static let live = PushSystem(
        status: { await UNUserNotificationCenter.current().notificationSettings().authorizationStatus },
        ask: { await allowed(options: [.alert, .sound]) },
        register: { await AppDelegate.register() }
    )

    private static func allowed(options: UNAuthorizationOptions) async -> Bool {
        await (try? UNUserNotificationCenter.current().requestAuthorization(options: options)) ?? false
    }
}

/// B5 (PRD §10, IOS I8): asks once, after a Spend is saved or a D8 switch turns on, when iOS never asked; once allowed,
/// registers on every launch and hands the token as hex to RegisterDevice, remembering it for Log out. A failure is
/// ignored.
@MainActor
final class PushRegistration {
    typealias Send = (_ token: String, _ headers: Headers) async -> SaveResult
    static let tokenKey = "apnsToken"
    private let system: PushSystem
    private let send: Send
    private let defaults: UserDefaults

    init(system: PushSystem = .live, defaults: UserDefaults = .standard, send: @escaping Send) {
        self.system = system
        self.defaults = defaults
        self.send = send
    }

    /// The token RegisterDevice last accepted, which Log out asks the API to forget.
    var token: String? {
        defaults.string(forKey: Self.tokenKey)
    }

    /// Every launch registers an allowed phone again, since the token can change.
    func launched() async {
        if await [.authorized, .provisional, .ephemeral].contains(system.status()) {
            await register()
        }
    }

    /// iOS's answer so far: D8 reads denied, never asked and allowed from it.
    func status() async -> UNAuthorizationStatus {
        await system.status()
    }

    /// A Spend save's sheet has closed or a D8 switch turned on: the system alert when iOS never asked, then
    /// registration when allowed.
    func askOnce() async {
        guard await system.status() == .notDetermined, await system.ask() else { return }
        await register()
    }

    private func register() async {
        guard let token = await system.register() else { return }
        let hex = token.map { String(format: "%02x", $0) }.joined()
        if await send(hex, SaveAttempt(hex).headers) == .saved {
            defaults.set(hex, forKey: Self.tokenKey)
        }
    }
}

/// Hands APNs' answer to `registerForRemoteNotifications` to whoever is waiting for it. With no notification centre
/// delegate, a push is never shown while the app is open, and tapping one opens it as a normal launch.
final class AppDelegate: NSObject, UIApplicationDelegate {
    private static var waiting: [CheckedContinuation<Data?, Never>] = []

    static func register() async -> Data? {
        await withCheckedContinuation { continuation in
            waiting.append(continuation)
            UIApplication.shared.registerForRemoteNotifications()
        }
    }

    func application(_: UIApplication, didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data) {
        Self.answer(deviceToken)
    }

    func application(_: UIApplication, didFailToRegisterForRemoteNotificationsWithError _: Error) {
        Self.answer(nil)
    }

    private static func answer(_ token: Data?) {
        let answered = waiting
        waiting = []
        answered.forEach { $0.resume(returning: token) }
    }
}
