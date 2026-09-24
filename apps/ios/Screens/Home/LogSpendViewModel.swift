import Connect
import Foundation
import MoneyRecapContract

/// B2 to B4 Log a spend (PRD §6.9): an amount in the Currency's decimals, a Category, a date in the last 60 days,
/// today by default (D34), and a note. Save answers the fresh Home, or nil after a failure, keeping every value.
@Observable
@MainActor
final class LogSpendViewModel: Identifiable {
    typealias Send = (MoneyRecap_V1_BffServiceLogSpendRequest, Headers) async
        -> Result<MoneyRecap_V1_BffServiceLogSpendResponse, ConnectError>

    static let failure = "We couldn't save this spend. Try again."
    static let dateHelper = "You can log up to 60 days back, never ahead."
    /// PRD §6.9's note limit, counted as ledger counts it: code points of the NFC form.
    static let noteLimit = 80

    struct Values: Equatable {
        let amount: String
        let category: Icon.Name
        let date: String
        let note: String
    }

    let currency: String
    let digits: Int
    var amount = ""
    private(set) var category: Icon.Name?
    var date: Date
    private(set) var note = ""
    private(set) var isSaving = false
    private(set) var failed = false
    /// Ledger's earliest date to today, the dates it takes (PRD §8.10, D72); today alone before Home names one.
    let dates: ClosedRange<Date>
    private var attempt: SaveAttempt<Values>?
    private let calendar: Calendar
    private let locale: Locale
    private let send: Send

    init(
        currency: String,
        digits: Int,
        earliest: String,
        today: Date = .now,
        calendar: Calendar = .autoupdatingCurrent,
        locale: Locale = .autoupdatingCurrent,
        send: @escaping Send
    ) {
        self.currency = currency
        self.digits = digits
        let day = calendar.startOfDay(for: today)
        date = day
        let iso = Date.ISO8601FormatStyle(timeZone: calendar.timeZone).year().month().day()
        dates = min((try? Date(earliest, strategy: iso)) ?? day, day) ... day
        self.calendar = calendar
        self.locale = locale
        self.send = send
    }

    func choose(_ category: Icon.Name) {
        self.category = category
    }

    /// Keeps the first 80 code points of the note's NFC form, the most ledger stores.
    func type(_ note: String) {
        let scalars = note.precomposedStringWithCanonicalMapping.unicodeScalars
        self.note = String(String.UnicodeScalarView(scalars.prefix(Self.noteLimit)))
    }

    /// The Date row: "Today, 26 September", or the chosen day alone, as the device writes a day and a month.
    var dateLabel: String {
        let style = Date.FormatStyle(locale: locale, calendar: calendar, timeZone: calendar.timeZone).day().month(.wide)
        let day = date.formatted(style)
        return calendar.isDate(date, inSameDayAs: dates.upperBound) ? "Today, \(day)" : day
    }

    var canSave: Bool {
        (Decimal(string: amount) ?? 0) > 0 && category != nil && !isSaving
    }

    /// The fresh Home once saved. A retry of the same values sends the same key; changed values a new one (IOS I3).
    func save() async -> MoneyRecap_V1_HomeView? {
        guard canSave, let category else { return nil }
        let day = date.formatted(Date.ISO8601FormatStyle(timeZone: calendar.timeZone).year().month().day())
        let values = Values(
            amount: amount.hasSuffix(".") ? String(amount.dropLast()) : amount,
            category: category,
            date: day,
            note: note
        )
        var current = attempt ?? SaveAttempt(values)
        current.update(values)
        attempt = current
        isSaving = true
        defer { isSaving = false }
        let result = await send(.with {
            $0.amount = .with {
                $0.amount = values.amount
                $0.currencyCode = currency
            }
            $0.category = Self.contract(category)
            $0.date = values.date
            $0.note = values.note
        }, current.headers)
        guard case let .success(answer) = result else {
            failed = true
            return nil
        }
        failed = false
        return answer.home
    }

    /// The contract's Category for a glyph: both follow PRD D20's order.
    static func contract(_ category: Icon.Name) -> MoneyRecap_V1_Category {
        MoneyRecap_V1_Category(rawValue: (Icon.Name.allCases.firstIndex(of: category) ?? 7) + 1) ?? .other
    }
}
