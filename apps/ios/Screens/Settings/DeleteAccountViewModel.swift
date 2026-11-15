import Connect
import Foundation
import MoneyRecapContract

@Observable
@MainActor
final class DeleteAccountViewModel: Identifiable {
    typealias Request = (Headers) async -> Result<MoneyRecap_V1_BffServiceRequestDeletionResponse, ConnectError>

    static let failedCopy = "We couldn't delete your account. Try again."

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
