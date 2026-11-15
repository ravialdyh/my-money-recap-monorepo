import Foundation

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

@Observable
@MainActor
final class SignInViewModel {
    static let errorCopy = MagicLinkFailure.generic.copy

    private(set) var tapped: SignInButton?
    private(set) var errorMessage: String?
    private let signIn: (SignInButton) async throws -> Void

    init(signIn: @escaping (SignInButton) async throws -> Void, initialError: String? = nil) {
        self.signIn = signIn
        errorMessage = initialError
    }

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

    func dims(_ button: SignInButton) -> Bool {
        tapped != nil && tapped != button
    }
}
