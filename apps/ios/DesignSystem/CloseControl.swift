import SwiftUI

/// The Recap pager's 44pt close control (PRD D46): the system plus turned 45°, named "Close", in the ground's type
/// colour: white, or ink `onLight` (the yellow ground).
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
