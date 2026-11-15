import SwiftUI

struct DayPicker: View {
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
