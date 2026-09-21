import AppAuth
import Foundation

/// The realm's `ios` client (infra/compose/README.md): public, PKCE S256, no secret anywhere in this target.
enum AuthConfig {
    /// Local Compose Keycloak. `MMR_ISSUER` overrides it for a physical device reaching the Mac over the LAN
    /// (`infra/compose/README.md` "A phone on the LAN"), the same seam as `MMR_BFF_URL`.
    static let issuer = URL(string: ProcessInfo.processInfo
        .environment["MMR_ISSUER"] ?? "http://localhost:8080/realms/mmr")!
    static let clientID = "ios"
    /// The custom scheme carries the authorization callback and, today, the magic link; a universal link
    /// joins it once a domain exists. Registered in App/Info.plist.
    static let redirectURL = URL(string: "mymoneyrecap://auth/callback")!
    static let scopes = ["openid", "profile", "email"]
    /// A10's direct grant: the optional client scope that selects the login-token authenticator, and the
    /// field the opaque `lt:<uuid>` travels in (ARCHITECTURE §7.5).
    static let loginTokenScope = "login-token"
    static let loginTokenParameter = "login_token"

    /// `prompt=login` makes a Sign in tap show the login page even with a Keycloak SSO cookie in the web
    /// sheet; the request generates the PKCE verifier and S256 challenge (IOS I1). The magic-link state is
    /// synthesized around this same request (spec #61).
    static func authorizationRequest(_ configuration: OIDServiceConfiguration) -> OIDAuthorizationRequest {
        OIDAuthorizationRequest(
            configuration: configuration,
            clientId: clientID,
            clientSecret: nil,
            scopes: scopes,
            redirectURL: redirectURL,
            responseType: OIDResponseTypeCode,
            additionalParameters: ["prompt": "login"]
        )
    }

    #if DEBUG
        /// The UI test's seam onto the authorization-code flow (spec #61 seam 6): the PKCE request, the web
        /// sheet and the state-from-authorization-response path keep a test once Continue with email stops
        /// opening a browser.
        static var usesAuthorizationCodeFlow: Bool {
            CommandLine.arguments.contains("--authorization-code-flow")
        }
    #else
        static var usesAuthorizationCodeFlow: Bool {
            false
        }
    #endif
}
