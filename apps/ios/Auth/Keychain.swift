import Foundation
import Security

/// One generic-password item, `kSecAttrAccessibleWhenUnlockedThisDeviceOnly` (PRD §9, IOS I2).
/// The `LAContext` gate and the 30-day rule are later slices.
enum Keychain {
    private static let query: [CFString: Any] = [
        kSecClass: kSecClassGenericPassword,
        kSecAttrService: "com.mymoneyrecap.auth",
        kSecAttrAccount: "authState"
    ]

    static func read() -> Data? {
        var q = query
        q[kSecReturnData] = true
        q[kSecMatchLimit] = kSecMatchLimitOne
        var out: AnyObject?
        return SecItemCopyMatching(q as CFDictionary, &out) == errSecSuccess ? out as? Data : nil
    }

    static func write(_ data: Data) {
        SecItemDelete(query as CFDictionary)
        var q = query
        q[kSecValueData] = data
        q[kSecAttrAccessible] = kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        SecItemAdd(q as CFDictionary, nil)
    }

    static func delete() {
        SecItemDelete(query as CFDictionary)
    }
}
