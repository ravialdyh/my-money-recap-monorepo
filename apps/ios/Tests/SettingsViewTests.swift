import MoneyRecapContract
import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// D1 and D2 as drawn (SCREEN-STATES D1, D2): filled, with nothing set, offline, and held sideways, where each scrolls
/// rather than clips.
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
}
