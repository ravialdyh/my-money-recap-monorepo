import SwiftUI

struct MMRToggle: View {
    let label: String
    @Binding var isOn: Bool

    var body: some View {
        Button { isOn.toggle() } label: {
            Capsule()
                .fill(isOn ? Tokens.Surface.toggleOn : Tokens.Surface.toggleOff)
                .frame(width: Tokens.Size.toggleWidth, height: Tokens.Size.toggleHeight)
                .overlay(alignment: isOn ? .trailing : .leading) {
                    Circle()
                        .fill(Tokens.Palette.white)
                        .frame(width: Tokens.Size.toggleKnob, height: Tokens.Size.toggleKnob)
                        .padding(2)
                }
                .animation(.easeOut(duration: Tokens.Motion.Duration.press), value: isOn)
                .frame(minHeight: Tokens.Size.touchTargetMin)
                .contentShape(Rectangle())
        }
        .buttonStyle(PressableStyle())
        .accessibilityRepresentation { Toggle(label, isOn: $isOn) }
    }
}
