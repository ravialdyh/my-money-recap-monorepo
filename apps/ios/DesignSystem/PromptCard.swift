import SwiftUI

struct PromptCard: View {
    var eyebrow: String?
    let title: String
    var helper: String?
    var height: CGFloat?
    var dashed = false
    var action: () -> Void = {}

    private var well: CGFloat {
        dashed ? 36 : Tokens.Size.touchTargetMin
    }

    var body: some View {
        Button(action: action) {
            HStack(spacing: 16) {
                VStack(alignment: .leading, spacing: Tokens.Space._2) {
                    if let eyebrow {
                        Eyebrow(text: eyebrow)
                    }
                    if dashed {
                        Text(title).textStyle(Tokens.TypeRamp.bodyS).foregroundStyle(Tokens.Text.onGroundMuted)
                    } else {
                        CardTitle(text: title)
                    }
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
                    .frame(width: well, height: well)
                    .overlay(Image(systemName: "chevron.right").font(.system(size: 16, weight: .heavy))
                        .foregroundStyle(dashed ? Tokens.Text.onGround : Tokens.Palette.violet500))
            }
            .padding(dashed ? EdgeInsets(top: 10, leading: Tokens.Space._7, bottom: 10, trailing: Tokens.Space._4)
                : EdgeInsets(
                    top: Tokens.Space._7,
                    leading: Tokens.Space._7,
                    bottom: Tokens.Space._7,
                    trailing: Tokens.Space._7
                ))
            .frame(maxWidth: .infinity, minHeight: dashed ? Tokens.Space._10 : height, alignment: .leading)
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
