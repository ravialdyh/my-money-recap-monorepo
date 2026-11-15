import AppAuth
import Foundation

enum AuthConfig {
    static let issuer = URL(string: ProcessInfo.processInfo
        .environment["MMR_ISSUER"] ?? "http://localhost:8080/realms/mmr")!
    static let clientID = "ios"
    static let redirectURL = URL(string: "mymoneyrecap://auth/callback")!
    static let offlineScope = "offline_access"
    static let scopes = ["openid", "profile", "email", offlineScope]
    static let loginTokenScope = "login-token"
    static let loginTokenParameter = "login_token"

    static func authorizationRequest(
        _ configuration: OIDServiceConfiguration,
        identityProvider: IdentityProvider? = nil
    ) -> OIDAuthorizationRequest {
        var parameters = ["prompt": "login"]
        parameters["kc_idp_hint"] = identityProvider?.rawValue
        return OIDAuthorizationRequest(
            configuration: configuration,
            clientId: clientID,
            clientSecret: nil,
            scopes: scopes,
            redirectURL: redirectURL,
            responseType: OIDResponseTypeCode,
            additionalParameters: parameters
        )
    }

    #if DEBUG
        static var usesAuthorizationCodeFlow: Bool {
            CommandLine.arguments.contains("--authorization-code-flow")
        }
    #else
        static var usesAuthorizationCodeFlow: Bool {
            false
        }
    #endif
}

enum IdentityProvider: String {
    case apple
    case google
}
