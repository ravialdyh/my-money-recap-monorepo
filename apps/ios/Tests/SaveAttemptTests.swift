import Connect
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

    /// No connection is what Connect reports for an unreachable BFF or a dropped network; anything else is a refusal.
    func testNoConnectionIsUnavailableOrATimeout() {
        XCTAssertEqual(SaveResult(Result<Int, ConnectError>.success(1)), .saved)
        for code in [Code.unavailable, .deadlineExceeded] {
            XCTAssertEqual(
                SaveResult(Result<Int, ConnectError>.failure(ConnectError(code: code, message: nil))),
                .offline
            )
        }
        for code in [Code.invalidArgument, .failedPrecondition, .resourceExhausted, .internalError, .unauthenticated] {
            XCTAssertEqual(
                SaveResult(Result<Int, ConnectError>.failure(ConnectError(code: code, message: nil))),
                .failed
            )
        }
    }
}
