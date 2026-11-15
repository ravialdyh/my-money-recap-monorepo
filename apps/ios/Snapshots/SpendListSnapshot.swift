import Foundation
import MoneyRecapContract
import SwiftData

@Model
final class SpendListSnapshot {
    @Attribute(.unique) var month: String
    var bytes: Data
    var updatedAt: Date

    init(month: String, bytes: Data, updatedAt: Date = .now) {
        self.month = month
        self.bytes = bytes
        self.updatedAt = updatedAt
    }

    static func load(_ month: String, from context: ModelContext) -> MoneyRecap_V1_SpendListView? {
        let stored = try? context.fetch(FetchDescriptor<SpendListSnapshot>(predicate: #Predicate { $0.month == month }))
        return stored?.first.flatMap { try? MoneyRecap_V1_SpendListView(serializedBytes: $0.bytes) }
    }

    static func replace(with list: MoneyRecap_V1_SpendListView, in context: ModelContext) throws {
        let month = list.month
        try context.delete(model: SpendListSnapshot.self, where: #Predicate { $0.month == month })
        try context.insert(SpendListSnapshot(month: month, bytes: list.serializedBytes()))
        try context.save()
    }
}
