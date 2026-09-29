import Connect
import MoneyRecapContract
import XCTest
@testable import MyMoneyRecap

/// D3 Salary and payday (PRD §8.10, D56) on 29 September in Jakarta, over the recorded Settings answers: when "Which
/// salary is this?" opens and when it does not, each row's answer, the three lines and the failed saves.
@MainActor
final class SalaryEditViewModelTests: XCTestCase {
    private typealias Sent = (request: MoneyRecap_V1_BffServiceSaveSalaryRequest, headers: Headers)

    private var calendar: Calendar {
        var gregorian = Calendar(identifier: .gregorian)
        gregorian.timeZone = TimeZone(identifier: "Asia/Jakarta") ?? .gmt
        return gregorian
    }

    private func settings(_ name: String, passed: Bool? = nil) throws -> MoneyRecap_V1_SettingsView {
        var recorded = try SettingsViewModelTests.recorded(name)
        if let passed {
            recorded.ledgerSettings.paydayPassedThisMonth = passed
        }
        return recorded
    }

    private func model(
        _ settings: MoneyRecap_V1_SettingsView,
        save: @escaping SalaryEditViewModel.Save = { _, _ in .success(.init()) }
    ) throws -> SalaryEditViewModel {
        let today = try XCTUnwrap(calendar.date(from: DateComponents(year: 2026, month: 9, day: 29, hour: 9)))
        return SalaryEditViewModel(
            settings: settings, today: today, calendar: calendar, locale: Locale(identifier: "en_GB"), save: save
        )
    }

    private typealias Answer = Result<MoneyRecap_V1_BffServiceSaveSalaryResponse, ConnectError>

    private func recording(_ result: @escaping () -> Answer) -> (SalaryEditViewModel.Save, () -> [Sent]) {
        var sent: [Sent] = []
        return ({ request, headers in
            sent.append((request, headers))
            return result()
        }, { sent })
    }

    func testAFirstSalarySavesWithNoQuestion() async throws {
        let (save, sent) = recording { .success(.init()) }
        let edit = try model(settings("settings-not-set"), save: save)
        XCTAssertEqual(edit.amount, "", "no Salary in force and none waiting")
        XCTAssertEqual(edit.payday, "25", "setup's default Payday")
        XCTAssertFalse(edit.canSave)
        XCTAssertEqual(edit.nextPaydays, ["25 Oct", "25 Nov", "25 Dec"], "counted from today, 25 September passed")
        edit.amount = "15000000"
        XCTAssertTrue(edit.canSave)
        XCTAssertNil(edit.startsLine, "no saved Payday to differ from")
        let saved = await edit.commit()
        XCTAssertTrue(saved)
        XCTAssertFalse(edit.asking)
        XCTAssertEqual(
            sent().map(\.request.appliesFrom),
            [.unspecified],
            "ledger counts a first Salary from this month"
        )
        XCTAssertEqual(sent().first?.request.amount.amount, "15000000")
        XCTAssertEqual(sent().first?.request.payday, "25")
    }

    func testAChangedAmountBeforePaydaySavesForThisMonthWithNoQuestion() async throws {
        let (save, sent) = recording { .success(.init()) }
        let edit = try model(settings("settings-filled", passed: false), save: save)
        XCTAssertEqual(edit.amount, "15000000", "the Salary in force")
        XCTAssertFalse(edit.canSave, "nothing changed")
        edit.amount = "17500000"
        let saved = await edit.commit()
        XCTAssertTrue(saved)
        XCTAssertFalse(edit.asking)
        XCTAssertEqual(sent().map(\.request.appliesFrom), [.unspecified])
    }

    func testAChangedAmountOnOrAfterPaydayAsksWhichSalaryItIs() async throws {
        let (save, sent) = recording { .success(.init()) }
        let edit = try model(settings("settings-filled"), save: save)
        edit.amount = "17500000"
        let first = await edit.commit()
        XCTAssertFalse(first)
        XCTAssertTrue(edit.asking, "payday_passed_this_month opens the sheet")
        XCTAssertTrue(sent().isEmpty, "nothing is sent before the answer")
        XCTAssertEqual(
            edit.sheetBody,
            "You were paid on 25 September. "
                + "A date alone can't tell a correction from a raise, so tell us which this is."
        )
        XCTAssertEqual(edit.correctionRow, "The IDR\u{00A0}17,500,000 I was paid on 25 September")
        XCTAssertEqual(edit.correctionDetail, "Fixes September")
        XCTAssertEqual(edit.raiseRow, "My new salary, from 25 October")
        XCTAssertEqual(edit.raiseDetail, "September stays at IDR\u{00A0}15,000,000")

        edit.asking = false
        XCTAssertTrue(sent().isEmpty, "Cancel saves nothing")
        _ = await edit.commit()
        let raise = await edit.answer(.nextMonth)
        XCTAssertTrue(raise)
        XCTAssertFalse(edit.asking)
        let correction = await edit.answer(.thisMonth)
        XCTAssertTrue(correction)
        XCTAssertEqual(sent().map(\.request.appliesFrom), [.nextMonth, .thisMonth], "each row's answer")
        XCTAssertNotEqual(sent()[0].headers, sent()[1].headers, "another answer is another save")
    }

    func testFailedPreconditionOpensTheSheet() async throws {
        var answer: Answer = .failure(ConnectError(code: .failedPrecondition, message: "this month's Payday has come"))
        let (save, sent) = recording { answer }
        let edit = try model(settings("settings-filled", passed: false), save: save)
        edit.amount = "17500000"
        let refused = await edit.commit()
        XCTAssertFalse(refused)
        XCTAssertTrue(edit.asking, "the Payday passed while D3 was open")
        XCTAssertFalse(edit.failed, "a question, not a failure")
        answer = .success(.init())
        let saved = await edit.answer(.thisMonth)
        XCTAssertTrue(saved)
        XCTAssertEqual(sent().map(\.request.appliesFrom), [.unspecified, .thisMonth])
    }

    func testAChangedPaydaySaysWhenItStartsAndNeverAsks() async throws {
        let (save, sent) = recording { .success(.init()) }
        let edit = try model(settings("settings-filled"), save: save)
        edit.choose("last")
        XCTAssertEqual(edit.startsLine, "Your new payday starts in October.")
        XCTAssertEqual(edit.nextPaydays, ["30 Sep", "31 Oct", "30 Nov"])
        edit.amount = "17500000"
        XCTAssertEqual(edit.startsLine, "Your new salary and payday start in October.")
        let saved = await edit.commit()
        XCTAssertTrue(saved)
        XCTAssertFalse(edit.asking, "both changed: no sheet, both start next month")
        XCTAssertEqual(sent().map(\.request.payday), ["last"])
        edit.choose("25")
        XCTAssertNil(edit.startsLine, "the saved day again: no line")
    }

    func testAWaitingSalaryIsNamedUnderTheAmountCard() throws {
        XCTAssertEqual(
            try model(settings("settings-waiting")).waitingLine,
            "IDR\u{00A0}17,500,000 from your next payday, 25 October."
        )
        XCTAssertNil(try model(settings("settings-filled")).waitingLine)
    }

    func testAFailedSaveKeepsWhatWasTypedAndRetriesUnderTheSameKey() async throws {
        let (save, sent) = recording { .failure(ConnectError(code: .unavailable, message: nil)) }
        let edit = try model(settings("settings-filled", passed: false), save: save)
        edit.amount = "17500000"
        let first = await edit.commit()
        let second = await edit.commit()
        XCTAssertFalse(first || second)
        XCTAssertTrue(edit.failed)
        XCTAssertFalse(edit.asking)
        XCTAssertEqual(edit.amount, "17500000", "what was typed stays")
        XCTAssertEqual(sent()[0].headers, sent()[1].headers, "the retry counts once")
    }

    /// D1 opens D3 over its stored answer; a save's Settings replace the stored copy and the row, keeping the stored
    /// notification preferences, which a Salary save's answer does not carry.
    func testASaveReplacesTheStoredSettings() async throws {
        let context = try SettingsViewModelTests.context()
        let answer = try MoneyRecap_V1_BffServiceSaveSalaryResponse.with {
            $0.settings = try settings("settings-waiting")
        }
        let settingsModel = SettingsViewModel(
            fetch: { try self.settings("settings-filled") }, saveProfile: { _, _ in .saved },
            saveSalary: { _, _ in .success(answer) }, context: context, locale: Locale(identifier: "id_ID")
        )
        settingsModel.openSalary()
        XCTAssertNil(settingsModel.salaryEdit, "no answer yet: nothing to fill D3 with")
        await settingsModel.refresh()
        settingsModel.openSalary()
        let edit = try XCTUnwrap(settingsModel.salaryEdit)
        edit.amount = "16000000"
        _ = await edit.commit()
        let saved = await edit.answer(.nextMonth)
        XCTAssertTrue(saved)
        var kept = answer.settings
        kept.notificationPrefs = try settings("settings-filled").notificationPrefs
        XCTAssertEqual(SettingsSnapshot.load(from: context), kept, "the answer, with the stored preferences it lacks")
        XCTAssertEqual(settingsModel.settings, kept)
    }
}
