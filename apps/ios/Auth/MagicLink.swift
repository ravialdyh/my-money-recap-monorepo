import AppAuth
import Connect
import Foundation
import MoneyRecapContract

/// The four copies SCREEN-STATES A3 names, chosen by failure code (PRD §6.3, spec #61). Keycloak answers
/// `invalid_grant` for both an expired and an already-used link, so the caller passes which one it is.
enum MagicLinkFailure: Equatable {
    case expired
    case alreadyUsed
    case rateLimited
    case generic

    var copy: String {
        switch self {
        case .expired: "That link has expired. Links work for 15 minutes."
        case .alreadyUsed: "That link has already been used. We can send another."
        case .rateLimited: "That's five links in an hour. Try again later."
        case .generic: "We couldn't sign you in. Try again."
        }
    }
}

/// The email path's one incoming-link handler (spec #61): the universal link once a domain exists, the
/// custom scheme today and the UI test's launch argument all carry the login token as the `lt` query, and
/// all three run the same direct grant.
enum MagicLink {
    static let tokenQuery = "lt"
    static let launchArgument = "--magic-link"

    static func loginToken(in url: URL) -> String? {
        URLComponents(url: url, resolvingAgainstBaseURL: false)?
            .queryItems?.first { $0.name == tokenQuery }?.value
    }

    /// The UI test's third source, read once at launch: `--magic-link <url>`.
    static var launchedLink: URL? {
        let arguments = CommandLine.arguments
        guard let flag = arguments.firstIndex(of: launchArgument), arguments.indices.contains(flag + 1) else {
            return nil
        }
        return URL(string: arguments[flag + 1])
    }

    /// A10's direct grant: one token request carrying the login token, no browser anywhere (PRD §6.1, I1).
    static func tokenRequest(_ configuration: OIDServiceConfiguration, loginToken: String) -> OIDTokenRequest {
        OIDTokenRequest(
            configuration: configuration,
            grantType: OIDGrantTypePassword,
            authorizationCode: nil,
            redirectURL: nil,
            clientID: AuthConfig.clientID,
            clientSecret: nil,
            scopes: [AuthConfig.loginTokenScope],
            refreshToken: nil,
            codeVerifier: nil,
            additionalParameters: [AuthConfig.loginTokenParameter: loginToken]
        )
    }

    /// AppAuth builds every refresh from the last authorization response's token endpoint and client id, so the
    /// token response goes onto a synthesized one, second: applying an authorization response clears it (IOS I1,
    /// spec #61).
    static func authState(
        _ configuration: OIDServiceConfiguration,
        tokenResponse: OIDTokenResponse
    ) -> OIDAuthState {
        OIDAuthState(
            authorizationResponse: OIDAuthorizationResponse(
                request: AuthConfig.authorizationRequest(configuration),
                parameters: [:]
            ),
            tokenResponse: tokenResponse,
            registrationResponse: nil
        )
    }

    /// `alreadyRedeemed` is whether this session has seen a successful redemption of the same token; the
    /// expired copy is the default, because the code alone cannot tell the two apart (spec #61).
    static func failure(redeeming error: Error, alreadyRedeemed: Bool) -> MagicLinkFailure {
        let error = error as NSError
        guard error.domain == OIDOAuthTokenErrorDomain,
              error.code == OIDErrorCodeOAuthToken.invalidGrant.rawValue
        else { return .generic }
        return alreadyRedeemed ? .alreadyUsed : .expired
    }

    /// One sender for both screens, so a refusal on Check your inbox lands on Enter email carrying its
    /// copy, which is where PRD §6.3 puts all three failures.
    @MainActor
    static func sender(_ session: AuthSession) -> (String) async -> MagicLinkFailure? {
        let client = BffClient.make(session: session)
        return { email in
            session.record(nil)
            let failure = await request(email, client: client)
            session.record(failure)
            return failure
        }
    }

    /// `RequestMagicLink` is sent with no bearer token at all (spec #61).
    static func request(
        _ email: String,
        client: MoneyRecap_V1_BffServiceClient
    ) async -> MagicLinkFailure? {
        await failure(requesting: client.requestMagicLink(request: .with { $0.email = email }).error)
    }

    /// `resource_exhausted` is PRD §11's five per address per hour, drawn as the A3 rate-limited copy.
    static func failure(requesting error: ConnectError?) -> MagicLinkFailure? {
        guard let error else { return nil }
        return error.code == .resourceExhausted ? .rateLimited : .generic
    }
}
