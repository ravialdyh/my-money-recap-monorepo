import SwiftUI

/// Empty-state card: the whole card is one tap target (PRD §13, D50). It holds the filled `Card`'s
/// height so nothing moves when the Member fills it in (PRD §6.8, D54, D69). `dashed` is the Home
/// preview card (D47), drawn as an outline on the ground.
struct PromptCard: View {
    let title: String
    var height: CGFloat?
    var dashed = false
    var action: () -> Void = {}

    var body: some View {
        Button(action: action) {
            HStack(spacing: 16) {
                CardTitle(text: title)
                    .multilineTextAlignment(.leading)
                Spacer(minLength: 0)
                Circle()
                    .fill(dashed ? Tokens.Surface.chipOnGround.opacity(0.2) : Tokens.Surface.wellViolet)
                    .frame(width: Tokens.Size.touchTargetMin, height: Tokens.Size.touchTargetMin)
                    .overlay(Image(systemName: "chevron.right").font(.system(size: 16, weight: .heavy))
                        .foregroundStyle(dashed ? Tokens.Text.onGround : Tokens.Palette.violet500))
            }
            .padding(Tokens.Space._7)
            .frame(maxWidth: .infinity, minHeight: height, alignment: .leading)
            .background(
                dashed ? Color.clear : Tokens.Surface.card,
                in: RoundedRectangle(cornerRadius: Tokens.Radius.card, style: .continuous)
            )
            .overlay(RoundedRectangle(cornerRadius: Tokens.Radius.card, style: .continuous)
                .strokeBorder(Tokens.Text.onGroundMuted, style: StrokeStyle(lineWidth: dashed ? 2 : 0, dash: [8, 6])))
            .foregroundStyle(dashed ? Tokens.Text.onGround : Tokens.Text.primary)
            .contentShape(RoundedRectangle(cornerRadius: Tokens.Radius.card, style: .continuous))
        }
        .buttonStyle(PressableStyle())
    }
}
