import Foundation
import LocalAuthentication

struct Vault {
    var exists: () -> Bool
    var read: (LAContext) async -> Keychain.Stored?
    var write: (Data) -> Void
    var delete: () -> Void
    var passcodeSet: () -> Bool

    static let keychain = Vault(
        exists: Keychain.exists,
        read: { context in
            await withCheckedContinuation { done in
                DispatchQueue.global(qos: .userInitiated).async {
                    done.resume(returning: Keychain.read(context: context))
                }
            }
        },
        write: Keychain.write,
        delete: Keychain.delete,
        passcodeSet: Keychain.passcodeSet
    )
}
