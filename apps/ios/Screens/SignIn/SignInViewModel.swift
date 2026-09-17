import Foundation

/// A1 states (SCREEN-STATES): loading while Keycloak round-trips; `InlineError` on cancel or failure, cleared on the next tap.
@Observable
@MainActor
final class SignInViewModel {
    static let errorCopy = "We couldn't sign you in. Try again."

    private(set) var isLoading = false
    private(set) var errorMessage: String?
    private let signIn: () async throws -> Void

    /// `initialError` is the expired-session copy when a failed token refresh sent the Member back here (I3).
    init(signIn: @escaping () async throws -> Void, initialError: String? = nil) {
        self.signIn = signIn
        self.errorMessage = initialError
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
