import MoneyRecapContract
import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// D1 to D3 as drawn (SCREEN-STATES D1 to D3): filled, with nothing set, offline, D3 with a waiting Salary, and held
/// sideways, where each, the D56 sheet included, scrolls rather than clips.
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
}
