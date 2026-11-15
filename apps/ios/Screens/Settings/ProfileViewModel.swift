import Foundation
import MoneyRecapContract

@Observable
@MainActor
final class ProfileViewModel {
    private(set) var name: String
    private(set) var avatar: String?
    private(set) var isSaving = false
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
