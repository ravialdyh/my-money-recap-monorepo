import Foundation
import MoneyRecapContract
import SwiftData

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

    static func erase(from context: ModelContext) {
        try? context.delete(model: HomeSnapshot.self)
        try? context.delete(model: SpendListSnapshot.self)
        try? context.delete(model: RecapSnapshot.self)
        try? context.delete(model: SettingsSnapshot.self)
        try? context.save()
    }

    static func replace(with home: MoneyRecap_V1_HomeView, in context: ModelContext) throws {
        try context.delete(model: HomeSnapshot.self)
        try context.insert(HomeSnapshot(bytes: home.serializedBytes()))
        try context.save()
    }

    @MainActor
    static let container: ModelContainer = {
        let container: ModelContainer
        do {
            container = try ModelContainer(
                for: HomeSnapshot.self, SpendListSnapshot.self, RecapSnapshot.self, SettingsSnapshot.self
            )
        } catch {
            fatalError("the HomeSnapshot store could not be opened: \(error)")
        }
        if CommandLine.arguments.contains("--reset-snapshot") {
            erase(from: container.mainContext)
        }
        return container
    }()
}
