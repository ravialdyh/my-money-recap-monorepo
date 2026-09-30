import MoneyRecapContract
import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// D1 to D4 as drawn (SCREEN-STATES D1 to D4): filled, with nothing set, offline, D3 with a waiting Salary, D4 with and
/// without its line and with no Salary, and held sideways, where each, the D56 sheet included, scrolls rather than
/// clips.
@MainActor
final class SettingsViewTests: XCTestCase {
    private let sideways = CGSize(width: 852, height: 393)

    private func settings(_ name: String?, online: Bool = true) throws -> SettingsView {
        let context = try SettingsViewModelTests.context()
        if let name {
            try SettingsSnapshot.replace(with: SettingsViewModelTests.recorded(name), in: context)
        }
        return SettingsView(model: SettingsViewModelTests.model({
            try await Task.sleep(for: .seconds(3600))
            return MoneyRecap_V1_SettingsView()
        }, context: context, online: online))
    }

    func testFilledNotSetAndOfflineDrawApart() throws {
        let filled = try picture(settings("settings-filled"))
        let notSet = try picture(settings("settings-not-set"))
        let offline = try picture(settings("settings-filled", online: false))
        let empty = try picture(settings(nil))
        XCTAssertNotNil(filled)
        XCTAssertNotEqual(filled, notSet, "the money rows read Not set")
        XCTAssertNotEqual(filled, offline, "the Banner and the rows at 40%")
        XCTAssertNotEqual(filled, empty, "nothing stored: the rows without values")
    }

    func testSettingsAndProfileScrollSidewaysRatherThanClip() throws {
        let list = try settings("settings-filled")
        XCTAssertGreaterThan(overflow(list, size: sideways), 0, "sideways it scrolls")
        let profile = ProfileView(model: ProfileViewModel(name: "Test", avatar: "owl") { _, _ in .saved }) {}
        XCTAssertGreaterThan(overflow(profile, size: sideways), 0, "D2 scrolls sideways too")
    }

    private func salary(_ name: String) throws -> SalaryEditViewModel {
        try SalaryEditViewModel(
            settings: SettingsViewModelTests.recorded(name), locale: Locale(identifier: "id_ID"),
            save: { _, _ in .success(.init()) }
        )
    }

    func testSalaryAndPaydayDrawsTheWaitingSalary() throws {
        let waiting = try picture(SalaryEditView(model: salary("settings-waiting")) {})
        let none = try picture(SalaryEditView(model: salary("settings-filled")) {})
        XCTAssertNotNil(waiting)
        XCTAssertNotEqual(waiting, none, "the waiting line under the amount card")
    }

    func testSalaryAndItsSheetScrollSidewaysRatherThanClip() throws {
        let edit = try salary("settings-filled")
        XCTAssertGreaterThan(overflow(SalaryEditView(model: edit) {}, size: sideways), 0, "D3 scrolls sideways")
        edit.amount = "17500000"
        XCTAssertGreaterThan(overflow(WhichSalarySheet(model: edit) { _ in }, size: sideways), 0, "so does its sheet")
    }

    private func target(_ name: String, starts: String? = nil) throws -> TargetEditView {
        var settings = try SettingsViewModelTests.recorded(name)
        if let starts {
            settings.ledgerSettings.newSavingTargetFrom = starts
        }
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "Asia/Jakarta") ?? .gmt
        let today = try XCTUnwrap(calendar.date(from: DateComponents(year: 2026, month: 9, day: 29, hour: 9)))
        return TargetEditView(model: TargetEditViewModel(
            settings: settings, today: today, calendar: calendar, locale: Locale(identifier: "id_ID")
        ) { _, _ in .success(.init()) }) {}
    }

    /// D4 draws the forward-only line only for a save that starts next month, and setup's no-Salary card in place of
    /// the slider; held sideways it scrolls rather than clips.
    func testSavingTargetDrawsItsLineAndTheNoSalaryCard() throws {
        let edit = try picture(target("settings-filled"))
        let first = try picture(target("settings-filled", starts: "2026-09"))
        let blocked = try picture(target("settings-not-set"))
        XCTAssertNotNil(edit)
        XCTAssertNotEqual(edit, first, "the line under the card")
        XCTAssertNotEqual(first, blocked, "Set a salary first in the slider's place")
        try XCTAssertGreaterThan(overflow(target("settings-filled"), size: sideways), 0, "D4 scrolls sideways")
    }
}
