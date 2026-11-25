import Foundation
import LocalAuthentication
import Security

enum Keychain {
    struct Stored {
        let data: Data
        let locked: Bool
    }

    private static let query: [CFString: Any] = [
        kSecClass: kSecClassGenericPassword,
        kSecAttrService: "com.mymoneyrecap.auth",
        kSecAttrAccount: "authState",
    ]

    static func exists() -> Bool {
        let silent = LAContext()
        silent.interactionNotAllowed = true
        var q = query
        q[kSecUseAuthenticationContext] = silent
        q[kSecReturnAttributes] = true
        let status = SecItemCopyMatching(q as CFDictionary, nil)
        return status == errSecSuccess || status == errSecInteractionNotAllowed
    }

    static func read(context: LAContext? = nil) -> Stored? {
        var q = query
        q[kSecReturnData] = true
        q[kSecReturnAttributes] = true
        q[kSecMatchLimit] = kSecMatchLimitOne
        if let context {
            q[kSecUseAuthenticationContext] = context
        }
        var out: AnyObject?
        guard SecItemCopyMatching(q as CFDictionary, &out) == errSecSuccess,
              let item = out as? [CFString: Any], let data = item[kSecValueData] as? Data
        else { return nil }
        let protection = item[kSecAttrAccessible] as? String
        return Stored(data: data, locked: protection == kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly as String)
    }

    static func read() -> Data? {
        read(context: nil)?.data
    }

    @discardableResult
    static func write(_ data: Data) -> Bool {
        SecItemDelete(query as CFDictionary)
        guard let lock = SecAccessControlCreateWithFlags(
            nil, kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly, .userPresence, nil
        ) else { return false }
        var q = query
        q[kSecValueData] = data
        q[kSecAttrAccessControl] = lock
        return SecItemAdd(q as CFDictionary, nil) == errSecSuccess
    }

    static func delete() {
        SecItemDelete(query as CFDictionary)
    }

    static func confirmPresence(_ context: LAContext) async -> Bool {
        await (try? context.evaluatePolicy(.deviceOwnerAuthentication, localizedReason: context.localizedReason)) ??
            false
    }

    static func passcodeSet() -> Bool {
        LAContext().canEvaluatePolicy(.deviceOwnerAuthentication, error: nil)
    }
}
