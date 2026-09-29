import Foundation
import MoneyRecapContract
import SwiftData

/// The last `SettingsView` as the bytes the BFF sent, plus when it arrived (IOS I4), so Settings draws at once and
/// reads
/// offline. An answer replaces it whole; nothing edits it; `HomeSnapshot.erase` deletes it with the rest.
@Model
final class SettingsSnapshot {
    var bytes: Data
    var updatedAt: Date

    init(bytes: Data, updatedAt: Date = .now) {
        self.bytes = bytes
        self.updatedAt = updatedAt
    }

    static func load(from context: ModelContext) -> MoneyRecap_V1_SettingsView? {
        let stored = try? context.fetch(FetchDescriptor<SettingsSnapshot>()).first
        return stored.flatMap { try? MoneyRecap_V1_SettingsView(serializedBytes: $0.bytes) }
    }

    static func replace(with settings: MoneyRecap_V1_SettingsView, in context: ModelContext) throws {
        try context.delete(model: SettingsSnapshot.self)
        try context.insert(SettingsSnapshot(bytes: settings.serializedBytes()))
        try context.save()
    }
}
