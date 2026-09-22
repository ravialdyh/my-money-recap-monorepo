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

    /// PKCE S256 (IOS I1). `kc_idp_hint` hands the sheet straight to Apple or Google (#71); `prompt=login` keeps a
    /// Keycloak SSO cookie from signing the last Member in silently instead. The magic-link state is built around it.
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
        /// The UI test's seam onto the authorization-code flow (spec #61 seam 6): the PKCE request, the web
        /// sheet and the state-from-authorization-response path keep a test now that Continue with email opens
        /// no browser.
        static var usesAuthorizationCodeFlow: Bool {
            CommandLine.arguments.contains("--authorization-code-flow")
        }
    #else
        static var usesAuthorizationCodeFlow: Bool {
            false
        }
    #endif
}

/// The realm's identity-provider aliases (`infra/tofu/realm/identity-providers.tf`), sent as `kc_idp_hint`.
enum IdentityProvider: String {
    case apple
    case google
}
