import XCTest
@testable import MyMoneyRecap

/// `Money` prints through the platform formatter (IOS I7); a fixed locale keeps the runner's locale out of the copy.
final class MoneyTests: XCTestCase {
    private let gb = Locale(identifier: "en_GB")

    func testWholeAmountsPrintWithoutAZeroFraction() {
        XCTAssertEqual(Money.format("0", currency: "GBP", locale: gb), "£0")
        XCTAssertEqual(Money.format("2400", currency: "GBP", locale: gb), "£2,400")
    }

    func testFractionsKeepTheCurrencysDecimals() {
        XCTAssertEqual(Money.format("12.40", currency: "GBP", locale: gb), "£12.40")
        XCTAssertEqual(Money.format("15000000", currency: "IDR", locale: Locale(identifier: "id_ID")), "Rp15.000.000")
    }
}
