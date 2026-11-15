import Foundation

@Observable
@MainActor
final class EnterEmailViewModel {
    static let invalidCopy = "That doesn't look like an email address."

    var email = ""
    private(set) var isSending = false
    private(set) var failure: MagicLinkFailure?
    private var corrects = false
    private let send: (String) async -> MagicLinkFailure?

    init(send: @escaping (String) async -> MagicLinkFailure?, initialFailure: MagicLinkFailure? = nil) {
        self.send = send
        failure = initialFailure
    }

    var isValid: Bool {
        email.trimmingCharacters(in: .whitespaces).wholeMatch(of: /[^@\s]+@[^@\s.]+(\.[^@\s.]+)+/) != nil
    }

    var canSend: Bool {
        isValid && !isSending && failure != .rateLimited
    }

    var invalidHelper: String? {
        corrects && !isValid ? Self.invalidCopy : nil
    }

    var errorMessage: String? {
        failure?.copy
    }

    func blur() {
        corrects = true
    }

    func sendLink() async -> String? {
        corrects = true
        guard isValid else { return nil }
        failure = nil
        isSending = true
        defer { isSending = false }
        let address = email.trimmingCharacters(in: .whitespaces)
        failure = await send(address)
        return failure == nil ? address : nil
    }

    func show(_ failure: MagicLinkFailure?) {
        guard let failure else { return }
        self.failure = failure
    }
}
