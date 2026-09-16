import SwiftUI

/// Nunito body line under a headline, muted on the ground.
struct SubCopy: View {
    let text: String

    var body: some View {
        Text(text)
            .textStyle(Tokens.TypeRamp.bodyM)
            .foregroundStyle(Tokens.Text.onGroundMuted)
            .fixedSize(horizontal: false, vertical: true)
    }
}
