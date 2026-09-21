import Connect
import Foundation

/// IOS I3's one interceptor: adds the bearer token and forces a refresh on `unauthenticated`; only a refresh the
/// realm refuses signs the Member out. A call with no fresh token fails: `unavailable` while the session holds (PRD
/// D61), else `unauthenticated`.
final class BearerInterceptor: UnaryInterceptor {
    /// Sent with no `Authorization` header and no token asked of the session. Both halves key off this path,
    /// not off having a session, which would swallow a refused refresh and leave the expired copy behind Enter
    /// email (spec #61).
    static let unauthenticatedPath = "/money_recap.v1.BffService/RequestMagicLink"

    private let session: AuthSession
    /// An interceptor is built once per request, so the request half tells the response half which path
    /// the call is on; `HTTPResponse` carries none.
    private nonisolated(unsafe) var isUnauthenticatedPath = false

    init(session: AuthSession) {
        self.session = session
    }

    func handleUnaryRawRequest(
        _ request: HTTPRequest<Data?>,
        proceed: @escaping @Sendable (Result<HTTPRequest<Data?>, ConnectError>) -> Void
    ) {
        isUnauthenticatedPath = request.url.path == Self.unauthenticatedPath
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
