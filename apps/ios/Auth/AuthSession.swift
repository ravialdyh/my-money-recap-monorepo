import AppAuth
import SwiftUI

/// AppAuth-iOS drives `ASWebAuthenticationSession` with PKCE against endpoints discovered from the
/// realm's well-known document (IOS I1). The resulting `OIDAuthState` lives in the Keychain (I2).
@Observable
@MainActor
final class AuthSession {
    private(set) var authState: OIDAuthState?
    /// AppAuth needs the in-flight session retained until the redirect arrives.
    private var flow: OIDExternalUserAgentSession?

    init() {
        // Test seam: the UI test starts from a clean Sign in on a simulator that may hold a session.
        if CommandLine.arguments.contains("--reset-session") { Keychain.delete() }
        if let data = Keychain.read() {
            authState = try? NSKeyedUnarchiver.unarchivedObject(ofClass: OIDAuthState.self, from: data)
        }
    }

    var isSignedIn: Bool { authState?.isAuthorized ?? false }

    var email: String? {
        guard let raw = authState?.lastTokenResponse?.idToken, let token = OIDIDToken(idTokenString: raw) else { return nil }
        return token.claims["email"] as? String
    }

    func signIn() async throws {
        let configuration = try await discover()
        let request = OIDAuthorizationRequest(
            configuration: configuration,
            clientId: AuthConfig.clientID,
            clientSecret: nil,
            scopes: AuthConfig.scopes,
            redirectURL: AuthConfig.redirectURL,
            responseType: OIDResponseTypeCode,
            additionalParameters: ["prompt": "login"] // a Sign in tap always shows the login page, even with a Keycloak SSO cookie in the web sheet
        ) // this initializer generates the PKCE verifier and S256 challenge
        let state: OIDAuthState = try await withCheckedThrowingContinuation { continuation in
            flow = OIDAuthState.authState(byPresenting: request, presenting: Self.presenter) { state, error in
                if let state { continuation.resume(returning: state) } else { continuation.resume(throwing: error ?? CancellationError()) }
            }
        }
        flow = nil
        authState = state
        if let data = try? NSKeyedArchiver.archivedData(withRootObject: state, requiringSecureCoding: true) {
            Keychain.write(data)
        }
    }

    private func discover() async throws -> OIDServiceConfiguration {
        try await withCheckedThrowingContinuation { continuation in
            OIDAuthorizationService.discoverConfiguration(forIssuer: AuthConfig.issuer) { configuration, error in
                if let configuration { continuation.resume(returning: configuration) } else { continuation.resume(throwing: error ?? CancellationError()) }
            }
        }
    }

    private static var presenter: UIViewController {
        let scenes = UIApplication.shared.connectedScenes.compactMap { $0 as? UIWindowScene }
        guard let root = scenes.flatMap(\.windows).first(where: \.isKeyWindow)?.rootViewController else {
            preconditionFailure("sign-in needs a key window")
        }
        return root
    }
}
