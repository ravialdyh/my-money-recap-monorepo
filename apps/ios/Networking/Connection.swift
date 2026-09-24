import Network
import Observation

/// Whether the phone has a network path, as the platform's `NWPathMonitor` reports it: the offline Banner and the
/// disabled ways into Log a spend follow it live (PRD D58). `shared` watches the device; a test builds one it sets.
@Observable
@MainActor
final class Connection {
    static let shared = Connection(watching: NWPathMonitor())
    var online: Bool

    init(online: Bool) {
        self.online = online
    }

    private convenience init(watching monitor: NWPathMonitor) {
        self.init(online: true)
        Task {
            for await path in monitor {
                online = path.status == .satisfied
            }
        }
    }
}
