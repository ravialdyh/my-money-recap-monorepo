import Connect
import MoneyRecapContract
import SwiftData
import XCTest
@testable import MyMoneyRecap

/// D1 Settings and D2 Profile (PRD §6.16) from the BFF's recorded `GetSettings` answers for the test Member in rupiah:
/// `settings-filled.json` with a Salary paid on the 25th and a 20% Target, `settings-not-set.json` after steps 1 and 2.
@MainActor
final class SettingsViewModelTests: XCTestCase {
    static func recorded(_ name: String) throws -> MoneyRecap_V1_SettingsView {
        let url = try XCTUnwrap(Bundle(for: Self.self).url(forResource: name, withExtension: "json"))
        return try MoneyRecap_V1_BffServiceGetSettingsResponse(jsonString: String(contentsOf: url, encoding: .utf8))
            .settings
    }

    static func context() throws -> ModelContext {
        try ModelContext(ModelContainer(
            for: HomeSnapshot.self, SpendListSnapshot.self, RecapSnapshot.self, SettingsSnapshot.self,
            configurations: ModelConfiguration(isStoredInMemoryOnly: true)
        ))
    }

    static func model(
        _ fetch: @escaping SettingsViewModel.Fetch,
        context: ModelContext,
        online: Bool = true,
        save: @escaping NameAvatarViewModel.Save = { _, _ in .saved }
    ) -> SettingsViewModel {
        SettingsViewModel(
            fetch: fetch, saveProfile: save, context: context, locale: Locale(identifier: "id_ID"),
            connection: Connection(online: online)
        )
    }

    func testEachRowShowsItsCurrentValue() async throws {
        let filled = try Self.recorded("settings-filled")
        let model = try Self.model({ filled }, context: Self.context())
        XCTAssertNil(model.name, "nothing stored and no answer yet: the rows show without values")
        XCTAssertNil(model.salary)
        await model.refresh()
        XCTAssertEqual(model.name, "Test")
        XCTAssertEqual(model.currency, "IDR", "ledger's code, which the BFF puts on the Profile")
        XCTAssertEqual(model.salary, "Rp15.000.000 · 25th")
        XCTAssertEqual(model.target, "20%")
    }

    func testAbsentMoneyRowsReadNotSet() async throws {
        let bare = try Self.recorded("settings-not-set")
        let model = try Self.model({ bare }, context: Self.context())
        await model.refresh()
        XCTAssertEqual(model.salary, "Not set")
        XCTAssertEqual(model.target, "Not set")
        XCTAssertEqual(model.name, "Test")
    }

    func testThePaydayReadsAsAnOrdinalOrLastDay() {
        XCTAssertEqual(SettingsViewModel.payday("1"), "1st")
        XCTAssertEqual(SettingsViewModel.payday("22"), "22nd")
        XCTAssertEqual(SettingsViewModel.payday("25"), "25th")
        XCTAssertEqual(SettingsViewModel.payday("last"), "Last day")
    }

    /// The stored copy is drawn at once and replaced whole by the answer; a failed read keeps it; erasing the store, as
    /// sign-out and `--reset-snapshot` do, deletes it (IOS I4).
    func testTheStoredCopyIsDrawnAtOnceThenReplaced() async throws {
        let context = try Self.context()
        let bare = try Self.recorded("settings-not-set")
        try SettingsSnapshot.replace(with: bare, in: context)
        var answer = try Self.recorded("settings-filled")
        answer.profile.displayName = "Tess"
        let model = Self.model({ answer }, context: context)
        XCTAssertEqual(model.salary, "Not set", "the stored copy, before any answer")
        await model.refresh()
        XCTAssertEqual(model.name, "Tess")
        XCTAssertEqual(SettingsSnapshot.load(from: context), answer, "the answer's bytes replace the stored copy")

        let offline = Self.model({ throw ConnectError(code: .unavailable, message: nil) }, context: context)
        await offline.refresh()
        XCTAssertEqual(offline.name, "Tess", "a failed read keeps the stored copy")
        HomeSnapshot.erase(from: context)
        XCTAssertNil(SettingsSnapshot.load(from: context))
    }

    func testOfflineIsTheConnections() throws {
        XCTAssertTrue(try Self.model({ MoneyRecap_V1_SettingsView() }, context: Self.context(), online: false).offline)
        XCTAssertFalse(try Self.model({ MoneyRecap_V1_SettingsView() }, context: Self.context()).offline)
    }

    /// Save turns on only for a valid, changed name or Avatar, sends those two alone, and keeps what was typed when it
    /// fails; the retry reuses the key.
    func testProfileSavesAChangedNameAndKeepsItWhenTheSaveFails() async throws {
        var sent: [(MoneyRecap_V1_BffServiceSaveProfileRequest, Headers)] = []
        var result = SaveResult.failed
        let save: NameAvatarViewModel.Save = { request, headers in
            sent.append((request, headers))
            return result
        }
        let model = try Self.model({ try Self.recorded("settings-filled") }, context: Self.context(), save: save)
        model.openProfile()
        XCTAssertNil(model.profile, "no answer yet: nothing to fill Profile with")
        await model.refresh()
        model.openProfile()
        let profile = try XCTUnwrap(model.profile)
        XCTAssertEqual(profile.name, "Test")
        XCTAssertEqual(profile.avatar, "owl")
        XCTAssertFalse(profile.canSave, "nothing changed")
        profile.type("Tess 2")
        XCTAssertEqual(profile.name, "Tess ", "the digit is dropped as typed")
        XCTAssertTrue(profile.canSave)
        let failed = await profile.commit()
        XCTAssertFalse(failed)
        XCTAssertTrue(profile.failed)
        XCTAssertEqual(profile.name, "Tess ", "what was typed stays")
        result = .saved
        let saved = await profile.commit()
        XCTAssertTrue(saved)
        XCTAssertFalse(profile.failed)
        XCTAssertEqual(sent.map(\.0.displayName), ["Tess", "Tess"])
        XCTAssertEqual(sent.first?.0.avatarKey, "owl")
        XCTAssertFalse(sent.first?.0.hasOnboardingStep ?? true, "name and Avatar only")
        XCTAssertFalse(sent.first?.0.hasCurrencyCode ?? true)
        XCTAssertEqual(sent[0].1["Idempotency-Key"], sent[1].1["Idempotency-Key"], "the retry counts once")

        profile.type("Test")
        XCTAssertFalse(profile.canSave, "back to the saved name")
        profile.choose("fox")
        XCTAssertTrue(profile.canSave, "a new Avatar alone is a change")
        profile.type("  ")
        XCTAssertFalse(profile.canSave, "no name")
    }
}
