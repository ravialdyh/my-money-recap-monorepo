import SwiftUI

/// Settings, Recap and Currency list row (`ListRow.tsx`): label over an optional detail, a right-aligned value, then
/// a chevron, a selected row's violet check, or nothing; a divider below. Destructive is coral; disabled, at 40%.
struct ListRow: View {
    enum Trailing {
        case chevron, check, none
    }

    let label: String
    var detail: String?
    var value: String?
    var trailing = Trailing.chevron
    var selected = false
    var destructive = false
    var action: () -> Void = {}

    var body: some View {
        Button(action: action) {
            HStack(spacing: 12) {
                VStack(alignment: .leading, spacing: 2) {
                    Text(label).textStyle(Tokens.TypeRamp.labelM)
                        .foregroundStyle(destructive ? Tokens.Text.accentCoral : Tokens.Text.primary)
                    if let detail {
                        Text(detail).textStyle(Tokens.TypeRamp.labelS).foregroundStyle(Tokens.Text.secondary)
                    }
                }
                .multilineTextAlignment(.leading)
                Spacer(minLength: 0)
                if let value {
                    Text(value).textStyle(Tokens.TypeRamp.labelM).foregroundStyle(Tokens.Text.secondary)
                }
                switch trailing {
                case .chevron:
                    Image(systemName: "chevron.right").font(.system(size: 14, weight: .heavy))
                        .foregroundStyle(Tokens.Text.secondary)
                case .check:
                    check.opacity(selected ? 1 : 0)
                case .none:
                    EmptyView()
                }
            }
            .padding(.vertical, Tokens.Space._5)
            .contentShape(Rectangle())
            .overlay(alignment: .bottom) { Tokens.Surface.divider.frame(height: 1) }
        }
        .buttonStyle(PressableStyle())
        .accessibilityAddTraits(trailing == .check && selected ? .isSelected : [])
    }

    /// An 18pt violet disc with a white tick (`.mmr-list-row-check`), its space kept when unselected so no row moves.
    private var check: some View {
        Circle()
            .fill(Tokens.Palette.violet500)
            .frame(width: 18, height: 18)
            .overlay(Image(systemName: "checkmark").font(.system(size: 9, weight: .black)).foregroundStyle(.white))
    }
}
