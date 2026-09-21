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
        guard case let .failure(error) = try await intercept(session) else {
            return XCTFail("the call went out with the stale token")
        }
        XCTAssertEqual(error.code, .unavailable, "the phone keeps its snapshot (PRD D61)")
    }

    func testARefusedRefreshOnTheRequestPathSignsOut() async throws {
        let session = expiredSession(realm: "refusing.invalid")
        try Keychain.write(NSKeyedArchiver.archivedData(
            withRootObject: XCTUnwrap(session.authState), requiringSecureCoding: true
        ))
        guard case let .failure(error) = try await intercept(session) else {
            return XCTFail("the call went out with a refused refresh")
        }
        XCTAssertEqual(error.code, .unauthenticated)
        XCTAssertNil(session.authState, "RootView observes authState, so Sign in shows")
        XCTAssertNil(Keychain.read(), "a relaunch must not reload the refused state")
        XCTAssertEqual(session.signInMessage, AuthSession.expiredCopy)
    }

    func testTheMagicLinkCallGoesOutWithNoHeaderAndWithoutAskingForAToken() async throws {
        let session = expiredSession(realm: "refusing.invalid")
        let sent = try await intercept(session, path: BearerInterceptor.unauthenticatedPath)
        guard case let .success(request) = sent else {
            return XCTFail("RequestMagicLink was failed before it was sent, so email sign-in cannot start")
        }
        XCTAssertNil(request.headers["Authorization"], "the one procedure sent with no bearer token")
        XCTAssertTrue(session.isSignedIn, "the session was never asked for a token, so nothing refused it")
    }

    func testOnlyTheMagicLinkPathSkipsTheRecoverStepOnAnUnauthenticatedAnswer() async throws {
        let skipped = expiredSession(realm: "refusing.invalid")
        let onTheEmailPath = BearerInterceptor(session: skipped)
        _ = try await intercept(skipped, path: BearerInterceptor.unauthenticatedPath, through: onTheEmailPath)
        await respondUnauthenticated(onTheEmailPath)
        XCTAssertTrue(skipped.isSignedIn, "a tokenless call's refusal is not a refused refresh")
        XCTAssertNil(skipped.signInMessage, "else the expired copy waits on Sign in behind Enter email")

        let recovered = expiredSession(realm: "refusing.invalid")
        await respondUnauthenticated(BearerInterceptor(session: recovered))
        XCTAssertFalse(recovered.isSignedIn, "every other path still recovers, and signs out when refused")
        XCTAssertEqual(recovered.signInMessage, AuthSession.expiredCopy)
    }

    /// Runs `BearerInterceptor`'s request step alone, where no response interceptor follows a failure.
    private func intercept(
        _ session: AuthSession,
        path: String = "/money_recap.v1.BffService/GetHome",
        through interceptor: BearerInterceptor? = nil
    ) async throws -> Result<HTTPRequest<Data?>, ConnectError> {
        let request = try HTTPRequest<Data?>(
            url: XCTUnwrap(URL(string: BffClient.baseURL + path)), headers: [:], message: nil, method: .post,
            trailers: nil, idempotencyLevel: .unknown
        )
        let bearer = interceptor ?? BearerInterceptor(session: session)
        return await withCheckedContinuation { continuation in
            bearer.handleUnaryRawRequest(request) { continuation.resume(returning: $0) }
        }
    }

    /// The BFF's `unauthenticated`, which every path but one answers by forcing a refresh.
    private func respondUnauthenticated(_ interceptor: BearerInterceptor) async {
        let answer = HTTPResponse(
            code: .unauthenticated, headers: [:], message: nil, trailers: [:], error: nil, tracingInfo: nil
        )
        await withCheckedContinuation { (continuation: CheckedContinuation<Void, Never>) in
            interceptor.handleUnaryRawResponse(answer) { _ in continuation.resume() }
        }
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
