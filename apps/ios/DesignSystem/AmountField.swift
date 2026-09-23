import SwiftUI
import UIKit

/// Amount entry in a Card (`AmountField.tsx`): the symbol fixed, the amount grouped as the device writes it at 48,
/// 36 or 28pt, measured (PRD §11). `amount` is an exact decimal: 11 whole digits (D57), `digits` decimals (#92).
struct AmountField: View {
    static let cap = "That's more than the app can store."

    let label: String
    let currency: String
    let digits: Int
    @Binding var amount: String
    var locale: Locale = .autoupdatingCurrent
    @State private var width: CGFloat = 0
    @State private var focused = false

    var body: some View {
        let full = AmountText.isFull(amount)
        let size = fitted
        VStack(alignment: .leading, spacing: Tokens.Space._2) {
            Text(label).textStyle(Tokens.TypeRamp.labelM).foregroundStyle(Tokens.Text.secondary)
            HStack(spacing: Tokens.Space._3) {
                Text(symbol)
                    .font(Font.custom(Tokens.FontFamily.display, size: size).weight(.bold))
                    .foregroundStyle(Tokens.Text.primary)
                    .accessibilityHidden(true)
                AmountInput(
                    amount: $amount,
                    focused: $focused,
                    label: label,
                    digits: digits,
                    locale: locale,
                    size: size
                )
            }
            .onGeometryChange(for: CGFloat.self) { $0.size.width } action: { width = $0 }
            .padding(.vertical, Tokens.Space._4)
            .padding(.horizontal, Tokens.Space._6)
            .background(focused ? Tokens.Surface.card : Tokens.Surface.field, in: Capsule())
            .overlay(Capsule().strokeBorder(full ? Tokens.State.invalidRing : Tokens.State.focusRing, lineWidth: 2)
                .opacity(full || focused ? 1 : 0))
            if full {
                Text(Self.cap)
                    .textStyle(Tokens.TypeRamp.labelS)
                    .foregroundStyle(Tokens.Text.accentCoral)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
    }

    /// The symbol the platform formatter prints for the Currency in this locale: "Rp", "£", or the code itself.
    private var symbol: String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.locale = locale
        formatter.currencyCode = currency
        return formatter.currencySymbol
    }

    private var fitted: CGFloat {
        Self.size(for: symbol + " " + AmountText.shown(amount, locale: locale), in: width)
    }

    /// The largest amount step at which `line`, the symbol, a space and the grouped amount, drawn in Fredoka bold,
    /// fits `width`; 48pt before the row is measured, and 28pt when nothing fits.
    static func size(for line: String, in width: CGFloat) -> CGFloat {
        let steps = [Tokens.Size.amountL, Tokens.Size.amountM, Tokens.Size.amountS]
        return steps
            .first { width == 0 || (line as NSString).size(withAttributes: [.font: font(of: $0)]).width <= width }
            ?? Tokens.Size.amountS
    }

    /// Fredoka at the bold weight the symbol is drawn in, which the typed amount is drawn and measured in.
    static func font(of size: CGFloat) -> UIFont {
        let family = UIFontDescriptor(fontAttributes: [.family: Tokens.FontFamily.display])
        let bold = family.addingAttributes([.traits: [UIFontDescriptor.TraitKey.weight: UIFont.Weight.bold]])
        return UIFont(descriptor: bold, size: size)
    }
}

/// The UIKit field under `AmountField`: each keystroke is read, kept to the cap and regrouped inside
/// `shouldChangeCharactersIn`, before the next one arrives, so no keystroke is lost to a redraw.
private struct AmountInput: UIViewRepresentable {
    @Binding var amount: String
    @Binding var focused: Bool
    let label: String
    let digits: Int
    let locale: Locale
    let size: CGFloat

    func makeUIView(context: Context) -> UITextField {
        let field = UITextField()
        field.delegate = context.coordinator
        field.keyboardType = digits > 0 ? .decimalPad : .numberPad
        field.accessibilityLabel = label
        field.textColor = UIColor(Tokens.Text.primary)
        field.tintColor = UIColor(Tokens.State.focusRing)
        field.setContentHuggingPriority(.defaultLow, for: .horizontal)
        field.setContentCompressionResistancePriority(.defaultLow, for: .horizontal)
        return field
    }

    func updateUIView(_ field: UITextField, context: Context) {
        context.coordinator.input = self
        field.font = AmountField.font(of: size)
        let shown = AmountText.shown(amount, locale: locale)
        if field.text != shown {
            field.text = shown
        }
    }

    func makeCoordinator() -> Coordinator {
        Coordinator(self)
    }

    final class Coordinator: NSObject, UITextFieldDelegate {
        var input: AmountInput

        init(_ input: AmountInput) {
            self.input = input
        }

        func textField(
            _ field: UITextField,
            shouldChangeCharactersIn range: NSRange,
            replacementString string: String
        ) -> Bool {
            let edited = ((field.text ?? "") as NSString).replacingCharacters(in: range, with: string)
            let amount = AmountText.parsed(edited, digits: input.digits, locale: input.locale)
            field.text = AmountText.shown(amount, locale: input.locale)
            input.amount = amount
            return false
        }

        func textFieldDidBeginEditing(_: UITextField) {
            input.focused = true
        }

        func textFieldDidEndEditing(_: UITextField) {
            input.focused = false
        }
    }
}

/// What the amount field reads and writes: the exact decimal the contract carries, and that decimal as the device
/// locale groups it. Pure, so the rules are tested without a screen.
enum AmountText {
    /// D57: whole digits the app can store and every screen can show.
    static let wholeDigits = 11

    /// True at 11 whole digits, where the field takes no more and says so (D57).
    static func isFull(_ amount: String) -> Bool {
        amount.split(separator: ".", omittingEmptySubsequences: false)[0].count >= wholeDigits
    }

    /// The exact decimal in what was typed, grouping dropped and the locale's separator read as the point, kept to 11
    /// whole digits and the Currency's decimals, so a keystroke past either is dropped, as is a point with none.
    static func parsed(_ typed: String, digits: Int, locale: Locale) -> String {
        let point = locale.decimalSeparator ?? "."
        var whole = ""
        var fraction: String?
        for character in typed {
            if let digit = character.wholeNumberValue, digit < 10 {
                if fraction == nil {
                    whole += String(digit)
                } else {
                    fraction? += String(digit)
                }
            } else if String(character) == point, fraction == nil, digits > 0 {
                fraction = ""
            }
        }
        let trimmed = String(whole.drop { $0 == "0" })
        whole = trimmed.isEmpty && (!whole.isEmpty || fraction != nil) ? "0" : trimmed
        return String(whole.prefix(wholeDigits)) + (fraction.map { "." + $0.prefix(digits) } ?? "")
    }

    /// `amount` grouped as the locale writes it, keeping a trailing point and the decimals exactly as typed.
    static func shown(_ amount: String, locale: Locale) -> String {
        guard !amount.isEmpty else { return "" }
        let parts = amount.split(separator: ".", omittingEmptySubsequences: false)
        let whole = (Int64(parts[0]) ?? 0).formatted(.number.locale(locale))
        guard parts.count > 1 else { return whole }
        return whole + (locale.decimalSeparator ?? ".") + parts[1]
    }
}
