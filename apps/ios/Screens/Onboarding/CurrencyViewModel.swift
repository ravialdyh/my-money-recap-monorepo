import Connect
import Foundation
import MoneyRecapContract

@Observable
@MainActor
final class CurrencyViewModel {
    typealias Save = (MoneyRecap_V1_BffServiceSaveProfileRequest, Headers) async -> SaveResult

    struct Option: Equatable, Identifiable {
        let code: String
        let name: String

        var id: String {
            code
        }

        var label: String {
            "\(code) · \(name)"
        }
    }

    nonisolated static func options(names: Locale = Locale(identifier: "en_GB")) -> [Option] {
        Locale.commonISOCurrencyCodes
            .map { Option(code: $0, name: names.localizedString(forCurrencyCode: $0) ?? $0) }
            .sorted { $0.name.compare($1.name, locale: names) == .orderedAscending }
    }

    let options: [Option]
    var query = ""
    private(set) var selected: String?
    private(set) var isSaving = false
    private(set) var offline = false
    private var attempt: SaveAttempt<String>?
    private let save: Save

    init(
        saved: String = "",
        region: Locale = .autoupdatingCurrent,
        options: [Option] = CurrencyViewModel.options(),
        save: @escaping Save
    ) {
        self.options = options
        let own = region.currency?.identifier
        let listed = options.contains { $0.code == own } ? own : nil
        selected = saved.isEmpty ? listed : saved
        self.save = save
    }

    var visible: [Option] {
        Self.matching(query, in: options)
    }

    var noMatch: String? {
        Self.noMatch(query, in: options)
    }

    nonisolated static func matching(_ query: String, in options: [Option]) -> [Option] {
        let needle = query.trimmingCharacters(in: .whitespaces)
        guard !needle.isEmpty else { return options }
        return options.filter {
            $0.code.range(of: needle, options: .caseInsensitive) != nil
                || $0.name.range(of: needle, options: [.caseInsensitive, .diacriticInsensitive]) != nil
        }
    }

    nonisolated static func noMatch(_ query: String, in options: [Option]) -> String? {
        let needle = query.trimmingCharacters(in: .whitespaces)
        return matching(needle, in: options).isEmpty ? "No currency matches \"\(needle)\"." : nil
    }

    func choose(_ code: String) {
        selected = code
    }

    func clearSearch() {
        query = ""
    }

    var canContinue: Bool {
        selected != nil && !isSaving
    }

    func next() async -> Bool {
        guard let selected, !isSaving else { return false }
        var current = attempt ?? SaveAttempt(selected)
        current.update(selected)
        attempt = current
        isSaving = true
        defer { isSaving = false }
        let request = MoneyRecap_V1_BffServiceSaveProfileRequest.with {
            $0.currencyCode = selected
            $0.onboardingStep = 2
        }
        let result = await save(request, current.headers)
        offline = result == .offline
        return result == .saved
    }
}
