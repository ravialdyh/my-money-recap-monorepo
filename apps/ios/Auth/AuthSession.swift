import AppAuth
import SwiftUI

/// AppAuth-iOS drives `ASWebAuthenticationSession` with PKCE against endpoints discovered from the
/// realm's well-known document (IOS I1). The resulting `OIDAuthState` lives in the Keychain (I2).
@Observable
@MainActor
final class AuthSession {
    /// Copy Sign in shows after a token refresh failed (IOS I3): the session is gone, not the tap.
    static let expiredCopy = "Your session expired. Sign in again."

    private(set) var authState: OIDAuthState?
    /// Set when sign-out was not the Member's choice, read once by Sign in.
    private(set) var signInMessage: String?
    /// AppAuth needs the in-flight session retained until the redirect arrives.
    private var flow: OIDExternalUserAgentSession?

    init() {
        if CommandLine.arguments.contains("--reset-session") {
            Keychain.delete()
        }
        if let data = Keychain.read() {
            authState = try? NSKeyedUnarchiver.unarchivedObject(ofClass: OIDAuthState.self, from: data)
        }
    }

    var isSignedIn: Bool {
        authState?.isAuthorized ?? false
    }

    /// `prompt=login` makes a Sign in tap show the login page even with a Keycloak SSO cookie in the
    /// web sheet; `OIDAuthorizationRequest` generates the PKCE verifier and S256 challenge (IOS I1).
    func signIn() async throws {
        let configuration = try await discover()
        let request = OIDAuthorizationRequest(
            configuration: configuration,
            clientId: AuthConfig.clientID,
            clientSecret: nil,
            scopes: AuthConfig.scopes,
            redirectURL: AuthConfig.redirectURL,
            responseType: OIDResponseTypeCode,
            additionalParameters: ["prompt": "login"]
        )
        let state: OIDAuthState = try await withCheckedThrowingContinuation { continuation in
            flow = OIDAuthState.authState(byPresenting: request, presenting: Self.presenter) { state, error in
                if let state {
                    continuation.resume(returning: state)
                } else {
                    continuation.resume(throwing: error ?? CancellationError())
                }
            }
        }
        flow = nil
        authState = state
        signInMessage = nil
        persist()
    }

    /// A current access token, refreshed by AppAuth first when it has expired (I1, I3).
    func freshAccessToken() async throws -> String {
        guard let authState else { throw CancellationError() }
        let token: String = try await withCheckedThrowingContinuation { continuation in
            authState.performAction { accessToken, _, error in
                if let accessToken {
                    continuation.resume(returning: accessToken)
                } else {
                    continuation.resume(throwing: error ?? CancellationError())
                }
            }
        }
        persist()
        return token
    }

    /// The BFF rejected the token: force a refresh; if that fails too, back to Sign in with the inline error (I3).
    func recoverFromUnauthenticated() async {
        authState?.setNeedsTokenRefresh()
        do { _ = try await freshAccessToken() } catch { signOut(message: Self.expiredCopy) }
    }

    /// Erases the session; the cache goes with it in `RootView` (PRD §7 "on device only", I4).
    func signOut(message: String? = nil) {
        signInMessage = message
        authState = nil
        Keychain.delete()
    }

    private func persist() {
        guard let authState, let data = try? NSKeyedArchiver.archivedData(
            withRootObject: authState,
            requiringSecureCoding: true
        ) else { return }
        Keychain.write(data)
    }

    private func discover() async throws -> OIDServiceConfiguration {
        try await withCheckedThrowingContinuation { continuation in
            OIDAuthorizationService.discoverConfiguration(forIssuer: AuthConfig.issuer) { configuration, error in
                if let configuration {
                    continuation.resume(returning: configuration)
                } else {
                    continuation.resume(throwing: error ?? CancellationError())
                }
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
