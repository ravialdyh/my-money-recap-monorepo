import Foundation

/// A1 states (SCREEN-STATES): loading while Keycloak round-trips; `InlineError` on cancel or failure, cleared on the next tap.
@Observable
@MainActor
final class SignInViewModel {
    static let errorCopy = "We couldn't sign you in. Try again."

    private(set) var isLoading = false
    private(set) var errorMessage: String?
    private let signIn: () async throws -> Void

    init(signIn: @escaping () async throws -> Void) {
        self.signIn = signIn
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
