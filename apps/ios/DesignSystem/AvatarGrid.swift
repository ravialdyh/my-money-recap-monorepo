import SwiftUI

struct AvatarGrid: View {
    static let keys = ["fox", "otter", "panda", "frog", "owl", "rabbit"]
    let selected: String?
    let onSelect: (String) -> Void
    private let cell = Tokens.Size.avatarGridCell
    private let hitArea: CGFloat = 88

    var body: some View {
        LazyVGrid(columns: Array(repeating: GridItem(.flexible()), count: 3), spacing: Tokens.Space._6) {
            ForEach(Self.keys, id: \.self) { key in
                Button { onSelect(key) } label: { avatar(key) }
                    .buttonStyle(PressableStyle())
                    .accessibilityLabel(key.capitalized)
                    .accessibilityAddTraits(key == selected ? .isSelected : [])
            }
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel("Pick your avatar")
    }

    private func avatar(_ key: String) -> some View {
        VStack(spacing: Tokens.Space._3) {
            ZStack {
                if key == selected {
                    Circle().fill(Tokens.Surface.wellViolet).frame(width: cell + 12, height: cell + 12)
                    Circle().fill(Tokens.State.selectedRing).frame(width: cell + 4, height: cell + 4)
                }
                Image("ill-avatar-\(key)").resizable().scaledToFill()
                    .frame(width: cell, height: cell)
                    .clipShape(Circle())
            }
            .frame(width: cell + 12, height: cell + 12)
            Text(key.capitalized).textStyle(Tokens.TypeRamp.labelS).foregroundStyle(Tokens.Text.secondary)
        }
        .frame(minWidth: hitArea, minHeight: hitArea)
        .contentShape(Rectangle())
    }
}
