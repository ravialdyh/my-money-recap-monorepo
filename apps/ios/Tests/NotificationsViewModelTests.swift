import Connect
import MoneyRecapContract
import UserNotifications
import XCTest
@testable import MyMoneyRecap

@MainActor
final class NotificationsViewModelTests: XCTestCase {
    @MainActor
    final class FakeIOS {
        var status: UNAuthorizationStatus
        var allows = true
        var asked = 0
        var opened = 0

        init(_ status: UNAuthorizationStatus) {
            self.status = status
        }

        var push: PushRegistration {
            PushRegistration(
                system: PushSystem(
                    status: { self.status },
                    ask: {
                        self.asked += 1
                        self.status = self.allows ? .authorized : .denied
                        return self.allows
                    },
                    register: { nil },
                    unregister: {}
                ),
                defaults: UserDefaults(suiteName: "NotificationsViewModelTests") ?? .standard,
                send: { _, _ in .saved },
                forget: { _, _ in .saved }
            )
        }
    }

    final class FakeAPI {
        var fails = false
        var sent: [(prefs: MoneyRecap_V1_NotificationPrefs, key: String?)] = []

        func save(_ prefs: MoneyRecap_V1_NotificationPrefs, _ headers: Headers) -> MoneyRecap_V1_NotificationPrefs? {
            sent.append((prefs, headers["Idempotency-Key"]?.first))
            return fails ? nil : prefs
        }
    }

    nonisolated static let defaults = MoneyRecap_V1_NotificationPrefs.with {
        $0.recapReady = true
        $0.reminderTime = "20:00"
        $0.targetWarning = true
    }

    static func model(
        _ ios: FakeIOS,
        _ api: FakeAPI,
        prefs: MoneyRecap_V1_NotificationPrefs = defaults,
        online: Bool = true
    ) -> NotificationsViewModel {
        NotificationsViewModel(
            prefs: prefs, push: ios.push, connection: Connection(online: online), openSettings: { ios.opened += 1 },
            save: { api.save($0, $1) }
        )
    }

    func testEachChangeSendsTheWholeMessage() async {
        let ios = FakeIOS(.authorized), api = FakeAPI()
        let model = Self.model(ios, api)
        await model.readPermission()
        XCTAssertFalse(model.denied)
        await model.flip(.dailyReminder, to: true)
        await model.flip(.targetWarning, to: false)
        XCTAssertTrue(model.isOn(.dailyReminder))
        XCTAssertFalse(model.isOn(.targetWarning))
        XCTAssertEqual(api.sent.count, 2)
        XCTAssertEqual(api.sent.last?.prefs, .with {
            $0.recapReady = true
            $0.dailyReminder = true
            $0.reminderTime = "20:00"
        }, "the whole message, the first change included")
        XCTAssertNotEqual(api.sent[0].key, api.sent[1].key, "each change is its own write")
        XCTAssertEqual(ios.asked, 0, "already allowed: no alert")
    }

    func testAFailedSavePutsTheSwitchBack() async {
        let ios = FakeIOS(.authorized), api = FakeAPI()
        let model = Self.model(ios, api)
        api.fails = true
        await model.flip(.recapReady, to: false)
        XCTAssertTrue(model.isOn(.recapReady), "put back")
        XCTAssertTrue(model.failed)
        api.fails = false
        await model.flip(.recapReady, to: false)
        XCTAssertFalse(model.isOn(.recapReady))
        XCTAssertFalse(model.failed)
    }

    func testDeniedOpensIOSSettingsAndSavesNothing() async {
        let ios = FakeIOS(.denied), api = FakeAPI()
        let model = Self.model(ios, api)
        await model.readPermission()
        XCTAssertTrue(model.denied)
        await model.flip(.dailyReminder, to: true)
        XCTAssertEqual(ios.opened, 1)
        XCTAssertFalse(model.isOn(.dailyReminder))
        XCTAssertTrue(api.sent.isEmpty)
        XCTAssertEqual(ios.asked, 0, "iOS asks once only; after a no, only iOS Settings can change it")
    }

    func testNeverAskedAsksFirst() async {
        let ios = FakeIOS(.notDetermined), api = FakeAPI()
        let model = Self.model(ios, api)
        await model.flip(.targetWarning, to: false)
        XCTAssertEqual(ios.asked, 0, "turning one off asks nothing")
        await model.flip(.dailyReminder, to: true)
        XCTAssertEqual(ios.asked, 1)
        XCTAssertTrue(model.isOn(.dailyReminder), "allowed, so saved")
        XCTAssertEqual(api.sent.count, 2)

        let refusing = FakeIOS(.notDetermined), none = FakeAPI()
        refusing.allows = false
        let refused = Self.model(refusing, none)
        await refused.flip(.dailyReminder, to: true)
        XCTAssertEqual(refusing.asked, 1)
        XCTAssertFalse(refused.isOn(.dailyReminder))
        XCTAssertTrue(refused.denied, "the Banner shows")
        XCTAssertTrue(none.sent.isEmpty)
    }

    func testThePickedTimeIsSavedAsHoursAndMinutes() async throws {
        let ios = FakeIOS(.authorized), api = FakeAPI()
        let model = Self.model(ios, api)
        XCTAssertEqual(Calendar.current.dateComponents([.hour, .minute], from: model.reminderTime).hour, 20)
        let seven = try XCTUnwrap(Calendar.current.date(bySettingHour: 7, minute: 5, second: 0, of: .now))
        await model.pick(seven)
        await model.pick(seven)
        XCTAssertEqual(api.sent.map(\.prefs.reminderTime), ["07:05"])
        XCTAssertEqual(model.prefs.reminderTime, "07:05")
    }

    func testOfflineIsTheConnections() {
        XCTAssertTrue(Self.model(FakeIOS(.authorized), FakeAPI(), online: false).offline)
    }

    func testTheSettingsRowCountsWhatIsOn() async throws {
        let context = try SettingsViewModelTests.context()
        var answer = try SettingsViewModelTests.recorded("settings-filled")
        let ios = FakeIOS(.authorized)
        let settings = SettingsViewModel(
            calls: SettingsCalls(fetch: { answer }, saveProfile: { _, _ in .saved }), push: ios.push, context: context,
            connection: Connection(online: true)
        )
        XCTAssertNil(settings.notificationsValue, "nothing stored and no answer yet")
        settings.openNotifications()
        XCTAssertNil(settings.notifications)
        await settings.refresh()
        XCTAssertEqual(settings.notificationsValue, "3 on")
        settings.openNotifications()
        XCTAssertEqual(settings.notifications?.prefs, answer.notificationPrefs)

        answer.clearNotificationPrefs()
        await settings.refresh()
        XCTAssertEqual(settings.notificationsValue, "3 on", "notify did not answer: the stored ones stay")
        answer.notificationPrefs = .with { $0.reminderTime = "20:00" }
        await settings.refresh()
        XCTAssertEqual(settings.notificationsValue, "All off")
        ios.status = .denied
        await settings.refresh()
        XCTAssertEqual(settings.notificationsValue, "Off in iOS Settings")

        let bare = try SettingsViewModel(
            calls: SettingsCalls(fetch: {
                var none = try SettingsViewModelTests.recorded("settings-not-set")
                none.clearNotificationPrefs()
                return none
            }, saveProfile: { _, _ in .saved }), push: ios.push, context: SettingsViewModelTests.context()
        )
        ios.status = .authorized
        await bare.refresh()
        XCTAssertNil(bare.notificationsValue, "notify down with nothing stored: the row without a count")
    }
}
