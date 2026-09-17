import Foundation

/// The realm's `ios` client (infra/compose/README.md): public, PKCE S256, no secret anywhere in this target.
enum AuthConfig {
    static let issuer = URL(string: "http://localhost:8080/realms/mmr")!
    static let clientID = "ios"
    /// Custom scheme until slice 3 serves universal links; registered in App/Info.plist.
    static let redirectURL = URL(string: "mymoneyrecap://auth/callback")!
    static let scopes = ["openid", "profile", "email"]
}
