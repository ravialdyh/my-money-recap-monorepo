import SwiftUI

/// Full-width card, no shadow (`Banner.tsx`): coral for the Target Warning (PRD D38, D53), violet `info` for a notice,
/// grey for no connection (D58). An × for the warning's session-only dismissal (§6.8); a link, underlined, after the
/// text.
struct Banner: View {
    enum Tone {
        case warning, info, offline
    }

    var tone = Tone.warning
    let text: String
    var onDismiss: (() -> Void)?
    var link: (title: String, action: () -> Void)?

    var body: some View {
        HStack(alignment: .top, spacing: Tokens.Space._4) {
            VStack(alignment: .leading, spacing: Tokens.Space._2) {
                Text(text)
                    .textStyle(Tokens.TypeRamp.bodyS)
                    .fixedSize(horizontal: false, vertical: true)
                if let link {
                    Button(action: link.action) {
                        Text(link.title).textStyle(Tokens.TypeRamp.bodyS).fontWeight(.heavy).underline()
                            .frame(minHeight: Tokens.Size.touchTargetMin)
                            .contentShape(Rectangle())
                    }
                    .buttonStyle(PressableStyle())
                    .padding(.vertical, -11)
                }
            }
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
            fill,
            in: RoundedRectangle(cornerRadius: Tokens.Radius.card, style: .continuous)
        )
        .foregroundStyle(Tokens.Text.onInverse)
        .accessibilityElement(children: .contain)
    }

    private var fill: Color {
        switch tone {
        case .warning: Tokens.Surface.banner
        case .info: Tokens.Surface.cardAccent
        case .offline: Tokens.Surface.bannerOffline
        }
    }
}
