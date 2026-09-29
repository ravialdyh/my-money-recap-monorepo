import Foundation
import MoneyRecapContract
import SwiftData

/// D1 Settings (PRD §6.16, IOS I4, I6): the stored `SettingsView` at once, then the BFF's, each answer replacing the
/// stored copy whole. Every value is the server's; this only words it. Offline, every row that writes is disabled.
@Observable
@MainActor
final class SettingsViewModel: Identifiable {
    typealias Fetch = () async throws -> MoneyRecap_V1_SettingsView

    /// The line every Settings save shows when it fails (spec #169).
    static let saveFailure = "We couldn't save this. Try again."

    /// Nil until a stored copy or an answer exists: the rows show without values.
    private(set) var settings: MoneyRecap_V1_SettingsView?
    /// D2 Profile, while it is open.
    var profile: ProfileViewModel?
    private let fetch: Fetch
    private let saveProfile: NameAvatarViewModel.Save
    private let context: ModelContext
    private let locale: Locale
    private let connection: Connection

    init(
        fetch: @escaping Fetch,
        saveProfile: @escaping NameAvatarViewModel.Save,
        context: ModelContext,
        locale: Locale = .autoupdatingCurrent,
        connection: Connection? = nil
    ) {
        self.fetch = fetch
        self.saveProfile = saveProfile
        self.context = context
        self.locale = locale
        self.connection = connection ?? .shared
        settings = SettingsSnapshot.load(from: context)
    }

    /// The answer replaces the stored copy and what is shown; a failure keeps both, offline included.
    func refresh() async {
        guard let fresh = try? await fetch() else { return }
        try? SettingsSnapshot.replace(with: fresh, in: context)
        settings = fresh
    }

    /// No network path: the grey Banner, and every row but Log out at 40% (D1 offline).
    var offline: Bool {
        !connection.online
    }

    var name: String? {
        settings?.profile.displayName
    }

    var currency: String? {
        settings?.profile.currencyCode
    }

    /// "Rp15.000.000 · 25th", "{amount} · Last day", or "Not set".
    var salary: String? {
        guard let ledger = settings?.ledgerSettings else { return nil }
        guard ledger.hasSalary else { return "Not set" }
        let amount = Money.format(
            ledger.salary.amount.amount,
            currency: ledger.salary.amount.currencyCode,
            locale: locale
        )
        return "\(amount) · \(Self.payday(ledger.salary.payday))"
    }

    /// "20%", or "Not set".
    var target: String? {
        guard let ledger = settings?.ledgerSettings else { return nil }
        return ledger.hasSavingTarget ? "\(ledger.savingTarget.percent)%" : "Not set"
    }

    /// D2 over the saved answers; nil before Settings has any.
    func openProfile() {
        guard let saved = settings?.profile else { return }
        profile = ProfileViewModel(name: saved.displayName, avatar: saved.avatarKey, save: saveProfile)
    }

    /// A stored Payday as the row prints it: "25th", or "Last day" for the token (PRD D51).
    static func payday(_ stored: String) -> String {
        guard let day = Int(stored) else { return "Last day" }
        let ordinal = NumberFormatter()
        ordinal.numberStyle = .ordinal
        ordinal.locale = Locale(identifier: "en")
        return ordinal.string(from: day as NSNumber) ?? stored
    }
}
