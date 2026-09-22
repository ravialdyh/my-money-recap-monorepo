import AppAuth
import XCTest
@testable import MyMoneyRecap

/// The authorization request Apple and Google ride on (#71): the hint that skips Keycloak's page, the `prompt=login`
/// that keeps an SSO cookie from signing the last Member in instead, and the 30-day offline session (PRD D62).
final class AuthConfigTests: XCTestCase {
    private let configuration = OIDServiceConfiguration(
        authorizationEndpoint: URL(string: "http://localhost:8080/realms/mmr/protocol/openid-connect/auth")!,
        tokenEndpoint: URL(string: "http://localhost:8080/realms/mmr/protocol/openid-connect/token")!
    )

    func testAppleAndGoogleEachHintTheirOwnProvider() throws {
        for provider in [IdentityProvider.apple, .google] {
            let request = AuthConfig.authorizationRequest(configuration, identityProvider: provider)
            let url = request.authorizationRequestURL()
            let query = try XCTUnwrap(URLComponents(url: url, resolvingAgainstBaseURL: false)?.queryItems)
            XCTAssertEqual(query.first { $0.name == "kc_idp_hint" }?.value, provider.rawValue)
            XCTAssertEqual(query.first { $0.name == "prompt" }?.value, "login")
            XCTAssertEqual(query.first { $0.name == "code_challenge_method" }?.value, "S256")
        }
    }

    func testTheAliasesAreTheRealmsOwn() {
        XCTAssertEqual(IdentityProvider.apple.rawValue, "apple", "infra/tofu/realm/identity-providers.tf")
        XCTAssertEqual(IdentityProvider.google.rawValue, "google", "the Google provider's fixed alias")
    }

    func testEveryWayInAsksForTheThirtyDaySession() throws {
        let request = AuthConfig.authorizationRequest(configuration, identityProvider: .google)
        XCTAssertTrue(try XCTUnwrap(request.scope).split(separator: " ").contains("offline_access"), "PRD D62")
        let direct = MagicLink.tokenRequest(configuration, loginToken: "lt:f891f3fb-dc63-4e62-92e3-cd0b948b93b8")
        XCTAssertEqual(direct.scope, "login-token offline_access", "the magic link keeps its sign-in as long")
    }

    func testNoProviderMeansNoHint() {
        let request = AuthConfig.authorizationRequest(configuration)
        XCTAssertEqual(request.additionalParameters, ["prompt": "login"], "Keycloak's own page, the debug seam")
    }
}
