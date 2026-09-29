import Foundation
import MoneyRecapContract

/// D2 Profile (PRD §6.16): setup's First name field and Avatar grid over the saved answers. Save is enabled once the
/// name is valid and something changed, and sends the name and Avatar alone, one key per attempt (A14).
@Observable
@MainActor
final class ProfileViewModel {
    private(set) var name: String
    private(set) var avatar: String?
    private(set) var isSaving = false
    /// The last save failed: what was typed stays, over `SettingsViewModel.saveFailure`.
    private(set) var failed = false
    private let saved: NameAvatarViewModel.Answers
    private var attempt: SaveAttempt<NameAvatarViewModel.Answers>?
    private let save: NameAvatarViewModel.Save

    init(name: String, avatar: String, save: @escaping NameAvatarViewModel.Save) {
        self.name = DisplayName.sanitized(name)
        self.avatar = AvatarGrid.keys.contains(avatar) ? avatar : nil
        saved = .init(name: name, avatar: avatar)
        self.save = save
    }

    func type(_ text: String) {
        name = DisplayName.sanitized(text)
    }

    func choose(_ key: String) {
        avatar = key
    }

    private var answers: NameAvatarViewModel.Answers? {
        guard let stored = DisplayName.stored(name), let avatar else { return nil }
        return .init(name: stored, avatar: avatar)
    }

    var canSave: Bool {
        answers.map { $0 != saved } ?? false && !isSaving
    }

    /// True once saved. A retry of the same answers sends the same key; changed ones a new key.
    func commit() async -> Bool {
        guard canSave, let answers else { return false }
        var current = attempt ?? SaveAttempt(answers)
        current.update(answers)
        attempt = current
        isSaving = true
        defer { isSaving = false }
        let request = MoneyRecap_V1_BffServiceSaveProfileRequest.with {
            $0.displayName = answers.name
            $0.avatarKey = answers.avatar
        }
        failed = await save(request, current.headers) != .saved
        return !failed
    }
}
