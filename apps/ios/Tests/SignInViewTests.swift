import XCTest
@testable import MyMoneyRecap

@MainActor
final class SignInViewTests: XCTestCase {
    func testSignInShowsTheNoPasscodeCopy() {
        let fake = FakeVault()
        let plain = picture(SignInView(session: AuthSession(vault: fake.vault)))
        fake.passcode = false
        let noPasscode = picture(SignInView(session: AuthSession(vault: fake.vault)))
        XCTAssertNotNil(noPasscode)
        XCTAssertNotEqual(plain, noPasscode, "the copy shows before any sign-in is tried")
    }
}
