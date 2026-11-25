import AppAuth
import XCTest
@testable import MyMoneyRecap

@MainActor
final class MagicLinkSessionTests: XCTestCase {
    override func setUp() {
        super.setUp()
        Keychain.delete()
        StubLoginTokenRealm.refused = []
        StubLoginTokenRealm.install()
    }

    override func tearDown() {
        OIDURLSessionProvider.setSession(.shared)
        Keychain.delete()
        super.tearDown()
    }

    func testALinkSignsInWithAStateThatCanRefresh() async throws {
        let session = AuthSession()
        try await session.openMagicLink(XCTUnwrap(URL(string: "mymoneyrecap://magic?lt=lt%3Aone")))
        XCTAssertTrue(session.isSignedIn)
        XCTAssertNil(session.magicLinkFailure)
        let state = try XCTUnwrap(session.authState)
        XCTAssertNotNil(
            (state as NSObject).value(forKey: "lastAuthorizationResponse"),
            "the synthesized response every refresh is built from"
        )
        XCTAssertEqual(state.refreshToken, "refresh", "applied after the synthesized response, not cleared by it")
    }

    func testTheSameLinkTwiceIsAlreadyUsedAndAnyOtherRefusalIsExpired() async throws {
        let session = AuthSession()
        let link = try XCTUnwrap(URL(string: "mymoneyrecap://magic?lt=lt%3Aone"))
        await session.openMagicLink(link)
        session.signOut()
        await session.openMagicLink(link)
        XCTAssertEqual(session.magicLinkFailure, .alreadyUsed, "this session redeemed that token itself")
        StubLoginTokenRealm.refused.insert("lt:stale")
        try await session.openMagicLink(XCTUnwrap(URL(string: "mymoneyrecap://magic?lt=lt%3Astale")))
        XCTAssertEqual(session.magicLinkFailure, .expired, "refused, but never redeemed by this session")
    }

    func testALinkWithNoLoginTokenIsTheGenericCopy() async throws {
        let session = AuthSession()
        try await session.openMagicLink(XCTUnwrap(URL(string: "mymoneyrecap://magic")))
        XCTAssertEqual(session.magicLinkFailure, .generic)
        XCTAssertFalse(session.isSignedIn)
    }

    func testTheArchivedStateStillBuildsARefreshWithAnEndpointAndAClientID() async throws {
        let endpoint = try XCTUnwrap(URL(string: "https://realm.invalid/token"))
        let configuration = OIDServiceConfiguration(authorizationEndpoint: endpoint, tokenEndpoint: endpoint)
        let redeemed = OIDTokenResponse(
            request: MagicLink.tokenRequest(configuration, loginToken: "lt:token"),
            parameters: [
                "access_token": "access" as NSString, "expires_in": 300 as NSNumber,
                "refresh_token": "refresh" as NSString, "token_type": "Bearer" as NSString,
            ]
        )
        let state = MagicLink.authState(configuration, tokenResponse: redeemed)
        AuthSession(authState: state).persist(state)

        let relaunched = AuthSession()
        await relaunched.unlock()
        let reloaded = try XCTUnwrap(relaunched.authState, "the archived magic-link state did not come back")
        _ = try XCTUnwrap(reloaded.refreshToken, "applied before the synthesized response, which clears it")
        let refresh = try XCTUnwrap(reloaded.tokenRefreshRequest()) as NSObject
        XCTAssertNotNil(
            refresh.value(forKeyPath: "configuration.tokenEndpoint"),
            "a refresh built from a token response alone has nowhere to go"
        )
        XCTAssertNotNil(
            refresh.value(forKey: "clientID"),
            "a nil client id raises an uncatchable exception when the refresh body is built"
        )
    }
}

final class StubLoginTokenRealm: StubURLProtocol {
    nonisolated(unsafe) static var refused: Set<String> = []

    override func startLoading() {
        guard let url = request.url else { return }
        let (status, body) = answer(for: url)
        respond(status, body)
    }

    private func answer(for url: URL) -> (Int, String) {
        let issuer = AuthConfig.issuer.absoluteString
        if url.path.hasSuffix("openid-configuration") {
            return (200, """
            {"issuer":"\(issuer)","authorization_endpoint":"\(issuer)/auth",
            "token_endpoint":"\(issuer)/token","jwks_uri":"\(issuer)/certs",
            "response_types_supported":["code"],"subject_types_supported":["public"],
            "id_token_signing_alg_values_supported":["RS256"]}
            """)
        }
        guard let token = loginToken, Self.refused.insert(token).inserted else {
            return (401, #"{"error":"invalid_grant","error_description":"Invalid or expired login token"}"#)
        }
        return (200, #"{"access_token":"access","expires_in":300,"refresh_token":"refresh","token_type":"Bearer"}"#)
    }

    private var loginToken: String? {
        guard let stream = request.httpBodyStream else { return nil }
        stream.open()
        defer { stream.close() }
        var body = Data()
        var chunk = [UInt8](repeating: 0, count: 1024)
        while case let count = stream.read(&chunk, maxLength: chunk.count), count > 0 {
            body.append(chunk, count: count)
        }
        guard let form = String(bytes: body, encoding: .utf8) else { return nil }
        return URLComponents(string: "?" + form)?.queryItems?.first { $0.name == AuthConfig.loginTokenParameter }?.value
    }
}
