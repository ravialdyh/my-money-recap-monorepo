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

    static func replace(with home: MoneyRecap_V1_HomeView, in context: ModelContext) throws {
        try context.delete(model: HomeSnapshot.self)
        try context.insert(HomeSnapshot(bytes: home.serializedBytes()))
        try context.save()
    }

    /// The on-device store. `--reset-snapshot` empties it at launch, the seam the UI test uses for the first-load
    /// skeleton.
    @MainActor
    static let container: ModelContainer = {
        let container = try! ModelContainer(for: HomeSnapshot.self)
        if CommandLine.arguments.contains("--reset-snapshot") {
            try? container.mainContext.delete(model: HomeSnapshot.self)
            try? container.mainContext.save()
        }
        return container
    }()
}
