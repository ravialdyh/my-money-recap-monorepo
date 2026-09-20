import AppAuth
import SwiftUI

/// AppAuth-iOS drives `ASWebAuthenticationSession` with PKCE against endpoints discovered from the
/// realm's well-known document (IOS I1). The resulting `OIDAuthState` lives in the Keychain (I2).
@Observable
@MainActor
final class AuthSession {
    /// Copy Sign in shows after the realm refused a token refresh (IOS I3): the session is gone, not the tap.
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

    /// Starts from a given state instead of the Keychain; the tests' seam.
    init(authState: OIDAuthState) {
        self.authState = authState
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

    /// The access token, refreshed if expired (I1, I3). Any refresh error throws, as AppAuth may pass the stale token
    /// too. Only the realm's refusal signs out; a network error, a realm 5xx, a garbled answer or a failed ID token
    /// check keeps the session (PRD D61).
    func freshAccessToken() async throws -> String {
        guard let authState else { throw CancellationError() }
        do {
            let token: String = try await withCheckedThrowingContinuation { continuation in
                authState.performAction { accessToken, _, error in
                    if let error {
                        continuation.resume(throwing: error)
                    } else if let accessToken {
                        continuation.resume(returning: accessToken)
                    } else {
                        continuation.resume(throwing: CancellationError())
                    }
                }
            }
            persist()
            return token
        } catch {
            if !isSignedIn {
                signOut(message: Self.expiredCopy)
            }
            throw error
        }
    }

    /// The BFF rejected the token: force a refresh, which ends the session only if the realm refuses it (I3).
    func recoverFromUnauthenticated() async {
        authState?.setNeedsTokenRefresh()
        _ = try? await freshAccessToken()
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
