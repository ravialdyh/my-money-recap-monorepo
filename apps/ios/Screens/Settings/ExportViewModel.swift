import Connect
import Foundation
import MoneyRecapContract

@Observable
@MainActor
final class ExportViewModel: Identifiable {
    typealias Send = (Headers) async -> Result<MoneyRecap_V1_BffServiceRequestExportResponse, ConnectError>

    enum State: Equatable {
        case ready, sending, sent, refused, failed
    }

    let email: String
    private(set) var state = State.ready
    private let send: Send
    private let wait: (Duration) async -> Void
    private var attempt: SaveAttempt<Bool>?

    init(
        email: String,
        send: @escaping Send,
        wait: @escaping (Duration) async -> Void = { try? await Task.sleep(for: $0) }
    ) {
        self.email = email
        self.send = send
        self.wait = wait
    }

    var helper: String {
        let what = "Every spend you've ever logged, with its date, category, amount and note."
        return state == .refused ? what : "It goes to \(email). \(what)"
    }

    var error: String? {
        switch state {
        case .refused: "We can't reach \(email)."
        case .failed: "We couldn't send it. Try again."
        default: nil
        }
    }

    var button: String {
        state == .sent ? "Sent ✓" : "Send it"
    }

    var canSend: Bool {
        state == .ready || state == .failed
    }

    func commit() async {
        guard canSend else { return }
        let current = attempt ?? SaveAttempt(true)
        attempt = current
        state = .sending
        switch await send(current.headers) {
        case .success:
            attempt = nil
            state = .sent
            await wait(.seconds(60))
            if state == .sent {
                state = .ready
            }
        case let .failure(error):
            let bounced: [MoneyRecap_V1_EmailBounced] = error.unpackedDetails()
            state = bounced.isEmpty ? .failed : .refused
        }
    }
}
