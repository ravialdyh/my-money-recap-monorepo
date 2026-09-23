import SwiftUI

/// Empty-state card, one tap target (PRD §13, D50): eyebrow, title and helper beside a chevron well, at the filled
/// `Card`'s height so nothing moves when filled in (§6.8, D54, D69). `dashed` is Home's preview card (D47).
struct PromptCard: View {
    var eyebrow: String?
    let title: String
    var helper: String?
    var height: CGFloat?
    var dashed = false
    var action: () -> Void = {}

    var body: some View {
        Button(action: action) {
            HStack(spacing: 16) {
                VStack(alignment: .leading, spacing: Tokens.Space._2) {
                    if let eyebrow {
                        Eyebrow(text: eyebrow)
                    }
                    CardTitle(text: title)
                    if let helper {
                        Text(helper)
                            .textStyle(Tokens.TypeRamp.bodyS)
                            .foregroundStyle(dashed ? Tokens.Text.onGround : Tokens.Text.secondary)
                            .fixedSize(horizontal: false, vertical: true)
                    }
                }
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
        .accessibilityLabel(title)
        .accessibilityHint(helper ?? "")
    }
}
