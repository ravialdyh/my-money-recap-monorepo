import XCTest
@testable import MyMoneyRecap

@MainActor
final class UnlockViewTests: XCTestCase {
    func testTheUnlockScreenAsksForFaceIDAsItAppears() {
        let fake = FakeVault(stored: SessionLockTests.stored(locked: true))
        fake.stored = nil
        let session = AuthSession(vault: fake.vault)
        XCTAssertNotNil(picture(UnlockView(session: session)))
        XCTAssertFalse(fake.reads.isEmpty, "the prompt comes up without a tap")
        XCTAssertEqual(session.lock, .locked, "a cancelled prompt leaves the Unlock button")
    }
}
