import SwiftUI

struct TextLink: View {
    let title: String
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Text(title)
                .textStyle(Tokens.TypeRamp.labelM)
                .underline()
                .foregroundStyle(Tokens.Palette.violet500)
                .padding(.horizontal, 4)
                .frame(minHeight: Tokens.Size.touchTargetMin)
                .contentShape(Rectangle())
        }
        .buttonStyle(PressableStyle())
    }
}

struct GroundLink: View {
    let title: String
    let action: () -> Void

    var body: some View {
        Button(title, action: action)
            .textStyle(Tokens.TypeRamp.bodyM)
            .fontWeight(.heavy)
            .foregroundStyle(Tokens.Text.onGround)
            .frame(maxWidth: .infinity, minHeight: Tokens.Size.touchTargetMin)
            .buttonStyle(PressableStyle())
    }
}
