import SwiftUI

/// `Toggle.tsx`: the iOS-proportioned switch, 51 by 31 with a 27pt white knob, violet when on and ink-100 when off.
/// Drawn, since the system switch changes size between iOS versions; VoiceOver hears a native switch. Disabled, 40%.
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
