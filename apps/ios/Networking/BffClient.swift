import Connect
import Foundation
import MoneyRecapContract

enum BffClient {
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
