import Connect
import Foundation
import MoneyRecapContract

/// The generated `connect-swift` client for the BFF, the only service the phone talks to (IOS I3).
/// No request or response type is written in this target; every message comes from `MoneyRecapContract`.
enum BffClient {
    /// The Compose BFF (`infra/compose/README.md`, host port 8082). `MMR_BFF_URL` overrides it: a device on
    /// the LAN, or the UI test relaunching against an unreachable port to prove the snapshot renders offline.
    static var baseURL: String {
        ProcessInfo.processInfo.environment["MMR_BFF_URL"] ?? "http://localhost:8082"
    }

    static func make(session: AuthSession) -> MoneyRecap_V1_BffServiceClient {
        let config = ProtocolClientConfig(
            host: baseURL,
            networkProtocol: .connect,
            codec: ProtoCodec(),
            timeout: 15,
            interceptors: [InterceptorFactory { _ in BearerInterceptor(session: session) }]
        )
        return MoneyRecap_V1_BffServiceClient(client: ProtocolClient(config: config))
    }

    /// Calls once more when the first answer is `unauthenticated` and the session still holds, after the interceptor
    /// has refreshed the token (IOS I3).
    @MainActor
    static func retried<Output>(
        _ session: AuthSession,
        _ call: () async -> ResponseMessage<Output>
    ) async -> ResponseMessage<Output> {
        let first = await call()
        guard first.code == .unauthenticated, session.isSignedIn else { return first }
        return await call()
    }
}
