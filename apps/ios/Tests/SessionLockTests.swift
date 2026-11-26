import AppAuth
import LocalAuthentication
import XCTest
@testable import MyMoneyRecap

final class FakeVault {
    var stored: Keychain.Read
    var present: Bool
    var passcode = true
    var reads: [LAContext] = []
    var writes: [Data] = []
    var deletes = 0
    var writable = true

    init(stored: Keychain.Read = .missing) {
        self.stored = stored
        present = stored != .missing
    }

    var vault: Vault {
        Vault(
            exists: { self.present },
            read: { context in
                self.reads.append(context)
                await Task.yield()
                return self.stored
            },
            write: {
                self.writes.append($0)
                return self.writable
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
        let fake = FakeVault(stored: Self.found())
        let session = AuthSession(vault: fake.vault)
        XCTAssertEqual(session.lock, .locked)
        XCTAssertNil(session.authState)
        XCTAssertTrue(fake.reads.isEmpty, "a read would show the Face ID prompt during init")
    }

    func testNoEntryStartsSignedOut() {
        XCTAssertEqual(AuthSession(vault: FakeVault().vault).lock, .signedOut)
    }

    func testAPhoneWithNoPasscodeSaysSoAsTheAppOpens() {
        let fake = FakeVault()
        fake.passcode = false
        XCTAssertEqual(AuthSession(vault: fake.vault).signInMessage, AuthSession.noPasscodeCopy)
    }

    func testLockedIsStillSignedInSoNothingIsErased() {
        XCTAssertTrue(AuthSession(vault: FakeVault(stored: Self.found()).vault).isSignedIn)
    }

    func testUnlockReadsWithTheUnlockReasonAndUnlocks() async {
        let fake = FakeVault(stored: Self.found())
        let session = AuthSession(vault: fake.vault)
        await session.unlock()
        XCTAssertEqual(session.lock, .unlocked)
        XCTAssertNotNil(session.authState)
        XCTAssertEqual(fake.reads.first?.localizedReason, "Unlock My Money Recap.")
        XCTAssertTrue(fake.writes.isEmpty, "unlocking rewrites nothing")
    }

    func testACancelledUnlockStaysLockedAndKeepsTheEntry() async {
        let fake = FakeVault(stored: Self.found())
        let session = AuthSession(vault: fake.vault)
        fake.stored = .refused
        await session.unlock()
        XCTAssertEqual(session.lock, .locked)
        XCTAssertEqual(fake.deletes, 0)
    }

    func testTwoUnlocksAtOnceAskOnce() async {
        let fake = FakeVault(stored: Self.found())
        let session = AuthSession(vault: fake.vault)
        async let first: Void = session.unlock()
        async let second: Void = session.unlock()
        _ = await (first, second)
        XCTAssertEqual(fake.reads.count, 1, "a second prompt would stack on the first")
        XCTAssertEqual(session.lock, .unlocked)
    }

    func testAnEntryGoneWithThePasscodeSignsOutWithTheCopy() async {
        let fake = FakeVault(stored: Self.found())
        let session = AuthSession(vault: fake.vault)
        fake.stored = .missing
        fake.passcode = false
        await session.unlock()
        XCTAssertEqual(session.lock, .signedOut)
        XCTAssertEqual(session.signInMessage, AuthSession.noPasscodeCopy)
    }

    func testAMagicLinkOpensOnlyFromSignIn() async throws {
        let link = try XCTUnwrap(URL(string: "mymoneyrecap://magic"))
        let locked = AuthSession(vault: FakeVault(stored: Self.found()).vault)
        await locked.openMagicLink(link)
        XCTAssertEqual(locked.lock, .locked, "a link must not get past Unlock")
        XCTAssertNil(locked.magicLinkFailure, "the link was not even looked at")

        let fake = FakeVault()
        let unlocked = AuthSession(authState: Self.state(), vault: fake.vault)
        await unlocked.openMagicLink(link)
        XCTAssertNil(unlocked.magicLinkFailure, "another account's link would keep this one's stored screens")
        XCTAssertTrue(fake.writes.isEmpty)
    }

    func testNoPasscodeStoresNothingAndSaysSo() {
        let fake = FakeVault()
        fake.passcode = false
        let session = AuthSession(authState: Self.state(), vault: fake.vault)
        let before = session.signOuts
        session.persist(Self.state())
        XCTAssertTrue(fake.writes.isEmpty)
        XCTAssertEqual(session.lock, .signedOut)
        XCTAssertEqual(session.signInMessage, AuthSession.noPasscodeCopy)
        XCTAssertEqual(session.signOuts, before + 1, "Sign in starts fresh, so the copy shows")
    }

    func testReturningAfterMoreThanFiveMinutesRelocksAndKeepsTheEntry() async {
        let fake = FakeVault(stored: Self.found())
        let session = AuthSession(vault: fake.vault)
        await session.unlock()
        let start = ContinuousClock.now
        session.left(at: start)
        session.returned(at: start + .seconds(300))
        XCTAssertEqual(session.lock, .unlocked, "five minutes exactly is still within the grace")
        session.left(at: start)
        session.returned(at: start + .milliseconds(300_500))
        XCTAssertEqual(session.lock, .locked, "half a second over is over")
        XCTAssertNil(session.authState)
        XCTAssertTrue(session.isSignedIn)
        XCTAssertEqual(fake.deletes, 0)
    }

    func testABounceWhileAwayKeepsTheFirstLeavingTime() async {
        let session = AuthSession(vault: FakeVault(stored: Self.found()).vault)
        await session.unlock()
        let start = ContinuousClock.now
        session.left(at: start)
        session.left(at: start + .seconds(270))
        session.returned(at: start + .seconds(301))
        XCTAssertEqual(session.lock, .locked)
    }

    func testRemovingThePasscodeWhileAwaySignsOutOnReturn() async {
        let fake = FakeVault(stored: Self.found())
        let session = AuthSession(vault: fake.vault)
        await session.unlock()
        let start = ContinuousClock.now
        session.left(at: start)
        fake.passcode = false
        session.returned(at: start + .seconds(10))
        XCTAssertEqual(session.lock, .signedOut)
        XCTAssertEqual(session.signInMessage, AuthSession.noPasscodeCopy)
    }

    func testAFailedKeychainWriteSignsOutRatherThanStayOpenUnlocked() {
        let fake = FakeVault()
        fake.writable = false
        let session = AuthSession(authState: Self.state(), vault: fake.vault)
        session.persist(Self.state())
        XCTAssertEqual(session.lock, .signedOut, "an unstored session could never relock")
    }

    func testSignOutDeletesTheEntry() async {
        let fake = FakeVault(stored: Self.found())
        let session = AuthSession(vault: fake.vault)
        await session.unlock()
        session.signOut()
        XCTAssertEqual(session.lock, .signedOut)
        XCTAssertEqual(fake.deletes, 1)
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

    static func found() -> Keychain.Read {
        .found((try? NSKeyedArchiver.archivedData(withRootObject: state(), requiringSecureCoding: true)) ?? Data())
    }
}
