import Connect
import Foundation
import MoneyRecapContract

final class BearerInterceptor: UnaryInterceptor {
    static let unauthenticatedPath = "/" + MoneyRecap_V1_BffServiceClient.Metadata.Methods.requestMagicLink.path

    private let session: AuthSession
    private nonisolated(unsafe) var isUnauthenticatedPath = false

    init(session: AuthSession) {
        self.session = session
    }

    func handleUnaryRawRequest(
        _ request: HTTPRequest<Data?>,
        proceed: @escaping @Sendable (Result<HTTPRequest<Data?>, ConnectError>) -> Void
    ) {
        isUnauthenticatedPath = request.url.path.hasSuffix(Self.unauthenticatedPath)
        guard !isUnauthenticatedPath else { return proceed(.success(request)) }
        Task { @MainActor in
            do {
                let token = try await session.freshAccessToken()
                var headers = request.headers
                headers["Authorization"] = ["Bearer \(token)"]
                proceed(.success(HTTPRequest(
                    url: request.url, headers: headers, message: request.message, method: request.method,
                    trailers: request.trailers, idempotencyLevel: request.idempotencyLevel
                )))
            } catch {
                let code: Code = session.isSignedIn ? .unavailable : .unauthenticated
                proceed(.failure(ConnectError(code: code, message: "no fresh access token", exception: error)))
            }
        }
    }

    func handleUnaryRawResponse(_ response: HTTPResponse, proceed: @escaping @Sendable (HTTPResponse) -> Void) {
        guard response.code == .unauthenticated, !isUnauthenticatedPath else { return proceed(response) }
        Task { @MainActor in
            await session.recoverFromUnauthenticated()
            proceed(response)
        }
    }
}
