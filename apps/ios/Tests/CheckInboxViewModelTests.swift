import XCTest
@testable import MyMoneyRecap

/// A3 Check your inbox (SCREEN-STATES): the named address, the live m:ss countdown that reaches zero once,
/// and Resend restarting it (PRD §6.3).
@MainActor
final class CheckInboxViewModelTests: XCTestCase {
    func testItNamesTheAddressAndThe15Minutes() {
        let model = CheckInboxViewModel(email: "maya@example.com", send: { _ in nil })
        XCTAssertEqual(model.subCopy, "We sent a link to maya@example.com. It works for 15 minutes.")
        XCTAssertEqual(model.resendTitle, "Resend in 0:30")
        XCTAssertFalse(model.canResend)
    }

    func testTheCountdownReachesZeroExactlyOnce() async {
        var ticks = 0
        var zeroes = 0
        var model: CheckInboxViewModel!
        model = CheckInboxViewModel(email: "maya@example.com", send: { _ in nil }, tick: {
            ticks += 1
            if model.secondsLeft == 0 {
                zeroes += 1
            }
        })
        await model.countDown()
        await model.countDown()
        XCTAssertEqual(ticks, CheckInboxViewModel.resendDelay, "a second appearance does not tick again")
        XCTAssertEqual(zeroes, 0, "the loop stops at zero rather than counting past it")
        XCTAssertEqual(model.secondsLeft, 0)
        XCTAssertEqual(model.resendTitle, "Resend")
        XCTAssertTrue(model.canResend)
    }

    func testTheCountdownIsDrawnAsMinutesAndSeconds() async {
        let model = CheckInboxViewModel(email: "maya@example.com", send: { _ in nil }, tick: {})
        var titles: [String] = [model.resendTitle]
        await model.countDown()
        titles.append(model.resendTitle)
        XCTAssertEqual(titles, ["Resend in 0:30", "Resend"])
    }

    func testResendSendsAgainAndRestartsTheCountdown() async {
        var sent: [String] = []
        let model = CheckInboxViewModel(email: "maya@example.com", send: {
            sent.append($0)
            return nil
        }, tick: {})
        var resent = await model.resend()
        XCTAssertFalse(resent, "Resend sleeps for the first 30 seconds")
        XCTAssertEqual(sent, [])
        await model.countDown()
        resent = await model.resend()
        XCTAssertTrue(resent)
        XCTAssertEqual(sent, ["maya@example.com"], "the address is not retyped")
        XCTAssertEqual(model.secondsLeft, CheckInboxViewModel.resendDelay)
        XCTAssertFalse(model.canResend)
    }

    func testEachResendFailureKeepsItsOwnCopyAndTheCountdownSpent() async {
        for failure in [MagicLinkFailure.rateLimited, .generic] {
            let model = CheckInboxViewModel(email: "maya@example.com", send: { _ in failure }, tick: {})
            await model.countDown()
            let resent = await model.resend()
            XCTAssertFalse(resent)
            XCTAssertEqual(model.failure, failure)
            XCTAssertEqual(model.failure?.copy, failure.copy)
            XCTAssertEqual(model.secondsLeft, 0, "a link that never went out starts no new 30 seconds")
        }
    }
}
