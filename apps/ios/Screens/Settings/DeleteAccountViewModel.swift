import Connect
import Foundation
import MoneyRecapContract

/// D10 and D11 (PRD §6.16, D39, D60): `Delete my account` works only on exactly `DELETE`, asks the API to stamp the
/// request, then signs out as Log out does; a failure keeps the Member in, and a retry repeats its key (A14).
@Observable
@MainActor
final class DeleteAccountViewModel: Identifiable {
    typealias Request = (Headers) async -> Result<MoneyRecap_V1_BffServiceRequestDeletionResponse, ConnectError>

    static let failedCopy = "We couldn't delete your account. Try again."

    /// D11, pushed from D10's button.
    var confirming = false
    var typed = ""
    private(set) var deleting = false
    private(set) var failed = false
    private let request: Request
    private let signOut: () async -> Void
    private var attempt: SaveAttempt<Bool>?

    init(request: @escaping Request, signOut: @escaping () async -> Void) {
        self.request = request
        self.signOut = signOut
    }

    var canDelete: Bool {
        typed == "DELETE" && !deleting
    }

    func delete() async {
        guard canDelete else { return }
        let current = attempt ?? SaveAttempt(true)
        attempt = current
        deleting = true
        failed = false
        if case .success = await request(current.headers) {
            await signOut()
        } else {
            failed = true
        }
        deleting = false
    }
}
