import SwiftUI

/// Single-line input inside a Card (`TextField.tsx`): a 52pt pill on the field surface with a 2pt stroke,
/// violet when focused and coral when invalid, and a helper line that turns coral with it.
struct MMRTextField: View {
    let label: String
    @Binding var text: String
    var placeholder = ""
    var helper: String?
    var invalid = false
    var focused = false
    var keyboard: UIKeyboardType = .default
    var content: UITextContentType?
    var capitalization: TextInputAutocapitalization = .never
    var submit: SubmitLabel = .send
    var onSubmit: () -> Void = {}

    var body: some View {
        VStack(alignment: .leading, spacing: Tokens.Space._3) {
            Text(label).textStyle(Tokens.TypeRamp.labelS).foregroundStyle(Tokens.Text.secondary)
            TextField(placeholder, text: $text)
                .textStyle(Tokens.TypeRamp.bodyM)
                .foregroundStyle(Tokens.Text.primary)
                .textInputAutocapitalization(capitalization)
                .autocorrectionDisabled()
                .keyboardType(keyboard)
                .textContentType(content)
                .submitLabel(submit)
                .onSubmit(onSubmit)
                .accessibilityLabel(label)
                .padding(.horizontal, Tokens.Space._6)
                .frame(maxWidth: .infinity, minHeight: Tokens.Size.fieldHeight, alignment: .leading)
                .background(Tokens.Surface.field, in: Capsule())
                .overlay(Capsule().strokeBorder(ring, lineWidth: 2))
            if let helper {
                Text(helper)
                    .textStyle(Tokens.TypeRamp.bodyS)
                    .foregroundStyle(invalid ? Tokens.Text.accentCoral : Tokens.Text.secondary)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
    }

    private var ring: Color {
        if invalid {
            Tokens.State.invalidRing
        } else if focused {
            Tokens.State.focusRing
        } else {
            Tokens.Palette.ink100
        }
    }
}
