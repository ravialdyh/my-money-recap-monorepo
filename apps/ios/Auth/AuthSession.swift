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
    /// A refused login token, which returns the Member to Enter email carrying its copy (PRD §6.3).
    private(set) var magicLinkFailure: MagicLinkFailure?
    /// AppAuth needs the in-flight session retained until the redirect arrives.
    private var flow: OIDExternalUserAgentSession?
    /// The login tokens redeemed successfully in this session: the only thing that tells an already-used
    /// link from an expired one, since both are `invalid_grant` (spec #61).
    private var redeemed: Set<String> = []

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

    /// The authorization-code flow of IOS I1: Apple and Google through their `kc_idp_hint`, and Keycloak's own
    /// page with none, which a debug launch argument drives for the UI test (spec #61 seam 6).
    func signIn(identityProvider: IdentityProvider? = nil) async throws {
        let configuration = try await discover()
        let request = AuthConfig.authorizationRequest(configuration, identityProvider: identityProvider)
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

    /// Where all three incoming-link sources land (spec #61): the login token is redeemed by one direct
    /// grant, and the session is built from a synthesized authorization response so it can refresh.
    func openMagicLink(_ url: URL) async {
        magicLinkFailure = nil
        guard let token = MagicLink.loginToken(in: url) else {
            magicLinkFailure = .generic
            return
        }
        do {
            let configuration = try await discover()
            let response = try await redeem(MagicLink.tokenRequest(configuration, loginToken: token))
            redeemed.insert(token)
            authState = MagicLink.authState(configuration, tokenResponse: response)
            signInMessage = nil
            persist()
        } catch {
            magicLinkFailure = MagicLink.failure(redeeming: error, alreadyRedeemed: redeemed.contains(token))
        }
    }

    /// The copy Enter email shows: a refused login token, or a request either screen could not send.
    func record(_ failure: MagicLinkFailure?) {
        magicLinkFailure = failure
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

    /// The archiver the Keychain item is written with (I2); the magic-link state's round trip is a test.
    func persist() {
        guard let authState, let data = try? NSKeyedArchiver.archivedData(
            withRootObject: authState,
            requiringSecureCoding: true
        ) else { return }
        Keychain.write(data)
    }

    private func redeem(_ request: OIDTokenRequest) async throws -> OIDTokenResponse {
        try await withCheckedThrowingContinuation { continuation in
            OIDAuthorizationService.perform(request) { response, error in
                if let response {
                    continuation.resume(returning: response)
                } else {
                    continuation.resume(throwing: error ?? CancellationError())
                }
            }
        }
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
