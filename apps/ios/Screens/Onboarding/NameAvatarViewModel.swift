import Connect
import Foundation
import MoneyRecapContract

/// A4 Name and avatar (SCREEN-STATES, PRD §6.4, IOS I6): the First name field, prefilled with what the provider
/// shared, and the six Avatars. Next stays off until both answers are valid, and saves them with step 1 in one
/// write, one key per attempt (A14).
@Observable
@MainActor
final class NameAvatarViewModel {
    typealias Save = (MoneyRecap_V1_BffServiceSaveProfileRequest, Headers) async -> SaveResult

    struct Answers: Equatable {
        let name: String
        let avatar: String
    }

    private(set) var name: String
    private(set) var avatar: String?
    private(set) var isSaving = false
    /// The last save could not reach the BFF: the step says so and Next tries again (`NoConnectionView.copy`).
    private(set) var offline = false
    private var attempt: SaveAttempt<Answers>?
    private let save: Save

    init(name: String, avatar: String, save: @escaping Save) {
        self.name = DisplayName.sanitized(name)
        self.avatar = AvatarGrid.keys.contains(avatar) ? avatar : nil
        self.save = save
    }

    func type(_ text: String) {
        name = DisplayName.sanitized(text)
    }

    func choose(_ key: String) {
        avatar = key
    }

    var canContinue: Bool {
        DisplayName.stored(name) != nil && avatar != nil && !isSaving
    }

    /// True once the answers are saved. A retry of the same answers sends the same key; changed ones a new key.
    func next() async -> Bool {
        guard let stored = DisplayName.stored(name), let avatar, !isSaving else { return false }
        let answers = Answers(name: stored, avatar: avatar)
        var current = attempt ?? SaveAttempt(answers)
        current.update(answers)
        attempt = current
        isSaving = true
        defer { isSaving = false }
        let request = MoneyRecap_V1_BffServiceSaveProfileRequest.with {
            $0.displayName = stored
            $0.avatarKey = avatar
            $0.onboardingStep = 1
        }
        let result = await save(request, current.headers)
        offline = result == .offline
        return result == .saved
    }
}
