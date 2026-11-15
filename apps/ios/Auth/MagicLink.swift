import AppAuth
import Connect
import Foundation
import MoneyRecapContract

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

enum MagicLink {
    static let tokenQuery = "lt"
    static let launchArgument = "--magic-link"

    static func loginToken(in url: URL) -> String? {
        URLComponents(url: url, resolvingAgainstBaseURL: false)?
            .queryItems?.first { $0.name == tokenQuery }?.value
    }

    static var launchedLink: URL? {
        let arguments = CommandLine.arguments
        guard let flag = arguments.firstIndex(of: launchArgument), arguments.indices.contains(flag + 1) else {
            return nil
        }
        return URL(string: arguments[flag + 1])
    }

    static func tokenRequest(_ configuration: OIDServiceConfiguration, loginToken: String) -> OIDTokenRequest {
        OIDTokenRequest(
            configuration: configuration,
            grantType: OIDGrantTypePassword,
            authorizationCode: nil,
            redirectURL: nil,
            clientID: AuthConfig.clientID,
            clientSecret: nil,
            scopes: [AuthConfig.loginTokenScope, AuthConfig.offlineScope],
            refreshToken: nil,
            codeVerifier: nil,
            additionalParameters: [AuthConfig.loginTokenParameter: loginToken]
        )
    }

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

    static func failure(redeeming error: Error, alreadyRedeemed: Bool) -> MagicLinkFailure {
        let error = error as NSError
        guard error.domain == OIDOAuthTokenErrorDomain,
              error.code == OIDErrorCodeOAuthToken.invalidGrant.rawValue
        else { return .generic }
        return alreadyRedeemed ? .alreadyUsed : .expired
    }

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

    static func request(
        _ email: String,
        client: MoneyRecap_V1_BffServiceClient
    ) async -> MagicLinkFailure? {
        await failure(requesting: client.requestMagicLink(request: .with { $0.email = email }).error)
    }

    static func failure(requesting error: ConnectError?) -> MagicLinkFailure? {
        guard let error else { return nil }
        return error.code == .resourceExhausted ? .rateLimited : .generic
    }
}
