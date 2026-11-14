import Connect
import Foundation

/// One save attempt's `Idempotency-Key` (ARCHITECTURE A14, IOS I3): every retry of the same values sends the same
/// key, so the owning service writes once, and changed values start a new attempt with a new key.
struct SaveAttempt<Values: Equatable> {
    private(set) var values: Values
    private(set) var key = UUID().uuidString

    init(_ values: Values) {
        self.values = values
    }

    /// Keeps the key while `values` are the ones already being saved.
    mutating func update(_ values: Values) {
        guard values != self.values else { return }
        self = SaveAttempt(values)
    }

    /// The header the BFF copies into the owning service's request (`shared.proto`).
    var headers: Headers {
        ["Idempotency-Key": [key]]
    }
}

/// How one save ended: `offline` when the BFF could not be reached, which Connect reports as `unavailable` or, for a
/// dropped network, `deadline_exceeded` (spec #92: the step says so and its Next tries again), `failed` otherwise.
enum SaveResult: Equatable {
    case saved
    case offline
    case failed

    init(_ result: Result<some Any, ConnectError>) {
        switch result {
        case .success:
            self = .saved
        case let .failure(error) where error.code == .unavailable || error.code == .deadlineExceeded:
            self = .offline
        case .failure:
            self = .failed
        }
    }
}

/// The answer of a call nobody wired, as with no signal: the default the view models give every call, so a test or a
/// picture names only the calls it drives.
func unavailable<Answer>(_: some Any) async -> Result<Answer, ConnectError> {
    .failure(ConnectError(code: .unavailable, message: nil))
}

/// The same for a call taking its request and its headers.
func unavailable<Answer>(_: some Any, _: Headers) async -> Result<Answer, ConnectError> {
    .failure(ConnectError(code: .unavailable, message: nil))
}
