import Foundation

/// A2 Enter email (SCREEN-STATES, IOS I6): Send stays disabled until the address is syntactically an
/// address, and the coral helper appears only after the field loses focus or Send is pressed, so the
/// Member is never corrected mid-word (PRD §6.2).
@Observable
@MainActor
final class EnterEmailViewModel {
    static let invalidCopy = "That doesn't look like an email address."

    var email = ""
    private(set) var isSending = false
    private(set) var failure: MagicLinkFailure?
    private var corrects = false
    private let send: (String) async -> MagicLinkFailure?

    /// `initialFailure` is the copy a refused login token left behind (PRD §6.3).
    init(send: @escaping (String) async -> MagicLinkFailure?, initialFailure: MagicLinkFailure? = nil) {
        self.send = send
        failure = initialFailure
    }

    /// `something@domain.tld`, the design's reading of PRD §6.2's syntax check: full RFC 5322 is wider than
    /// any check worth writing, and this one only keeps a typo from spending one of five hourly links.
    var isValid: Bool {
        email.trimmingCharacters(in: .whitespaces).wholeMatch(of: /[^@\s]+@[^@\s.]+(\.[^@\s.]+)+/) != nil
    }

    /// Five links in an hour is refused before any email is sent, so Send stays disabled (A3 rate limited).
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

    /// Answers the address the link went to, which Check your inbox names, or nil when nothing was sent.
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
