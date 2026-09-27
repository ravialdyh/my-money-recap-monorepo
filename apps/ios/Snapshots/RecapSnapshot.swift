import Foundation
import MoneyRecapContract
import SwiftData

/// One month's Recap as the BFF last sent it (IOS I4), beside `HomeSnapshot`, so a Recap already opened reads offline
/// (D61). An answer replaces its month whole; `HomeSnapshot.erase` deletes every month with the rest.
@Model
final class RecapSnapshot {
    @Attribute(.unique) var month: String
    var bytes: Data
    var updatedAt: Date

    init(month: String, bytes: Data, updatedAt: Date = .now) {
        self.month = month
        self.bytes = bytes
        self.updatedAt = updatedAt
    }

    static func load(_ month: String, from context: ModelContext) -> MoneyRecap_V1_RecapView? {
        let stored = try? context.fetch(FetchDescriptor<RecapSnapshot>(predicate: #Predicate { $0.month == month }))
        return stored?.first.flatMap { try? MoneyRecap_V1_RecapView(serializedBytes: $0.bytes) }
    }

    static func replace(with recap: MoneyRecap_V1_RecapView, in context: ModelContext) throws {
        let month = recap.recap.month
        try context.delete(model: RecapSnapshot.self, where: #Predicate { $0.month == month })
        try context.insert(RecapSnapshot(month: month, bytes: recap.serializedBytes()))
        try context.save()
    }
}
