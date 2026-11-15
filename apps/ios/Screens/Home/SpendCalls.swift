import Connect
import MoneyRecapContract

struct SpendCalls {
    var log: LogSpendViewModel.Send = unavailable
    var list: SpendListViewModel.Fetch = { _ in throw ConnectError(code: .unavailable, message: nil) }
    var edit: LogSpendViewModel.Edit = unavailable
    var delete: LogSpendViewModel.Delete = unavailable

    static func live(session: AuthSession, client: MoneyRecap_V1_BffServiceClient) -> SpendCalls {
        SpendCalls(log: { request, headers in
            await BffClient.retried(session) { await client.logSpend(request: request, headers: headers) }.result
        }, list: { month in
            try await BffClient.retried(session) { await client.listSpends(request: .with { $0.month = month }) }
                .result.get().spends
        }, edit: { request, headers in
            await BffClient.retried(session) { await client.editSpend(request: request, headers: headers) }.result
        }, delete: { request, headers in
            await BffClient.retried(session) { await client.deleteSpend(request: request, headers: headers) }.result
        })
    }
}
