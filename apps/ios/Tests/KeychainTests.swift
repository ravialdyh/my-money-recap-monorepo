import XCTest
@testable import MyMoneyRecap

final class KeychainTests: XCTestCase {
    func testRoundTripAndDelete() {
        Keychain.delete()
        XCTAssertNil(Keychain.read())
        Keychain.write(Data("first".utf8))
        Keychain.write(Data("second".utf8))
        XCTAssertEqual(Keychain.read(), Data("second".utf8), "write replaces the single item")
        Keychain.delete()
        XCTAssertNil(Keychain.read())
    }
}
