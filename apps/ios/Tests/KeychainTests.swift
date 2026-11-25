import LocalAuthentication
import Security
import XCTest
@testable import MyMoneyRecap

final class KeychainTests: XCTestCase {
    private let old: [CFString: Any] = [
        kSecClass: kSecClassGenericPassword,
        kSecAttrService: "com.mymoneyrecap.auth",
        kSecAttrAccount: "authState",
    ]

    override func tearDown() {
        Keychain.delete()
        SecItemDelete(old as CFDictionary)
        super.tearDown()
    }

    func testRoundTripUnderTheLockAndDelete() {
        Keychain.delete()
        XCTAssertEqual(Keychain.read(context: LAContext()), .missing)
        XCTAssertTrue(Keychain.write(Data("first".utf8)))
        XCTAssertTrue(Keychain.write(Data("second".utf8)))
        XCTAssertTrue(Keychain.exists())
        XCTAssertEqual(Keychain.read(context: LAContext()), .found(Data("second".utf8)), "write replaces the item")
        Keychain.delete()
        XCTAssertFalse(Keychain.exists())
        XCTAssertEqual(Keychain.read(context: LAContext()), .missing)
    }

    func testAnEntryFromBeforeTheLockIsIgnoredAndDroppedByTheNextWrite() {
        var item = old
        item[kSecValueData] = Data("unlocked".utf8)
        item[kSecAttrAccessible] = kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        XCTAssertEqual(SecItemAdd(item as CFDictionary, nil), errSecSuccess)
        XCTAssertFalse(Keychain.exists(), "an entry with no lock never counts as a session")
        XCTAssertTrue(Keychain.write(Data("locked".utf8)))
        XCTAssertEqual(SecItemCopyMatching(old as CFDictionary, nil), errSecItemNotFound)
    }
}

extension Keychain {
    static func read() -> Data? {
        if case let .found(data) = read(context: LAContext()) {
            data
        } else {
            nil
        }
    }
}
