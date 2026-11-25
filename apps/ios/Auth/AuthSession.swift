import AppAuth
import LocalAuthentication
import SwiftUI

@Observable
@MainActor
final class AuthSession {
    enum Lock: Equatable {
        case signedOut, locked, unlocked
    }

    static let refusedCopy = "Your session expired. Sign in again."
    static let noPasscodeCopy = "Set a passcode in iOS Settings to use My Money Recap."
    static let unlockReason = "Unlock My Money Recap with Face ID."
    static let relockAfter: TimeInterval = 300

    private(set) var lock: Lock

    private(set) var authState: OIDAuthState?
    private(set) var signInMessage: String?
    private(set) var magicLinkFailure: MagicLinkFailure?
    private var flow: OIDExternalUserAgentSession?
    private var redeemed: Set<String> = []
    private let vault: Vault

    init(vault: Vault = .keychain) {
        self.vault = vault
        if CommandLine.arguments.contains("--reset-session") {
            vault.delete()
        }
        lock = vault.exists() ? .locked : .signedOut
    }

    init(authState: OIDAuthState, vault: Vault = .keychain) {
        self.vault = vault
        self.authState = authState
        lock = .unlocked
    }

    var isSignedIn: Bool {
        switch lock {
        case .signedOut: false
        case .locked: true
        case .unlocked: authState?.isAuthorized ?? false
        }
    }

    func unlock() async {
        guard lock == .locked else { return }
        let context = LAContext()
        context.localizedReason = Self.unlockReason
        guard let stored = await vault.read(context) else {
            if !vault.exists() {
                signOut(message: vault.passcodeSet() ? nil : Self.noPasscodeCopy)
            }
            return
        }
        if !stored.locked, await !vault.confirmPresence(context) {
            return
        }
        guard let state = try? NSKeyedUnarchiver.unarchivedObject(ofClass: OIDAuthState.self, from: stored.data) else {
            signOut()
            return
        }
        authState = state
        lock = .unlocked
        if !stored.locked {
            persist()
        }
    }

    func returned(after away: TimeInterval) {
        guard lock == .unlocked, away > Self.relockAfter else { return }
        authState = nil
        lock = .locked
    }

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
        lock = .unlocked
        signInMessage = nil
        persist()
    }

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
            lock = .unlocked
            signInMessage = nil
            persist()
        } catch {
            magicLinkFailure = MagicLink.failure(redeeming: error, alreadyRedeemed: redeemed.contains(token))
        }
    }

    func record(_ failure: MagicLinkFailure?) {
        magicLinkFailure = failure
    }

    func freshAccessToken() async throws -> String {
        guard let authState else { throw CancellationError() }
        let held = authState.lastTokenResponse
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
            if authState.lastTokenResponse !== held {
                persist()
            }
            return token
        } catch {
            if !isSignedIn {
                signOut(message: Self.refusedCopy)
            }
            throw error
        }
    }

    func recoverFromUnauthenticated() async {
        authState?.setNeedsTokenRefresh()
        _ = try? await freshAccessToken()
    }

    func logOut(forget: () async -> Void) async {
        await forget()
        await endRealmSession()
        signOut()
    }

    func signOut(message: String? = nil) {
        signInMessage = message
        authState = nil
        lock = .signedOut
        vault.delete()
    }

    func persist() {
        guard let authState, let data = try? NSKeyedArchiver.archivedData(
            withRootObject: authState,
            requiringSecureCoding: true
        ) else { return }
        guard vault.passcodeSet() else {
            signOut(message: Self.noPasscodeCopy)
            return
        }
        if !vault.write(data) {
            signOut(message: Self.refusedCopy)
        }
    }

    private func endRealmSession() async {
        guard let authState, let refresh = authState.refreshToken,
              let endpoint = authState.lastAuthorizationResponse.request.configuration.endSessionEndpoint
        else { return }
        var form = URLComponents()
        form.queryItems = [
            URLQueryItem(name: "client_id", value: AuthConfig.clientID),
            URLQueryItem(name: "refresh_token", value: refresh),
        ]
        var request = URLRequest(url: endpoint, timeoutInterval: 10)
        request.httpMethod = "POST"
        request.setValue("application/x-www-form-urlencoded", forHTTPHeaderField: "Content-Type")
        request.httpBody = form.percentEncodedQuery.map { Data($0.utf8) }
        _ = try? await OIDURLSessionProvider.session().data(for: request)
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
        try await OIDAuthorizationService.discoverConfiguration(forIssuer: AuthConfig.issuer)
    }

    private static var presenter: UIViewController {
        let scenes = UIApplication.shared.connectedScenes.compactMap { $0 as? UIWindowScene }
        guard let root = scenes.flatMap(\.windows).first(where: \.isKeyWindow)?.rootViewController else {
            preconditionFailure("sign-in needs a key window")
        }
        return root
    }
}
