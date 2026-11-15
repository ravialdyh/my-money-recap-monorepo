import Connect
import Foundation

struct SaveAttempt<Values: Equatable> {
    private(set) var values: Values
    private(set) var key = UUID().uuidString

    init(_ values: Values) {
        self.values = values
    }

    mutating func update(_ values: Values) {
        guard values != self.values else { return }
        self = SaveAttempt(values)
    }

    var headers: Headers {
        ["Idempotency-Key": [key]]
    }
}

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

func unavailable<Answer>(_: some Any) async -> Result<Answer, ConnectError> {
    .failure(ConnectError(code: .unavailable, message: nil))
}

func unavailable<Answer>(_: some Any, _: Headers) async -> Result<Answer, ConnectError> {
    .failure(ConnectError(code: .unavailable, message: nil))
}
