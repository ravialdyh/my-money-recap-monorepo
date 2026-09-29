import Connect
import MoneyRecapContract
import XCTest
@testable import MyMoneyRecap

/// D4 Saving target (D65, D44) on 29 September in Jakarta over the recorded Settings answers: the forward-only line,
/// the no-Salary card, a failed save, and D4 opened from D1 and from each Recap target card.
@MainActor
final class TargetEditViewModelTests: XCTestCase {
    private typealias Answer = Result<MoneyRecap_V1_BffServiceSaveTargetResponse, ConnectError>

    private func model(
        _ name: String,
        save: @escaping TargetEditViewModel.Save = { _, _ in .success(.init()) }
    ) throws -> TargetEditViewModel {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "Asia/Jakarta") ?? .gmt
        let today = try XCTUnwrap(calendar.date(from: DateComponents(year: 2026, month: 9, day: 29, hour: 9)))
        return try TargetEditViewModel(
            settings: SettingsViewModelTests.recorded(name), today: today, calendar: calendar,
            locale: Locale(identifier: "en_GB"), save: save
        )
    }

    func testAnEditSaysItStartsNextMonth() throws {
        let edit = try model("settings-filled")
        XCTAssertFalse(edit.isBlocked)
        XCTAssertEqual(edit.percent, 20, "the Target in force")
        XCTAssertEqual(edit.startsLine, "Your new target starts in October.")
        XCTAssertEqual(edit.preview, "IDR\u{00A0}3,000,000 of IDR\u{00A0}15,000,000 each month")
        edit.percent = 35
        XCTAssertEqual(edit.preview, "IDR\u{00A0}5,250,000 of IDR\u{00A0}15,000,000 each month")
    }

    func testAFirstTargetAfterSetupHasNoLine() throws {
        var settings = try SettingsViewModelTests.recorded("settings-filled")
        settings.ledgerSettings.clearSavingTarget()
        settings.ledgerSettings.newSavingTargetFrom = "2026-09"
        let edit = TargetEditViewModel(settings: settings, locale: Locale(identifier: "en_GB")) { _, _ in
            .success(.init())
        }
        XCTAssertEqual(edit.percent, 20, "setup's default with none saved")
        XCTAssertNil(try model("settings-not-set").startsLine, "a first Target counts from this month")
    }

    func testAWaitingTargetIsTheOneShown() throws {
        var settings = try SettingsViewModelTests.recorded("settings-filled")
        settings.ledgerSettings.nextSavingTarget = .with {
            $0.percent = 30
            $0.effectiveFrom = "2026-10"
        }
        XCTAssertEqual(TargetEditViewModel(settings: settings) { _, _ in .success(.init()) }.percent, 30)
    }

    func testNoSalaryBlocksTheSliderAndSavesNothing() async throws {
        var calls = 0
        let edit = try model("settings-not-set") { _, _ in
            calls += 1
            return .success(.init())
        }
        XCTAssertTrue(edit.isBlocked, "D44: setup's Set a salary first card")
        XCTAssertNil(edit.preview)
        XCTAssertFalse(edit.canSave)
        let saved = await edit.commit()
        XCTAssertFalse(saved)
        XCTAssertEqual(calls, 0)
    }

    func testAFailedSaveKeepsTheSliderAndItsKey() async throws {
        var sent: [(MoneyRecap_V1_BffServiceSaveTargetRequest, Headers)] = []
        var answer: Answer = .failure(ConnectError(code: .unavailable, message: nil))
        let edit = try model("settings-filled") { request, headers in
            sent.append((request, headers))
            return answer
        }
        edit.percent = 40
        let failed = await edit.commit()
        XCTAssertFalse(failed)
        XCTAssertTrue(edit.failed)
        XCTAssertEqual(edit.percent, 40, "the slider stays where it was")
        answer = .success(.init())
        let saved = await edit.commit()
        XCTAssertTrue(saved)
        XCTAssertFalse(edit.failed)
        XCTAssertEqual(sent.map(\.0.percent), [40, 40])
        XCTAssertEqual(sent[0].1, sent[1].1, "the retry counts once")
    }

    /// D1 opens D4 over its stored answer; a save's Settings replace the stored copy, keeping the stored notification
    /// preferences the answer lacks; D4's no-Salary card puts D3 in its place.
    func testSettingsOpensD4AndItsSaveReplacesTheStoredSettings() async throws {
        let context = try SettingsViewModelTests.context()
        let answer = try MoneyRecap_V1_BffServiceSaveTargetResponse.with {
            $0.settings = try SettingsViewModelTests.recorded("settings-waiting")
        }
        let settings = SettingsViewModel(
            fetch: { try SettingsViewModelTests.recorded("settings-filled") }, saveProfile: { _, _ in .saved },
            saveTarget: { _, _ in .success(answer) }, context: context, locale: Locale(identifier: "id_ID")
        )
        settings.openTarget()
        XCTAssertNil(settings.targetEdit, "no answer yet: nothing to fill D4 with")
        await settings.refresh()
        settings.openTarget()
        let saved = try await XCTUnwrap(settings.targetEdit).commit()
        XCTAssertTrue(saved)
        var kept = answer.settings
        kept.notificationPrefs = try SettingsViewModelTests.recorded("settings-filled").notificationPrefs
        XCTAssertEqual(SettingsSnapshot.load(from: context), kept, "the answer, with the stored preferences it lacks")
        XCTAssertEqual(settings.settings, kept)
        settings.setSalaryFirst()
        XCTAssertNil(settings.targetEdit)
        XCTAssertNotNil(settings.salaryEdit, "D3 opens in D4's place")
    }

    /// Each Recap target card, on screens 2, 3 and 4, opens D4 over the Recap; its save replaces Home and closing it
    /// leaves the Recap on the same page.
    func testEachRecapTargetCardOpensD4OverItsPage() async throws {
        let context = try SettingsViewModelTests.context()
        let recap = try RecapViewModelTests.recorded("recap-no-target")
        var saved = MoneyRecap_V1_BffServiceSaveTargetResponse()
        saved.home.month = "2026-10"
        let home = HomeViewModel(
            fetch: { MoneyRecap_V1_HomeView() }, context: context, getRecap: { _ in recap },
            getSettings: { try SettingsViewModelTests.recorded("settings-filled") },
            saveTarget: { _, _ in .success(saved) }
        )
        for page in 1 ... 3 {
            let model = home.recap(recap.recap.month)
            await model.load()
            model.page = page
            await model.editTarget()
            let edit = try XCTUnwrap(model.targetEdit, "page \(page + 1)'s card opens D4")
            let done = await edit.commit()
            XCTAssertTrue(done)
            model.targetEdit = nil
            XCTAssertEqual(model.page, page, "back on the same page")
        }
        XCTAssertEqual(home.home?.month, "2026-10", "the save's Home replaces Home's")
    }

    func testWithNothingStoredAndNoAnswerTheCardOpensNothing() async throws {
        let home = try HomeViewModel(fetch: { MoneyRecap_V1_HomeView() }, context: SettingsViewModelTests.context())
        let model = home.recap("2026-08")
        await model.editTarget()
        XCTAssertNil(model.targetEdit)
    }
}
