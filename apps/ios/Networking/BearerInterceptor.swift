import Connect
import Foundation

/// The one interceptor IOS I3 names: adds the bearer token to every call and, when the BFF answers
/// `unauthenticated`, asks AppAuth to refresh. A failed refresh signs the Member out, back to Sign
/// in with the inline error.
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
                proceed(.failure(ConnectError(code: .unauthenticated, message: "no usable session", exception: error)))
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
