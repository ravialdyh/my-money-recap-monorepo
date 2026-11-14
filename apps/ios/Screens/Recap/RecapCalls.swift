import Connect
import MoneyRecapContract

/// The Recap's server calls (IOS I3): one box, so a second Recap call is one more field here. The default answers as
/// with no signal.
struct RecapCalls {
    var fetch: RecapViewModel.Fetch = { _ in throw ConnectError(code: .unavailable, message: nil) }

    /// The signed-in BFF client's call, retried once after a token refresh (IOS I3).
    static func live(session: AuthSession, client: MoneyRecap_V1_BffServiceClient) -> RecapCalls {
        RecapCalls(fetch: { month in
            try await BffClient.retried(session) { await client.getRecap(request: .with { $0.month = month }) }
                .result.get().recap
        })
    }
}
