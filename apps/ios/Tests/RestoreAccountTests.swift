import Connect
import MoneyRecapContract
import XCTest
@testable import MyMoneyRecap

/// D11b Restore account (D60, #105): only a deletion under 24 hours old opens it, before the setup check and before
/// Home; its lines name the real days; a failed keep says so in the sheet and tries again with the same key.
@MainActor
final class RestoreAccountTests: XCTestCase {
    private var calendar: Calendar {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "Asia/Jakarta")!
        return calendar
    }

    private func jakarta(_ day: Int, _ hour: Int, _ minute: Int) -> Date {
        calendar.date(from: DateComponents(year: 2026, month: 9, day: day, hour: hour, minute: minute))!
    }

    private func stamped(_ asked: Date?) -> MoneyRecap_V1_Profile {
        .with {
            if let asked {
                $0.deletionRequestedAt = .init(date: asked)
            }
        }
    }

    func testOnlyARequestUnder24HoursOldShows() {
        let now = jakarta(23, 12, 0)
        XCTAssertTrue(RestoreAccountViewModel.shows(asked: now.addingTimeInterval(-60), now: now))
        XCTAssertTrue(RestoreAccountViewModel.shows(asked: now.addingTimeInterval(-86399), now: now))
        XCTAssertFalse(RestoreAccountViewModel.shows(asked: now.addingTimeInterval(-86400), now: now))
        XCTAssertFalse(RestoreAccountViewModel.shows(asked: now.addingTimeInterval(-90000), now: now))
        XCTAssertFalse(RestoreAccountViewModel.shows(asked: nil, now: now))
    }

    func testTheSheetComesBeforeTheSetupCheckAndBeforeAStoredHome() async {
        let now = jakarta(23, 12, 0)
        let asked = jakarta(23, 9, 12)
        for greeted in [false, true] {
            var states = 0
            let model = LaunchViewModel(greetedBefore: greeted, setZone: { self.stamped(asked) }, state: {
                states += 1
                return .with { $0.nextStep = 3 }
            }, now: { now })
            await model.start()
            XCTAssertEqual(model.route, .restore(asked: asked), "greeted \(greeted)")
            XCTAssertEqual(states, 0, "the setup check waits for the answer")
            await model.kept()
            XCTAssertEqual(states, 1)
            guard case .stepThree = model.route else {
                return XCTFail("kept with step 3 unanswered: \(model.route), want step 3")
            }
        }
    }

    func testAnOldRequestOrNoneNeverShows() async {
        let now = jakarta(23, 12, 0)
        for asked in [jakarta(22, 11, 59), nil] {
            let fresh = LaunchViewModel(greetedBefore: false, setZone: { self.stamped(asked) }, state: {
                .with { $0.nextStep = 0 }
            }, now: { now })
            await fresh.start()
            XCTAssertEqual(fresh.route, .home)
            let greeted = LaunchViewModel(greetedBefore: true, setZone: { self.stamped(asked) }, state: { nil }, now: {
                now
            })
            await greeted.start()
            XCTAssertEqual(greeted.route, .home)
        }
    }

    func testTryAgainAlsoChecksTheRequest() async {
        let now = jakarta(23, 12, 0)
        var answers: [MoneyRecap_V1_Profile?] = [nil, stamped(jakarta(23, 9, 12))]
        let model = LaunchViewModel(greetedBefore: false, setZone: { answers.removeFirst() }, state: { nil }, now: {
            now
        })
        await model.start()
        XCTAssertEqual(model.route, .noConnection)
        await model.retry()
        XCTAssertEqual(model.route, .restore(asked: jakarta(23, 9, 12)))
    }

    func testTheLinesNameTheRealDays() {
        let sheet = { (asked: Date, now: Date) in
            RestoreAccountViewModel(asked: asked, now: { now }, calendar: self.calendar, keep: { _ in .saved })
        }
        let today = sheet(jakarta(23, 9, 12), jakarta(23, 12, 0))
        XCTAssertEqual(today.askedLine, "You asked us to delete this account today at 09:12. Nothing has gone yet.")
        XCTAssertEqual(
            today.loseLine,
            "Keep it and every spend, every recap and every setting stays exactly as it was. "
                + "Continue and it all goes at 09:12 tomorrow, for good."
        )
        let yesterday = sheet(jakarta(22, 21, 5), jakarta(23, 8, 0))
        XCTAssertEqual(
            yesterday.askedLine,
            "You asked us to delete this account yesterday at 21:05. Nothing has gone yet."
        )
        XCTAssertTrue(yesterday.loseLine.hasSuffix("Continue and it all goes at 21:05 today, for good."))
    }

    func testAFailedKeepSaysSoAndRetriesWithTheSameKey() async {
        var results: [SaveResult] = [.offline, .failed, .saved]
        var keys: [String] = []
        let model = RestoreAccountViewModel(asked: .now, keep: { headers in
            keys.append(headers["Idempotency-Key"]?.first ?? "")
            return results.removeFirst()
        })
        let first = await model.keep()
        XCTAssertFalse(first)
        XCTAssertTrue(model.failed)
        let second = await model.keep()
        XCTAssertFalse(second)
        XCTAssertTrue(model.failed)
        let third = await model.keep()
        XCTAssertTrue(third)
        XCTAssertFalse(model.failed)
        XCTAssertEqual(Set(keys).count, 1)
        XCTAssertFalse(keys[0].isEmpty)
    }

    func testTheFailureLineIsDrawnInTheSheet() async {
        let model = RestoreAccountViewModel(asked: .now, keep: { _ in .failed })
        let drawn = { self.picture(RestoreAccountView(model: model, onKept: {}, onContinue: {})) }
        let before = drawn()
        XCTAssertNotNil(before)
        _ = await model.keep()
        XCTAssertNotEqual(before, drawn())
    }
}
