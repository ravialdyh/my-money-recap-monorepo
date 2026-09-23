import SwiftUI

/// Payday picker (`DayPicker.tsx`, PRD D51): 44pt pills, 1 to 31 then `Last day`, nothing disabled, the selection
/// the ink pill, centred on open, and the next three dates below, which are the caller's arithmetic (PRD §8.2).
struct DayPicker: View {
    /// The contract's Payday values in the order the picker draws them.
    static let values = (1 ... 31).map(String.init) + ["last"]

    let selected: String
    let nextPaydays: [String]
    let choose: (String) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: Tokens.Space._4) {
            ScrollViewReader { reader in
                ScrollView(.horizontal) {
                    HStack(spacing: Tokens.Space._3) {
                        ForEach(Self.values, id: \.self) { pill($0).id($0) }
                    }
                }
                .scrollIndicators(.hidden)
                .onAppear { reader.scrollTo(selected, anchor: .center) }
            }
            Text("Next paydays: " + nextPaydays.joined(separator: " · "))
                .textStyle(Tokens.TypeRamp.bodyS)
                .monospacedDigit()
                .foregroundStyle(Tokens.Text.secondary)
        }
    }

    private func pill(_ value: String) -> some View {
        let chosen = value == selected
        let last = value == "last"
        return Button { choose(value) } label: {
            Text(last ? "Last day" : value)
                .textStyle(Tokens.TypeRamp.labelM)
                .monospacedDigit()
                .padding(.horizontal, last ? Tokens.Space._6 : Tokens.Space._4)
                .frame(minWidth: Tokens.Size.touchTargetMin, minHeight: Tokens.Size.touchTargetMin)
                .background(chosen ? Tokens.Surface.buttonPrimary : Tokens.Surface.field, in: Capsule())
                .foregroundStyle(chosen ? Tokens.Text.onInverse : Tokens.Text.primary)
        }
        .buttonStyle(PressableStyle())
        .accessibilityAddTraits(chosen ? .isSelected : [])
    }
}
