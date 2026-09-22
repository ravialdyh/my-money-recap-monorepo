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
        let session = signedIn(realm: "unreachable.invalid")
        do {
            let token = try await session.freshAccessToken()
            XCTFail("freshAccessToken returned \(token) alongside the refresh error")
        } catch {}
        await session.recoverFromUnauthenticated()
        XCTAssertTrue(session.isSignedIn, "a network blip keeps the session")
        XCTAssertNil(session.signInMessage)
    }

    func testARefusedRefreshSignsOutWithTheExpiredCopy() async {
        let session = signedIn(realm: "refusing.invalid")
        await session.recoverFromUnauthenticated()
        XCTAssertFalse(session.isSignedIn)
        XCTAssertEqual(session.signInMessage, AuthSession.expiredCopy)
    }

    func testATransientRefreshFailureFailsTheCallAsUnavailable() async throws {
        let session = signedIn(realm: "unreachable.invalid")
        guard case let .failure(error) = try await intercept(session) else {
            return XCTFail("the call went out with the stale token")
        }
        XCTAssertEqual(error.code, .unavailable, "the phone keeps its snapshot (PRD D61)")
    }

    func testARefusedRefreshOnTheRequestPathSignsOut() async throws {
        let session = signedIn(realm: "refusing.invalid")
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
        let session = signedIn(realm: "refusing.invalid")
        let sent = try await intercept(session, path: BearerInterceptor.unauthenticatedPath)
        guard case let .success(request) = sent else {
            return XCTFail("RequestMagicLink was failed before it was sent, so email sign-in cannot start")
        }
        XCTAssertNil(request.headers["Authorization"], "the one procedure sent with no bearer token")
        XCTAssertTrue(session.isSignedIn, "the session was never asked for a token, so nothing refused it")
    }

    func testTheMagicLinkCallBehindTheEdgesAPIPrefixStillGoesOutWithNoHeader() async throws {
        let session = signedIn(realm: "refusing.invalid")
        let edge = "https://mymoneyrecap.test/api"
        let sent = try await intercept(session, base: edge, path: BearerInterceptor.unauthenticatedPath)
        guard case let .success(request) = sent else {
            return XCTFail("behind the edge's /api/ route, RequestMagicLink asked a signed-out session for a token")
        }
        XCTAssertNil(request.headers["Authorization"], "ARCHITECTURE §5 routes /api/ to the BFF")
    }

    func testOnlyTheMagicLinkPathSkipsTheRecoverStepOnAnUnauthenticatedAnswer() async throws {
        let skipped = signedIn(realm: "refusing.invalid")
        let onTheEmailPath = BearerInterceptor(session: skipped)
        _ = try await intercept(skipped, path: BearerInterceptor.unauthenticatedPath, through: onTheEmailPath)
        await respondUnauthenticated(onTheEmailPath)
        XCTAssertTrue(skipped.isSignedIn, "a tokenless call's refusal is not a refused refresh")
        XCTAssertNil(skipped.signInMessage, "else the expired copy waits on Sign in behind Enter email")

        let recovered = signedIn(realm: "refusing.invalid")
        await respondUnauthenticated(BearerInterceptor(session: recovered))
        XCTAssertFalse(recovered.isSignedIn, "every other path still recovers, and signs out when refused")
        XCTAssertEqual(recovered.signInMessage, AuthSession.expiredCopy)
    }

    func testTheKeychainIsRewrittenOnlyWhenARefreshChangedTheSession() async throws {
        Keychain.delete()
        let held = try await signedIn(realm: "unreachable.invalid", expiresIn: 300).freshAccessToken()
        XCTAssertEqual(held, "stale")
        XCTAssertNil(Keychain.read(), "AppAuth handed back the token it held, so there was nothing to write")

        let refreshed = try await signedIn(realm: "refreshing.invalid").freshAccessToken()
        XCTAssertEqual(refreshed, "refreshed")
        let stored = try XCTUnwrap(Keychain.read(), "else a relaunch reloads the spent token")
        let reloaded = try NSKeyedUnarchiver.unarchivedObject(ofClass: OIDAuthState.self, from: stored)
        XCTAssertEqual(reloaded?.lastTokenResponse?.accessToken, "refreshed")
    }

    /// Runs `BearerInterceptor`'s request step alone, where no response interceptor follows a failure.
    private func intercept(
        _ session: AuthSession,
        base: String = BffClient.baseURL,
        path: String = "/money_recap.v1.BffService/GetHome",
        through interceptor: BearerInterceptor? = nil
    ) async throws -> Result<HTTPRequest<Data?>, ConnectError> {
        let request = try HTTPRequest<Data?>(
            url: XCTUnwrap(URL(string: base + path)), headers: [:], message: nil, method: .post,
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

    /// A signed-in state whose access token expires in `expiresIn` seconds; once it has, the next action refreshes
    /// against `realm`.
    private func signedIn(realm: String, expiresIn: Int = 0) -> AuthSession {
        StubRealm.install()
        let endpoint = URL(string: "https://\(realm)/token")!
        let request = OIDAuthorizationRequest(
            configuration: OIDServiceConfiguration(authorizationEndpoint: endpoint, tokenEndpoint: endpoint),
            clientId: AuthConfig.clientID, clientSecret: nil, scopes: AuthConfig.scopes,
            redirectURL: AuthConfig.redirectURL, responseType: OIDResponseTypeCode, additionalParameters: nil
        )
        let authorization = OIDAuthorizationResponse(request: request, parameters: ["code": "code" as NSString])
        let token = OIDTokenResponse(request: authorization.tokenExchangeRequest()!, parameters: [
            "access_token": "stale" as NSString, "expires_in": expiresIn as NSNumber,
            "refresh_token": "refresh" as NSString, "token_type": "Bearer" as NSString,
        ])
        return AuthSession(authState: OIDAuthState(
            authorizationResponse: authorization, tokenResponse: token, registrationResponse: nil
        ))
    }
}

/// The realm's token endpoint: `refusing.invalid` refuses the refresh token (RFC 6749 §5.2), `refreshing.invalid`
/// answers it with the access token `refreshed`, and any other host fails in transit, as with no signal.
final class StubRealm: StubURLProtocol {
    override func startLoading() {
        switch request.url?.host {
        case "refusing.invalid": respond(400, #"{"error":"invalid_grant"}"#)
        case "refreshing.invalid": respond(200, #"{"access_token":"refreshed","expires_in":300,"token_type":"Bearer"}"#)
        default: client?.urlProtocol(self, didFailWithError: URLError(.notConnectedToInternet))
        }
    }
}

/// A realm answered from memory: `install()` routes AppAuth's requests to the subclass's `startLoading`.
class StubURLProtocol: URLProtocol {
    override static func canInit(with _: URLRequest) -> Bool {
        true
    }

    override static func canonicalRequest(for request: URLRequest) -> URLRequest {
        request
    }

    override func stopLoading() {}

    static func install() {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [self]
        OIDURLSessionProvider.setSession(URLSession(configuration: configuration))
    }

    func respond(_ status: Int, _ json: String) {
        guard let url = request.url, let response = HTTPURLResponse(
            url: url, statusCode: status, httpVersion: nil, headerFields: ["Content-Type": "application/json"]
        ) else { return }
        client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
        client?.urlProtocol(self, didLoad: Data(json.utf8))
        client?.urlProtocolDidFinishLoading(self)
    }
}
