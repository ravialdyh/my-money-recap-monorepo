import Connect
import Foundation
import MoneyRecapContract

/// A5 Currency (SCREEN-STATES, PRD §6.5, IOS I6): the platform's common currencies by British English name, searched
/// by name or code, the region's currency chosen where it names one (spec #92). Next saves it with step 2 (A14).
@Observable
@MainActor
final class CurrencyViewModel {
    typealias Save = (MoneyRecap_V1_BffServiceSaveProfileRequest, Headers) async -> SaveResult

    /// One row, as the canvas labels it: "IDR · Indonesian Rupiah".
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

    /// `Locale.commonISOCurrencyCodes`, named and sorted in `names`, British English: today's currencies and a few
    /// recently withdrawn, such as the Croatian Kuna. The platform's longer list adds older money and gold.
    nonisolated static func options(names: Locale = Locale(identifier: "en_GB")) -> [Option] {
        Locale.commonISOCurrencyCodes
            .map { Option(code: $0, name: names.localizedString(forCurrencyCode: $0) ?? $0) }
            .sorted { $0.name.compare($1.name, locale: names) == .orderedAscending }
    }

    let options: [Option]
    var query = ""
    private(set) var selected: String?
    private(set) var isSaving = false
    /// The last save could not reach the BFF: the step says so and Next tries again (`NoConnectionView.copy`).
    private(set) var offline = false
    private var attempt: SaveAttempt<String>?
    private let save: Save

    /// `saved` is the Currency already saved, which starts chosen (spec #92: going back or resuming re-selects it);
    /// empty, the device `region`'s starts chosen when the list has it, and otherwise nothing does.
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

    private var needle: String {
        query.trimmingCharacters(in: .whitespaces)
    }

    /// The rows whose code or name holds the query, ignoring case and accents; every row for an empty query.
    var visible: [Option] {
        guard !needle.isEmpty else { return options }
        return options.filter {
            $0.code.range(of: needle, options: .caseInsensitive) != nil
                || $0.name.range(of: needle, options: [.caseInsensitive, .diacriticInsensitive]) != nil
        }
    }

    /// `No currency matches "{query}".` (SCREEN-STATES A5) while the query matches no row.
    var noMatch: String? {
        visible.isEmpty ? "No currency matches \"\(needle)\"." : nil
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

    /// True once the Currency is saved. A retry of the same choice sends the same key; another choice a new key.
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
