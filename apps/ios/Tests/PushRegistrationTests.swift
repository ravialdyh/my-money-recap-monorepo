import Connect
import SwiftData
import UserNotifications
import XCTest
@testable import MyMoneyRecap

/// B5 and IOS I8 over a faked iOS: the alert, APNs and the API are all stand-ins, so no test shows the system alert.
@MainActor
final class PushRegistrationTests: XCTestCase {
    /// iOS as the tests see it: `status` until the alert is answered with `allows`, and APNs answering `token`.
    private final class FakeSystem {
        var status = UNAuthorizationStatus.notDetermined
        var allows = true
        var token: Data? = Data([0xDE, 0xAD, 0x01])
        var asked = 0
        var registered = 0

        var system: PushSystem {
            PushSystem(
                status: { self.status },
                ask: {
                    self.asked += 1
                    self.status = self.allows ? .authorized : .denied
                    return self.allows
                },
                register: {
                    self.registered += 1
                    return self.token
                }
            )
        }
    }

    private final class FakeAPI {
        var answer = SaveResult.saved
        var sent: [(token: String, key: String?)] = []

        func send(_ token: String, _ headers: Headers) -> SaveResult {
            sent.append((token, headers["Idempotency-Key"]?.first))
            return answer
        }
    }

    private func defaults() throws -> UserDefaults {
        let suite = "PushRegistrationTests.\(name)"
        let defaults = try XCTUnwrap(UserDefaults(suiteName: suite))
        defaults.removePersistentDomain(forName: suite)
        return defaults
    }

    private func push(_ fake: FakeSystem, _ api: FakeAPI) throws -> PushRegistration {
        try PushRegistration(system: fake.system, defaults: defaults()) { api.send($0, $1) }
    }

    func testAllowingRegistersAndSendsTheTokenAsHexUnderAKey() async throws {
        let fake = FakeSystem(), api = FakeAPI()
        let push = try push(fake, api)
        await push.spendSaved()
        XCTAssertEqual(fake.asked, 1, "iOS never asked, so the alert shows")
        XCTAssertEqual(fake.registered, 1, "allowed, so the phone registers")
        XCTAssertEqual(api.sent.map(\.token), ["dead01"])
        XCTAssertNotNil(api.sent.first?.key, "RegisterDevice needs the Idempotency-Key header (A14)")
        XCTAssertEqual(push.token, "dead01", "remembered for Log out")
    }

    func testTheAlertNeverShowsTwice() async throws {
        let fake = FakeSystem(), api = FakeAPI()
        let push = try push(fake, api)
        await push.spendSaved()
        await push.spendSaved()
        XCTAssertEqual(fake.asked, 1, "iOS has asked now, whatever the answer")
        fake.status = .denied
        await push.spendSaved()
        XCTAssertEqual(fake.asked, 1)
    }

    func testDenyingRegistersNothing() async throws {
        let fake = FakeSystem(), api = FakeAPI()
        fake.allows = false
        let push = try push(fake, api)
        await push.spendSaved()
        XCTAssertEqual(fake.asked, 1)
        XCTAssertEqual(fake.registered, 0)
        XCTAssertNil(push.token)
    }

    /// The token can change, so every launch of an allowed phone registers again; an unasked or denied one never does.
    func testEveryLaunchRegistersOnlyWhenAllowed() async throws {
        let fake = FakeSystem(), api = FakeAPI()
        let push = try push(fake, api)
        for status in [UNAuthorizationStatus.notDetermined, .denied] {
            fake.status = status
            await push.launched()
        }
        XCTAssertEqual(fake.registered, 0)
        XCTAssertEqual(fake.asked, 0, "a launch never asks")
        fake.status = .authorized
        await push.launched()
        await push.launched()
        XCTAssertEqual(fake.registered, 2)
        XCTAssertEqual(api.sent.count, 2)
    }

    /// No token from APNs (the simulator until #87) and a refused RegisterDevice are both ignored, and neither is
    /// remembered.
    func testAFailedRegistrationIsIgnored() async throws {
        let fake = FakeSystem(), api = FakeAPI()
        fake.status = .authorized
        fake.token = nil
        let push = try push(fake, api)
        await push.launched()
        XCTAssertTrue(api.sent.isEmpty)
        fake.token = Data([0x0A])
        api.answer = .offline
        await push.launched()
        XCTAssertEqual(api.sent.map(\.token), ["0a"])
        XCTAssertNil(push.token)
    }

    /// Home asks only after the Log a spend sheet closes on a save, once per save, and a sheet closed without one never
    /// asks.
    func testHomeAsksOnlyAfterALoggedSpendsSheetCloses() async throws {
        let fake = FakeSystem(), api = FakeAPI()
        let context = try ModelContext(ModelContainer(
            for: HomeSnapshot.self, SpendListSnapshot.self, RecapSnapshot.self, SettingsSnapshot.self,
            configurations: ModelConfiguration(isStoredInMemoryOnly: true)
        ))
        let home = try HomeViewModel(fetch: { .init() }, context: context, push: push(fake, api))
        await home.logSheetClosed()
        XCTAssertEqual(fake.asked, 0, "closed with no save")
        home.logged(.with { $0.displayName = "Test" })
        XCTAssertEqual(home.greeting, "Hi, Test", "the save's Home replaces the numbers at once")
        await home.logSheetClosed()
        await home.logSheetClosed()
        XCTAssertEqual(fake.asked, 1)
        await home.launched()
        XCTAssertEqual(fake.registered, 2, "the save's registration, then the launch's")
    }
}
