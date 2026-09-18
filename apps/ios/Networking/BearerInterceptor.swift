import Connect
import Foundation

/// IOS I3's one interceptor: adds the bearer token and forces a refresh on `unauthenticated`; a refresh the realm
/// refuses signs the Member out. A call with no fresh token fails: `unavailable` while the session holds (PRD
/// D61), else `unauthenticated`.
final class BearerInterceptor: UnaryInterceptor {
    private let session: AuthSession

    init(session: AuthSession) {
        self.session = session
    }

    func handleUnaryRawRequest(
        _ request: HTTPRequest<Data?>,
        proceed: @escaping @Sendable (Result<HTTPRequest<Data?>, ConnectError>) -> Void
    ) {
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
        guard response.code == .unauthenticated else { return proceed(response) }
        Task { @MainActor in
            await session.recoverFromUnauthenticated()
            proceed(response)
        }
    }
}
