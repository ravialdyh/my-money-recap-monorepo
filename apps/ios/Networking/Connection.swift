import Network
import Observation

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
