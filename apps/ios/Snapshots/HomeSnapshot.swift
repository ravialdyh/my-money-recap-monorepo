import Foundation
import MoneyRecapContract
import SwiftData

/// The cache is one row per screen answer (IOS I4): the last `HomeView` as the bytes the BFF sent, plus
/// when it arrived. A successful response replaces it whole; nothing edits it; sign-out deletes it.
@Model
final class HomeSnapshot {
    var bytes: Data
    var updatedAt: Date

    init(bytes: Data, updatedAt: Date = .now) {
        self.bytes = bytes
        self.updatedAt = updatedAt
    }

    static func load(from context: ModelContext) -> HomeSnapshot? {
        try? context.fetch(FetchDescriptor<HomeSnapshot>()).first
    }

    /// Empties the store: sign-out erases what the device holds (PRD §9), and a Home that refuses deletes its answer
    /// (spec #92). The stored Home is also the only setup progress on the device: it opens Home with no round trip.
    static func erase(from context: ModelContext) {
        try? context.delete(model: HomeSnapshot.self)
        try? context.save()
    }

    static func replace(with home: MoneyRecap_V1_HomeView, in context: ModelContext) throws {
        try context.delete(model: HomeSnapshot.self)
        try context.insert(HomeSnapshot(bytes: home.serializedBytes()))
        try context.save()
    }

    /// The on-device store. `--reset-snapshot` empties it at launch, the seam the UI test starts from.
    @MainActor
    static let container: ModelContainer = {
        let container: ModelContainer
        do {
            container = try ModelContainer(for: HomeSnapshot.self)
        } catch {
            fatalError("the HomeSnapshot store could not be opened: \(error)")
        }
        if CommandLine.arguments.contains("--reset-snapshot") {
            erase(from: container.mainContext)
        }
        return container
    }()
}
