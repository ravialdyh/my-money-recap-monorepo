import Connect
import Foundation

/// D11b Restore account (SCREEN-STATES, PRD §6.18, D60): a deletion asked for under 24 hours ago opens it straight
/// after the launch call, before the setup check and before Home. "Keep my account" clears the request (A14).
@Observable
@MainActor
final class RestoreAccountViewModel {
    typealias Keep = (Headers) async -> SaveResult

    /// One line for every failure (SCREEN-STATES D11b), in the sheet, with both buttons live.
    static let failedCopy = "We couldn't keep your account. Try again."
    nonisolated static let window: TimeInterval = 24 * 60 * 60

    /// True for a request under 24 hours old, the window members still clears (D39).
    nonisolated static func shows(asked: Date?, now: Date) -> Bool {
        guard let asked else { return false }
        return now.timeIntervalSince(asked) < window
    }

    let asked: Date
    private(set) var isKeeping = false
    private(set) var failed = false
    /// One key for every try, so a keep whose answer was lost is not written twice.
    private let attempt = SaveAttempt(true)
    private let now: () -> Date
    private let calendar: Calendar
    private let time: Date.FormatStyle
    private let keepAccount: Keep

    init(
        asked: Date,
        now: @escaping () -> Date = Date.init,
        calendar: Calendar = .autoupdatingCurrent,
        locale: Locale = Locale(identifier: "en_GB"),
        keep: @escaping Keep
    ) {
        self.asked = asked
        self.now = now
        self.calendar = calendar
        time = Date.FormatStyle(locale: locale, calendar: calendar, timeZone: calendar.timeZone)
            .hour(.twoDigits(amPM: .omitted)).minute(.twoDigits)
        keepAccount = keep
    }

    /// "You asked us to delete this account today at 09:12. Nothing has gone yet."
    var askedLine: String {
        "You asked us to delete this account \(day(asked)) at \(asked.formatted(time)). Nothing has gone yet."
    }

    /// The canvas's second line, with the time the purge is due, 24 hours after asking.
    var loseLine: String {
        let due = asked.addingTimeInterval(Self.window)
        return "Keep it and every spend, every recap and every setting stays exactly as it was. "
            + "Continue and it all goes at \(due.formatted(time)) \(day(due)), for good."
    }

    /// True once members cleared the request; otherwise `failedCopy` shows and both buttons stay live.
    func keep() async -> Bool {
        guard !isKeeping else { return false }
        isKeeping = true
        defer { isKeeping = false }
        let kept = await keepAccount(attempt.headers) == .saved
        failed = !kept
        return kept
    }

    private func day(_ date: Date) -> String {
        if calendar.isDate(date, inSameDayAs: now()) {
            return "today"
        }
        return date < now() ? "yesterday" : "tomorrow"
    }
}
