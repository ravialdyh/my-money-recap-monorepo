import Foundation
@preconcurrency import LocalAuthentication

struct Vault {
    var exists: () -> Bool
    var read: (LAContext) async -> Keychain.Read
    var write: (Data) -> Bool
    var delete: () -> Void
    var passcodeSet: () -> Bool

    static let live = Vault(
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
