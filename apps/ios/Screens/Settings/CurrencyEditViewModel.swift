import Connect
import Foundation
import MoneyRecapContract

/// D5 to D7 Currency (PRD §8.9): another code is previewed, then changes at once with nothing to convert, or opens
/// D6 with a fetched rate, its failed fetch, or D7's typed rate. Every figure is ledger's (ARCHITECTURE A2).
@Observable
@MainActor
final class CurrencyEditViewModel: Identifiable {
    typealias Preview = (MoneyRecap_V1_BffServicePreviewCurrencyChangeRequest) async
        -> Result<MoneyRecap_V1_BffServicePreviewCurrencyChangeResponse, ConnectError>
    typealias Change = (MoneyRecap_V1_BffServiceChangeCurrencyRequest, Headers) async
        -> Result<MoneyRecap_V1_BffServiceChangeCurrencyResponse, ConnectError>

    /// Which confirm sheet is open: D6 with a fetched rate, D6 when the fetch failed, or D7.
    enum Step: Equatable {
        case confirm, failed, manual
    }

    static let changed = "Changed ✓"
    static let storeLimit = "That's more than the app can store."
    static let convertFailure = "We couldn't convert your amounts. Try again."

    let options: [CurrencyViewModel.Option]
    var query = ""
    /// The Currency every amount is held in.
    private(set) var current: String
    /// `Changed ✓` beside the current row after a direct change, until the list next opens (a new model).
    private(set) var changedValue: String?
    /// The code the open sheet converts to; nil while no sheet is open.
    private(set) var target: String?
    private(set) var step = Step.confirm
    /// The last preview answered for the sheet's code and, in D7, for the rate typed now.
    private(set) var preview: MoneyRecap_V1_CurrencyChangePreview?
    /// D7's rate as typed: digits and one separator, at most 11 digits before it and 12 after.
    private(set) var rate = ""
    /// The sheet's refusal line, `storeLimit` or `convertFailure`; the D5 list's is the Settings save line.
    private(set) var refusal: String?
    private(set) var isBusy = false
    /// D7's preview, asked `pause` after typing stops.
    private(set) var typing: Task<Void, Never>?
    private let askPreview: Preview
    private let change: Change
    private let pause: Duration
    private let locale: Locale
    private var attempt: SaveAttempt<[String]>?

    init(
        current: String,
        options: [CurrencyViewModel.Option] = CurrencyViewModel.options(),
        pause: Duration = .milliseconds(500),
        locale: Locale = .autoupdatingCurrent,
        preview: @escaping Preview,
        change: @escaping Change
    ) {
        self.current = current
        self.options = options
        self.pause = pause
        self.locale = locale
        askPreview = preview
        self.change = change
    }

    var visible: [CurrencyViewModel.Option] {
        CurrencyViewModel.matching(query, in: options)
    }

    var noMatch: String? {
        CurrencyViewModel.noMatch(query, in: options)
    }

    var sheetOpen: Bool {
        get { target != nil }
        set {
            if !newValue {
                cancel()
            }
        }
    }

    /// The current code does nothing; another asks for a preview without a rate.
    func choose(_ code: String) async {
        guard code != current, !isBusy else { return }
        refusal = nil
        isBusy = true
        defer { isBusy = false }
        switch await askPreview(.with { $0.toCurrencyCode = code }) {
        case let .success(answer) where answer.preview.nothingToConvert:
            await commit(to: code, rate: "", source: .unspecified)
        case let .success(answer):
            open(code, answer.preview.unsupportedPair ? .manual : .confirm, answer.preview)
        case let .failure(error) where error.code == .outOfRange:
            open(code, .confirm, nil)
            refusal = Self.storeLimit
        case .failure:
            open(code, .failed, nil)
        }
    }

    /// D6 failed's `Try again`: the fetch once more.
    func tryAgain() async {
        guard let target else { return }
        await choose(target)
    }

    /// D6 failed's `Enter it myself`: D7 with no rate typed.
    func enterItMyself() {
        step = .manual
        preview = nil
        refusal = nil
        rate = ""
    }

    func cancel() {
        typing?.cancel()
        target = nil
        preview = nil
        refusal = nil
        rate = ""
    }

    /// D7's field: the text clamped to the rate's limits, and the preview asked again a moment after typing stops.
    func type(_ text: String) {
        rate = Self.clamp(text)
        preview = nil
        refusal = nil
        typing?.cancel()
        guard let target, let sent = Self.sendable(rate) else { return }
        let asked = rate
        typing = Task { [weak self, pause] in
            try? await Task.sleep(for: pause)
            guard !Task.isCancelled else { return }
            await self?.ask(target, rate: sent, typed: asked)
        }
    }

    private func ask(_ code: String, rate sent: String, typed: String) async {
        let result = await askPreview(.with {
            $0.toCurrencyCode = code
            $0.manualRate = sent
        })
        guard rate == typed, target == code else { return }
        switch result {
        case let .success(answer):
            preview = answer.preview
        case let .failure(error):
            refusal = error.code == .outOfRange ? Self.storeLimit : nil
        }
    }

    /// Convert: enabled once a preview with its rate has answered and nothing refused it past D57.
    var canConvert: Bool {
        preview?.hasPreview == true && refusal != Self.storeLimit && !isBusy
    }

    /// Convert under one key per code and rate; the answer's Home and Settings replace the stored ones (the caller's
    /// `change`), and the sheet closes on the new Currency.
    func convert() async {
        guard canConvert, let target, let preview else { return }
        isBusy = true
        defer { isBusy = false }
        await commit(to: target, rate: preview.preview.rate, source: preview.source)
    }

    private func commit(to code: String, rate: String, source: MoneyRecap_V1_RateSource) async {
        var current = attempt ?? SaveAttempt([code, rate])
        current.update([code, rate])
        attempt = current
        let request = MoneyRecap_V1_BffServiceChangeCurrencyRequest.with {
            $0.toCurrencyCode = code
            $0.rate = rate
            $0.source = source
        }
        switch await change(request, current.headers) {
        case .success:
            let direct = target == nil
            cancel()
            self.current = code
            changedValue = direct ? Self.changed : nil
        case let .failure(error):
            refusal = error.code == .outOfRange ? Self.storeLimit
                : target == nil ? SettingsViewModel.saveFailure : Self.convertFailure
        }
    }

    private func open(_ code: String, _ step: Step, _ preview: MoneyRecap_V1_CurrencyChangePreview?) {
        target = code
        self.step = step
        self.preview = preview
        rate = ""
    }

    /// "Convert everything to GBP?"
    var title: String {
        "Convert everything to \(target ?? "")?"
    }

    /// D6's rate line: "All your amounts will be converted at today's rate, 1 IDR = 0.000046 GBP."
    var rateLine: String? {
        guard step == .confirm, let conversion = preview?.preview else { return nil }
        return "All your amounts will be converted at today's rate, 1 \(conversion.fromCurrencyCode) = "
            + "\(Self.number(conversion.rate, locale: locale)) \(conversion.toCurrencyCode)."
    }

    /// "Your Rp15.000.000 salary becomes £690.", with a Salary only.
    var salaryLine: String? {
        guard let conversion = preview?.preview, conversion.hasSalaryBefore else { return nil }
        return "Your \(money(conversion.salaryBefore)) salary becomes \(money(conversion.salaryAfter))."
    }

    /// The rows `Salary` (with a Salary only) and `Spends`, each "{before} → {after}".
    var rows: [(label: String, value: String)] {
        guard let conversion = preview?.preview else { return [] }
        let salary = conversion.hasSalaryBefore
            ? [("Salary", "\(money(conversion.salaryBefore)) → \(money(conversion.salaryAfter))")] : []
        return salary + [("Spends", "\(money(conversion.spendsBefore)) → \(money(conversion.spendsAfter))")]
    }

    private func money(_ amount: MoneyRecap_V1_Money) -> String {
        Money.format(amount.amount, currency: amount.currencyCode, locale: locale)
    }

    /// A decimal string in `locale`, every digit ledger sent kept.
    static func number(_ decimal: String, locale: Locale) -> String {
        let value = Decimal(string: decimal) ?? 0
        return value.formatted(.number.locale(locale).precision(.fractionLength(0 ... 12)))
    }

    /// Digits and the first `.` or `,`, at most 11 digits before it and 12 after; extra keys are dropped, never the
    /// whole edit.
    static func clamp(_ text: String) -> String {
        var whole = "", fraction = "", separator: Character?
        for key in text {
            if key == "." || key == "," {
                separator = separator ?? key
            } else if key.isASCII, key.isNumber {
                if separator == nil, whole.count < 11 {
                    whole.append(key)
                } else if separator != nil, fraction.count < 12 {
                    fraction.append(key)
                }
            }
        }
        return whole + (separator.map(String.init) ?? "") + fraction
    }

    /// The typed rate as the BFF reads it, "0.5" for ",5", or nil unless it is above zero.
    static func sendable(_ typed: String) -> String? {
        let parts = typed.split(separator: typed.contains(",") ? "," : ".", omittingEmptySubsequences: false)
        let whole = parts.first.map(String.init).flatMap { $0.isEmpty ? nil : $0 } ?? "0"
        let fraction = parts.count > 1 ? String(parts[1]) : ""
        let sent = fraction.isEmpty ? whole : "\(whole).\(fraction)"
        guard let value = Decimal(string: sent), value > 0 else { return nil }
        return sent
    }
}
