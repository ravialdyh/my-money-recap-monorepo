import Foundation

/// A1's three buttons, Apple first (PRD §6.1). Email opens no web sheet except under the debug launch argument,
/// where it opens Keycloak's own page (spec #61 seam 6).
enum SignInButton {
    case apple
    case google
    case email

    var identityProvider: IdentityProvider? {
        switch self {
        case .apple: .apple
        case .google: .google
        case .email: nil
        }
    }
}

/// A1 states (SCREEN-STATES): the tapped button spins and the other two dim while Keycloak round-trips; a cancel or
/// failure is an `InlineError` with all three live again, cleared on the next tap.
@Observable
@MainActor
final class SignInViewModel {
    /// One string: the email path shows the same words for every failure it cannot name (PRD §6.3).
    static let errorCopy = MagicLinkFailure.generic.copy

    private(set) var tapped: SignInButton?
    private(set) var errorMessage: String?
    private let signIn: (SignInButton) async throws -> Void

    /// `initialError` is the expired-session copy when the realm refused a token refresh (I3).
    init(signIn: @escaping (SignInButton) async throws -> Void, initialError: String? = nil) {
        self.signIn = signIn
        errorMessage = initialError
    }

    /// A tap while another round-trip is in flight does nothing, so only one web sheet is ever open.
    func tap(_ button: SignInButton) async {
        guard tapped == nil else { return }
        errorMessage = nil
        tapped = button
        defer { tapped = nil }
        do {
            try await signIn(button)
        } catch {
            errorMessage = Self.errorCopy
        }
    }

    /// SCREEN-STATES A1 loading: every button but the tapped one sits at 40%.
    func dims(_ button: SignInButton) -> Bool {
        tapped != nil && tapped != button
    }
}
