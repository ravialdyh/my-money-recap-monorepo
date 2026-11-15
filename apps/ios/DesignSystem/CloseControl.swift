import SwiftUI

struct CloseControl: View {
    var onLight = false
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Image(systemName: "plus")
                .font(.system(size: Tokens.Size.iconButton, weight: .bold))
                .rotationEffect(.degrees(45))
                .frame(width: Tokens.Size.touchTargetMin, height: Tokens.Size.touchTargetMin)
                .contentShape(Rectangle())
        }
        .buttonStyle(PressableStyle())
        .foregroundStyle(onLight ? Tokens.Text.onGroundDark : Tokens.Text.onGround)
        .accessibilityLabel("Close")
    }
}
