import SwiftUI

/// The Recap pager's 44pt close control (PRD D46): the system plus turned 45°, named "Close", in the ground's type
/// colour.
struct CloseControl: View {
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
        .foregroundStyle(Tokens.Text.onGround)
        .accessibilityLabel("Close")
    }
}
