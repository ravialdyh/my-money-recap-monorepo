import SwiftUI

/// Chips in a row that wraps (`StatChipRow`, `.mmr-stat-chip-row`): each at its own width, 8pt apart, a new line when
/// the next would not fit, as Home's stat chips do in rupiah (X3) and the Log a spend Categories do (PRD §6.9).
struct StatChipRow: Layout {
    var spacing = Tokens.Space._3

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache _: inout ()) -> CGSize {
        let frames = place(subviews, in: proposal.width ?? .infinity)
        return CGSize(width: proposal.width ?? frames.map(\.maxX).max() ?? 0, height: frames.map(\.maxY).max() ?? 0)
    }

    func placeSubviews(in bounds: CGRect, proposal _: ProposedViewSize, subviews: Subviews, cache _: inout ()) {
        for (subview, frame) in zip(subviews, place(subviews, in: bounds.width)) {
            subview.place(
                at: CGPoint(x: bounds.minX + frame.minX, y: bounds.minY + frame.minY),
                proposal: ProposedViewSize(frame.size)
            )
        }
    }

    private func place(_ subviews: Subviews, in width: CGFloat) -> [CGRect] {
        var origin = CGPoint.zero
        var line: CGFloat = 0
        return subviews.map { subview in
            let size = subview.sizeThatFits(ProposedViewSize(width: width, height: nil))
            if origin.x > 0, origin.x + size.width > width {
                origin = CGPoint(x: 0, y: origin.y + line + spacing)
                line = 0
            }
            defer {
                origin.x += size.width + spacing
                line = max(line, size.height)
            }
            return CGRect(origin: origin, size: size)
        }
    }
}
