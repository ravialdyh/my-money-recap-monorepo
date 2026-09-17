import Foundation

/// The realm's `ios` client (infra/compose/README.md): public, PKCE S256, no secret anywhere in this target.
enum AuthConfig {
    /// Local Compose Keycloak. `MMR_ISSUER` overrides it for a physical device reaching the Mac over the LAN
    /// (`infra/compose/README.md` "A phone on the LAN"), the same seam as `MMR_BFF_URL`.
    static let issuer = URL(string: ProcessInfo.processInfo.environment["MMR_ISSUER"] ?? "http://localhost:8080/realms/mmr")!
    static let clientID = "ios"
    /// Custom scheme until slice 3 serves universal links; registered in App/Info.plist.
    static let redirectURL = URL(string: "mymoneyrecap://auth/callback")!
    static let scopes = ["openid", "profile", "email"]
}
