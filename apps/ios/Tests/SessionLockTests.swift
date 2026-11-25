import AppAuth
import LocalAuthentication
import XCTest
@testable import MyMoneyRecap

final class FakeVault {
    var stored: Keychain.Stored?
    var present: Bool
    var passcode = true
    var reads: [LAContext] = []
    var writes: [Data] = []
    var deletes = 0
    var writable = true
    var confirmed = true
    var confirms = 0

    init(stored: Keychain.Stored? = nil) {
        self.stored = stored
        present = stored != nil
    }

    var vault: Vault {
        Vault(
            exists: { self.present },
            read: { context in
                self.reads.append(context)
                return self.stored
            },
            write: {
                self.writes.append($0)
                return self.writable
            },
            confirmPresence: { _ in
                self.confirms += 1
                return self.confirmed
            },
            delete: {
                self.deletes += 1
                self.present = false
            },
            passcodeSet: { self.passcode }
        )
    }
}

@MainActor
final class SessionLockTests: XCTestCase {
    func testAStoredEntryStartsLockedWithoutReadingIt() {
        let fake = FakeVault(stored: Self.stored(locked: true))
        let session = AuthSession(vault: fake.vault)
        XCTAssertEqual(session.lock, .locked)
        XCTAssertNil(session.authState)
        XCTAssertTrue(fake.reads.isEmpty, "a read would show the Face ID prompt during init")
    }

    func testNoEntryStartsSignedOut() {
        XCTAssertEqual(AuthSession(vault: FakeVault().vault).lock, .signedOut)
    }

    func testLockedIsStillSignedInSoNothingIsErased() {
        XCTAssertTrue(AuthSession(vault: FakeVault(stored: Self.stored(locked: true)).vault).isSignedIn)
    }

    func testUnlockReadsWithTheFaceIDReasonAndUnlocks() async {
        let fake = FakeVault(stored: Self.stored(locked: true))
        let session = AuthSession(vault: fake.vault)
        await session.unlock()
        XCTAssertEqual(session.lock, .unlocked)
        XCTAssertNotNil(session.authState)
        XCTAssertEqual(fake.reads.first?.localizedReason, "Unlock My Money Recap with Face ID.")
        XCTAssertTrue(fake.writes.isEmpty, "an entry already under the lock is not rewritten")
        XCTAssertEqual(fake.confirms, 0, "the Keychain read itself asked")
    }

    func testACancelledUnlockStaysLockedAndKeepsTheEntry() async {
        let fake = FakeVault(stored: Self.stored(locked: true))
        let session = AuthSession(vault: fake.vault)
        fake.stored = nil
        await session.unlock()
        XCTAssertEqual(session.lock, .locked)
        XCTAssertEqual(fake.deletes, 0)
    }

    func testAnEntryGoneWithThePasscodeSignsOutWithTheCopy() async {
        let fake = FakeVault(stored: Self.stored(locked: true))
        let session = AuthSession(vault: fake.vault)
        fake.stored = nil
        fake.present = false
        fake.passcode = false
        await session.unlock()
        XCTAssertEqual(session.lock, .signedOut)
        XCTAssertEqual(session.signInMessage, "Set a passcode in iOS Settings to use My Money Recap.")
    }

    func testAnEntryFromBeforeTheLockIsRewrittenUnderIt() async throws {
        let fake = FakeVault(stored: Self.stored(locked: false))
        let session = AuthSession(vault: fake.vault)
        await session.unlock()
        XCTAssertEqual(session.lock, .unlocked)
        XCTAssertEqual(fake.writes.count, 1)
        XCTAssertEqual(fake.confirms, 1)
        XCTAssertNotNil(try NSKeyedUnarchiver.unarchivedObject(
            ofClass: OIDAuthState.self,
            from: XCTUnwrap(fake.writes.first)
        ))
    }

    func testNoPasscodeStoresNothingAndSaysSo() {
        let fake = FakeVault()
        fake.passcode = false
        let session = AuthSession(authState: Self.state(), vault: fake.vault)
        session.persist()
        XCTAssertTrue(fake.writes.isEmpty)
        XCTAssertEqual(session.lock, .signedOut)
        XCTAssertEqual(session.signInMessage, "Set a passcode in iOS Settings to use My Money Recap.")
    }

    func testReturningAfterFiveMinutesRelocksAndKeepsTheEntry() async {
        let fake = FakeVault(stored: Self.stored(locked: true))
        let session = AuthSession(vault: fake.vault)
        await session.unlock()
        session.returned(after: 300)
        XCTAssertEqual(session.lock, .unlocked, "five minutes exactly is still within the grace")
        session.returned(after: 301)
        XCTAssertEqual(session.lock, .locked)
        XCTAssertNil(session.authState)
        XCTAssertTrue(session.isSignedIn)
        XCTAssertEqual(fake.deletes, 0)
    }

    func testAFailedKeychainWriteSignsOutRatherThanStayOpenUnlocked() {
        let fake = FakeVault()
        fake.writable = false
        let session = AuthSession(authState: Self.state(), vault: fake.vault)
        session.persist()
        XCTAssertEqual(session.lock, .signedOut, "an unstored session could never relock")
    }

    func testAnEntryFromBeforeTheLockNeedsPresenceFirst() async {
        let fake = FakeVault(stored: Self.stored(locked: false))
        fake.confirmed = false
        let session = AuthSession(vault: fake.vault)
        await session.unlock()
        XCTAssertEqual(session.lock, .locked, "an entry with no lock of its own must not open without Face ID")
        XCTAssertNil(session.authState)
        XCTAssertTrue(fake.writes.isEmpty)
    }

    func testSignOutDeletesTheEntry() async {
        let fake = FakeVault(stored: Self.stored(locked: true))
        let session = AuthSession(vault: fake.vault)
        await session.unlock()
        session.signOut()
        XCTAssertEqual(session.lock, .signedOut)
        XCTAssertEqual(fake.deletes, 1)
    }

    func testTheRealKeychainStoresUnderThePasscodeLock() throws {
        Keychain.delete()
        defer { Keychain.delete() }
        Keychain.write(Data("one".utf8))
        XCTAssertTrue(Keychain.exists())
        let stored = try XCTUnwrap(Keychain.read(context: LAContext()))
        XCTAssertEqual(stored.data, Data("one".utf8))
        XCTAssertTrue(stored.locked, "written with the passcode-required protection class")
    }

    static func state() -> OIDAuthState {
        let endpoint = URL(string: "https://realm.invalid/token")!
        let configuration = OIDServiceConfiguration(authorizationEndpoint: endpoint, tokenEndpoint: endpoint)
        let request = OIDAuthorizationRequest(
            configuration: configuration, clientId: AuthConfig.clientID, clientSecret: nil, scopes: AuthConfig.scopes,
            redirectURL: AuthConfig.redirectURL, responseType: OIDResponseTypeCode, additionalParameters: nil
        )
        let authorization = OIDAuthorizationResponse(request: request, parameters: ["code": "code" as NSString])
        let token = OIDTokenResponse(request: authorization.tokenExchangeRequest()!, parameters: [
            "access_token": "a" as NSString, "expires_in": 300 as NSNumber,
            "refresh_token": "r" as NSString, "token_type": "Bearer" as NSString,
        ])
        return OIDAuthState(authorizationResponse: authorization, tokenResponse: token, registrationResponse: nil)
    }

    static func stored(locked: Bool) -> Keychain.Stored {
        let data = (try? NSKeyedArchiver.archivedData(withRootObject: state(), requiringSecureCoding: true)) ?? Data()
        return Keychain.Stored(data: data, locked: locked)
    }
}
