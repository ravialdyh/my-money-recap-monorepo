import Foundation

/// A1 states (SCREEN-STATES): loading while Keycloak round-trips; `InlineError` on cancel or failure, cleared on the
/// next tap.
@Observable
@MainActor
final class SignInViewModel {
    /// One string: the email path shows the same words for every failure it cannot name (PRD §6.3).
    static let errorCopy = MagicLinkFailure.generic.copy

    private(set) var isLoading = false
    private(set) var errorMessage: String?
    private let signIn: () async throws -> Void

    /// `initialError` is the expired-session copy when the realm refused a token refresh (I3).
    init(signIn: @escaping () async throws -> Void, initialError: String? = nil) {
        self.signIn = signIn
        errorMessage = initialError
    }

    func continueWithEmail() async {
        errorMessage = nil
        isLoading = true
        defer { isLoading = false }
        do {
            try await signIn()
        } catch {
            errorMessage = Self.errorCopy
        }
    }
}
