import Connect
import MoneyRecapContract

struct RecapCalls {
    var fetch: RecapViewModel.Fetch = { _ in throw ConnectError(code: .unavailable, message: nil) }

    static func live(session: AuthSession, client: MoneyRecap_V1_BffServiceClient) -> RecapCalls {
        RecapCalls(fetch: { month in
            try await BffClient.retried(session) { await client.getRecap(request: .with { $0.month = month }) }
                .result.get().recap
        })
    }
}
