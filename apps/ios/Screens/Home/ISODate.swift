import Foundation

/// The server's "YYYY-MM-DD" and "YYYY-MM" strings are calendar dates, not instants, so they are read and printed
/// in one fixed zone; otherwise a phone west of Greenwich would show 30 September for `2026-10-01` (IOS I7).
enum ISODate {
    static func style(locale: Locale) -> Date.FormatStyle {
        Date.FormatStyle(locale: locale, timeZone: .gmt)
    }

    static func day(_ string: String) -> Date? {
        try? Date(string, strategy: .iso8601.year().month().day())
    }

    static func month(_ string: String) -> Date? {
        day(string + "-01")
    }
}
