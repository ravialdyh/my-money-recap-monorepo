import SwiftUI

/// Full-width card, `radius.card`, no shadow (`Banner.tsx`): coral for the Target Warning (PRD D38, D53), grey for
/// no connection (D58). A dismiss action adds the × the Target Warning's session-only dismissal needs (§6.8).
struct Banner: View {
    enum Tone {
        case warning, offline
    }

    var tone = Tone.warning
    let text: String
    var onDismiss: (() -> Void)?

    var body: some View {
        HStack(alignment: .top, spacing: Tokens.Space._4) {
            Text(text)
                .textStyle(Tokens.TypeRamp.bodyS)
                .fixedSize(horizontal: false, vertical: true)
                .frame(maxWidth: .infinity, alignment: .leading)
            if let onDismiss {
                Button(action: onDismiss) {
                    Image(systemName: "xmark")
                        .font(.system(size: 14, weight: .heavy))
                        .frame(width: Tokens.Size.touchTargetMin, height: Tokens.Size.touchTargetMin)
                        .contentShape(Rectangle())
                }
                .buttonStyle(PressableStyle())
                .padding(.vertical, -11)
                .padding(.trailing, -14)
                .accessibilityLabel("Dismiss")
            }
        }
        .padding(.vertical, Tokens.Space._6)
        .padding(.horizontal, Tokens.Space._7)
        .background(
            tone == .warning ? Tokens.Surface.banner : Tokens.Surface.bannerOffline,
            in: RoundedRectangle(cornerRadius: Tokens.Radius.card, style: .continuous)
        )
        .foregroundStyle(Tokens.Text.onInverse)
        .accessibilityElement(children: .contain)
    }
}
