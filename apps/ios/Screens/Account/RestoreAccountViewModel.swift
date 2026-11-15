import Connect
import Foundation

@Observable
@MainActor
final class RestoreAccountViewModel {
    typealias Keep = (Headers) async -> SaveResult

    static let failedCopy = "We couldn't keep your account. Try again."
    nonisolated static let window: TimeInterval = 24 * 60 * 60

    nonisolated static func shows(asked: Date?, now: Date) -> Bool {
        guard let asked else { return false }
        return now.timeIntervalSince(asked) < window
    }

    let asked: Date
    private(set) var isKeeping = false
    private(set) var failed = false
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

    var askedLine: String {
        "You asked us to delete this account \(day(asked)) at \(asked.formatted(time)). Nothing has gone yet."
    }

    var loseLine: String {
        let due = asked.addingTimeInterval(Self.window)
        return "Keep it and every spend, every recap and every setting stays exactly as it was. "
            + "Continue and it all goes at \(due.formatted(time)) \(day(due)), for good."
    }

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
