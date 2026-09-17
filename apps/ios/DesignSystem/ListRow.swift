import SwiftUI

/// Settings and Recap list row: label, a right-aligned secondary value and a chevron; a divider below.
struct ListRow: View {
    let label: String
    var value: String?
    var action: () -> Void = {}

    var body: some View {
        Button(action: action) {
            HStack(spacing: 12) {
                Text(label).textStyle(Tokens.TypeRamp.labelM).foregroundStyle(Tokens.Text.primary)
                Spacer(minLength: 0)
                if let value {
                    Text(value).textStyle(Tokens.TypeRamp.labelM).foregroundStyle(Tokens.Text.secondary)
                }
                Image(systemName: "chevron.right").font(.system(size: 14, weight: .heavy))
                    .foregroundStyle(Tokens.Text.secondary)
            }
            .padding(.vertical, Tokens.Space._5)
            .contentShape(Rectangle())
        }
        .buttonStyle(PressableStyle())
        .overlay(alignment: .bottom) { Tokens.Surface.divider.frame(height: 1) }
    }
}
