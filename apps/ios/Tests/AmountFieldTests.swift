import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// The amount field (PRD D57, D71, spec #92): it reads what is typed as the exact decimal the contract carries, in
/// the device's separators, drops a 12th whole digit or a decimal the Currency lacks, and draws the D57 line at 11.
@MainActor
final class AmountFieldTests: XCTestCase {
    private let us = Locale(identifier: "en_US")
    private let indonesia = Locale(identifier: "id_ID")

    func testWhatIsTypedIsReadAsAnExactDecimal() {
        XCTAssertEqual(AmountText.parsed("15,000,000", digits: 0, locale: us), "15000000", "grouping dropped")
        XCTAssertEqual(AmountText.parsed("15.000,5", digits: 2, locale: indonesia), "15000.5", "Indonesian separators")
        XCTAssertEqual(AmountText.parsed("007", digits: 0, locale: us), "7")
        XCTAssertEqual(AmountText.parsed(".", digits: 2, locale: us), "0.")
        XCTAssertEqual(AmountText.parsed("", digits: 2, locale: us), "")
        XCTAssertEqual(AmountText.parsed("12.", digits: 0, locale: us), "12", "a rupiah has no decimals to start")
        XCTAssertEqual(AmountText.parsed("12.345", digits: 3, locale: us), "12.345", "a dinar has three")
    }

    func testAKeystrokeBeyondTheCapOrTheCurrencysDecimalsIsDropped() {
        XCTAssertEqual(AmountText.parsed("12.345", digits: 2, locale: us), "12.34")
        XCTAssertEqual(AmountText.parsed("9,999,999,99999", digits: 0, locale: us), "99999999999", "two keys at once")
        XCTAssertEqual(AmountText.parsed("99,999,999,999", digits: 0, locale: us), "99999999999")
        XCTAssertTrue(AmountText.isFull("99999999999"), "11 digits draw the D57 line")
        XCTAssertTrue(AmountText.isFull("99999999999.5"))
        XCTAssertFalse(AmountText.isFull("9999999999.99"))
    }

    func testTheAmountIsShownAsTheLocaleGroupsIt() {
        XCTAssertEqual(AmountText.shown("15000000", locale: us), "15,000,000")
        XCTAssertEqual(AmountText.shown("15000.5", locale: indonesia), "15.000,5")
        XCTAssertEqual(AmountText.shown("12.", locale: us), "12.", "the point stays while the decimals are typed")
        XCTAssertEqual(AmountText.shown("", locale: us), "")
    }

    func testTheCapDrawsItsLineAndAnotherCurrencyItsSymbol() {
        let field = { (amount: String, currency: String) in
            AmountField(
                label: "Amount, after tax",
                currency: currency,
                digits: 0,
                amount: .constant(amount),
                locale: self.us
            )
            .padding()
            .background(.white)
        }
        XCTAssertNotEqual(picture(field("9999999999", "IDR")), picture(field("99999999999", "IDR")), "the D57 line")
        XCTAssertNotEqual(picture(field("15000000", "IDR")), picture(field("15000000", "JPY")), "IDR against ¥")
    }
}
