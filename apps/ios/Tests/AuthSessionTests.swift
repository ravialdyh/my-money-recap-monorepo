import AppAuth
import Connect
import XCTest
@testable import MyMoneyRecap

@MainActor
final class AuthSessionTests: XCTestCase {
    override func tearDown() {
        OIDURLSessionProvider.setSession(.shared)
        Keychain.delete()
        super.tearDown()
    }

    func testATransientRefreshFailureThrowsAndKeepsTheSession() async {
        let session = expiredSession(realm: "unreachable.invalid")
        do {
            let token = try await session.freshAccessToken()
            XCTFail("freshAccessToken returned \(token) alongside the refresh error")
        } catch {}
        await session.recoverFromUnauthenticated()
        XCTAssertTrue(session.isSignedIn, "a network blip keeps the session")
        XCTAssertNil(session.signInMessage)
    }

    func testARefusedRefreshSignsOutWithTheExpiredCopy() async {
        let session = expiredSession(realm: "refusing.invalid")
        await session.recoverFromUnauthenticated()
        XCTAssertFalse(session.isSignedIn)
        XCTAssertEqual(session.signInMessage, AuthSession.expiredCopy)
    }

    func testATransientRefreshFailureFailsTheCallAsUnavailable() async throws {
        let session = expiredSession(realm: "unreachable.invalid")
        let request = try HTTPRequest<Data?>(
            url: XCTUnwrap(URL(string: BffClient.baseURL)), headers: [:], message: nil, method: .post,
            trailers: nil, idempotencyLevel: .unknown
        )
        let result = await withCheckedContinuation { continuation in
            BearerInterceptor(session: session).handleUnaryRawRequest(request) { continuation.resume(returning: $0) }
        }
        guard case let .failure(error) = result else {
            return XCTFail("the call went out with the stale token")
        }
        XCTAssertEqual(error.code, .unavailable, "the phone keeps its snapshot (PRD D61)")
    }

    /// A signed-in state whose access token has expired, so the next action refreshes against `realm`.
    private func expiredSession(realm: String) -> AuthSession {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [StubRealm.self]
        OIDURLSessionProvider.setSession(URLSession(configuration: configuration))
        let endpoint = URL(string: "https://\(realm)/token")!
        let request = OIDAuthorizationRequest(
            configuration: OIDServiceConfiguration(authorizationEndpoint: endpoint, tokenEndpoint: endpoint),
            clientId: AuthConfig.clientID, clientSecret: nil, scopes: AuthConfig.scopes,
            redirectURL: AuthConfig.redirectURL, responseType: OIDResponseTypeCode, additionalParameters: nil
        )
        let authorization = OIDAuthorizationResponse(request: request, parameters: ["code": "code" as NSString])
        let token = OIDTokenResponse(request: authorization.tokenExchangeRequest()!, parameters: [
            "access_token": "stale" as NSString, "expires_in": 0 as NSNumber,
            "refresh_token": "refresh" as NSString, "token_type": "Bearer" as NSString,
        ])
        return AuthSession(authState: OIDAuthState(
            authorizationResponse: authorization, tokenResponse: token, registrationResponse: nil
        ))
    }
}

/// The realm's token endpoint: `refusing.invalid` refuses the refresh token (RFC 6749 §5.2), any other host
/// fails in transit, as with no signal.
final class StubRealm: URLProtocol {
    override static func canInit(with _: URLRequest) -> Bool {
        true
    }

    override static func canonicalRequest(for request: URLRequest) -> URLRequest {
        request
    }

    override func startLoading() {
        guard let url = request.url, url.host == "refusing.invalid", let response = HTTPURLResponse(
            url: url, statusCode: 400, httpVersion: nil, headerFields: ["Content-Type": "application/json"]
        ) else {
            client?.urlProtocol(self, didFailWithError: URLError(.notConnectedToInternet))
            return
        }
        client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
        client?.urlProtocol(self, didLoad: Data(#"{"error":"invalid_grant"}"#.utf8))
        client?.urlProtocolDidFinishLoading(self)
    }

    override func stopLoading() {}
}
