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
