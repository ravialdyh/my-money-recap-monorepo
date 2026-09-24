import SwiftUI

/// Itemised row (`DataRow.tsx`): the Category's `IconWell`, a one-line label with an optional secondary line under it,
/// the Fredoka amount, a divider below. With an `action` the row is a button; without one it only shows.
struct DataRow: View {
    let icon: Icon.Name
    let label: String
    var detail: String?
    let amount: String
    var action: (() -> Void)?

    /// The canvases' secondary line: Nunito 12pt bold in secondary ink.
    private static let detailStyle = TextStyle(
        family: Tokens.FontFamily.text, weight: .bold, size: 12, lineHeight: 1.2, letterSpacing: 0, uppercase: false
    )

    var body: some View {
        if let action {
            Button(action: action) { row }.buttonStyle(PressableStyle())
        } else {
            row
        }
    }

    private var row: some View {
        HStack(spacing: Tokens.Space._4) {
            IconWell(tone: icon)
            VStack(alignment: .leading, spacing: 2) {
                Text(label).textStyle(Tokens.TypeRamp.labelM).lineLimit(1).truncationMode(.tail)
                if let detail {
                    Text(detail).textStyle(Self.detailStyle).foregroundStyle(Tokens.Text.secondary)
                }
            }
            Spacer(minLength: 0)
            Text(amount).textStyle(Tokens.TypeRamp.titleS).monospacedDigit().lineLimit(1).fixedSize()
        }
        .foregroundStyle(Tokens.Text.primary)
        .padding(.vertical, 7)
        .contentShape(Rectangle())
        .overlay(alignment: .bottom) { Tokens.Surface.divider.frame(height: 1) }
    }
}
