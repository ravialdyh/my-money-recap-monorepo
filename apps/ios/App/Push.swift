import Connect
import Foundation
import MoneyRecapContract
import UIKit
import UserNotifications

struct PushSystem {
    var status: () async -> UNAuthorizationStatus
    var ask: () async -> Bool
    var register: @MainActor () async -> Data?
    var unregister: @MainActor () -> Void

    static let live = PushSystem(
        status: { await UNUserNotificationCenter.current().notificationSettings().authorizationStatus },
        ask: { await allowed(options: [.alert, .sound]) },
        register: { await AppDelegate.register() },
        unregister: { UIApplication.shared.unregisterForRemoteNotifications() }
    )

    private static func allowed(options: UNAuthorizationOptions) async -> Bool {
        await (try? UNUserNotificationCenter.current().requestAuthorization(options: options)) ?? false
    }
}

@MainActor
final class PushRegistration {
    typealias Send = (_ token: String, _ headers: Headers) async -> SaveResult
    static let tokenKey = "apnsToken"
    private let system: PushSystem
    private let send: Send
    private let forgetToken: Send
    private let defaults: UserDefaults

    init(system: PushSystem, defaults: UserDefaults, send: @escaping Send, forget: @escaping Send) {
        self.system = system
        self.defaults = defaults
        self.send = send
        forgetToken = forget
    }

    static func live(session: AuthSession, client: MoneyRecap_V1_BffServiceClient) -> PushRegistration {
        PushRegistration(system: .live, defaults: .standard, send: { token, headers in
            await SaveResult(BffClient.retried(session) {
                await client.registerDevice(request: .with { $0.apnsToken = token }, headers: headers)
            }.result)
        }, forget: { token, headers in
            await SaveResult(BffClient.retried(session) {
                await client.forgetDevice(request: .with { $0.apnsToken = token }, headers: headers)
            }.result)
        })
    }

    static func signedOut(system: PushSystem = .live, defaults: UserDefaults = .standard) {
        system.unregister()
        defaults.removeObject(forKey: tokenKey)
    }

    var token: String? {
        defaults.string(forKey: Self.tokenKey)
    }

    func forget() async {
        guard let token else { return }
        _ = await forgetToken(token, SaveAttempt(token).headers)
    }

    func launched() async {
        if await [.authorized, .provisional, .ephemeral].contains(system.status()) {
            await register()
        }
    }

    func status() async -> UNAuthorizationStatus {
        await system.status()
    }

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
