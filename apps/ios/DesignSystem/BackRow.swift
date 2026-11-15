import SwiftUI

struct BackRow: View {
    let label: String
    let back: () -> Void

    var body: some View {
        HStack(spacing: Tokens.Space._1) {
            Button(action: back) {
                Image(systemName: "chevron.left")
                    .font(.system(size: 20, weight: .bold))
                    .frame(width: Tokens.Size.touchTargetMin, height: Tokens.Size.touchTargetMin)
                    .contentShape(Rectangle())
            }
            .buttonStyle(PressableStyle())
            .accessibilityLabel("Back")
            Chip(text: label, asLabel: true)
        }
        .padding(.leading, -Tokens.Space._5)
    }
}
