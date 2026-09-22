import XCTest
@testable import MyMoneyRecap

final class SaveAttemptTests: XCTestCase {
    func testARetryKeepsTheKeyAndChangedValuesGetANewOne() {
        var attempt = SaveAttempt("Asia/Jakarta")
        let first = attempt.key
        XCTAssertEqual(attempt.headers, ["Idempotency-Key": [first]])
        attempt.update("Asia/Jakarta")
        XCTAssertEqual(attempt.key, first, "a retry of the same values is the same write")
        attempt.update("Europe/London")
        XCTAssertNotEqual(attempt.key, first, "changed values are a new write")
        XCTAssertEqual(attempt.values, "Europe/London")
    }
}
