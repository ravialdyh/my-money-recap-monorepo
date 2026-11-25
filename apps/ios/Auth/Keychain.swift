import Foundation
import LocalAuthentication
import Security

enum Keychain {
    enum Read: Equatable {
        case found(Data)
        case missing
        case refused
    }

    private static let everyEntry: [CFString: Any] = [
        kSecClass: kSecClassGenericPassword,
        kSecAttrService: "com.mymoneyrecap.auth",
    ]

    private static let query = everyEntry.merging([kSecAttrAccount: "session"]) { $1 }

    static func exists() -> Bool {
        let silent = LAContext()
        silent.interactionNotAllowed = true
        var q = query
        q[kSecUseAuthenticationContext] = silent
        q[kSecReturnAttributes] = true
        let status = SecItemCopyMatching(q as CFDictionary, nil)
        return status == errSecSuccess || status == errSecInteractionNotAllowed
    }

    static func read(context: LAContext) -> Read {
        var q = query
        q[kSecReturnData] = true
        q[kSecMatchLimit] = kSecMatchLimitOne
        q[kSecUseAuthenticationContext] = context
        var out: AnyObject?
        switch SecItemCopyMatching(q as CFDictionary, &out) {
        case errSecSuccess: return (out as? Data).map(Read.found) ?? .refused
        case errSecItemNotFound: return .missing
        default: return .refused
        }
    }

    @discardableResult
    static func write(_ data: Data) -> Bool {
        guard let lock = SecAccessControlCreateWithFlags(
            nil, kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly, .userPresence, nil
        ) else { return false }
        delete()
        var q = query
        q[kSecValueData] = data
        q[kSecAttrAccessControl] = lock
        return SecItemAdd(q as CFDictionary, nil) == errSecSuccess
    }

    static func delete() {
        SecItemDelete(everyEntry as CFDictionary)
    }

    static func passcodeSet() -> Bool {
        LAContext().canEvaluatePolicy(.deviceOwnerAuthentication, error: nil)
    }
}
