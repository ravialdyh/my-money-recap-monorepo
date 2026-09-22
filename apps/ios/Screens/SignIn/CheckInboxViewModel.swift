import Foundation

/// A3 Check your inbox (SCREEN-STATES, IOS I6): it names the address the link went to, says the link works
/// for 15 minutes, and wakes Resend after 30 seconds behind a live m:ss countdown (PRD §6.3).
@Observable
@MainActor
final class CheckInboxViewModel {
    static let resendDelay = 30
    static let title = "Check your inbox."

    let email: String
    private(set) var secondsLeft = resendDelay
    private(set) var isSending = false
    private var isCountingDown = false
    private let send: (String) async -> MagicLinkFailure?
    private let tick: () async -> Void

    /// `tick` is one second of waiting; the tests pass one that returns at once.
    init(
        email: String,
        send: @escaping (String) async -> MagicLinkFailure?,
        tick: @escaping () async -> Void = { try? await Task.sleep(for: .seconds(1)) }
    ) {
        self.email = email
        self.send = send
        self.tick = tick
    }

    var subCopy: String {
        "We sent a link to \(email). It works for 15 minutes."
    }

    var resendTitle: String {
        secondsLeft > 0 ? "Resend in \(secondsLeft / 60):\(String(format: "%02d", secondsLeft % 60))" : "Resend"
    }

    var canResend: Bool {
        secondsLeft == 0 && !isSending
    }

    /// Runs to zero once however often the screen appears: a second run while one is in flight, or after
    /// the countdown is spent, would resend early or tick past zero.
    func countDown() async {
        guard !isCountingDown, secondsLeft > 0 else { return }
        isCountingDown = true
        defer { isCountingDown = false }
        while secondsLeft > 0 {
            await tick()
            secondsLeft -= 1
        }
    }

    /// A link that went out again restarts the countdown; a refused one leaves it spent.
    func resend() async {
        guard canResend else { return }
        isSending = true
        defer { isSending = false }
        guard await send(email) == nil else { return }
        secondsLeft = Self.resendDelay
    }
}
