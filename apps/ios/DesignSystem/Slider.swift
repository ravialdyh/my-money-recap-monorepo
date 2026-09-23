import SwiftUI

/// The Saving Target slider (`Slider.tsx`, PRD §6.7): the violet `ProgressBar` with a 28pt ink knob ringed in white,
/// dragged anywhere on the full 44pt row, snapping to `step`. VoiceOver and UI tests meet it as the system slider.
struct Slider: View {
    let label: String
    @Binding var value: Int
    var range: ClosedRange<Int> = 0 ... 100
    var step = 5

    var body: some View {
        GeometryReader { proxy in
            let width = proxy.size.width
            let knob = Tokens.Size.sliderKnob
            ZStack(alignment: .leading) {
                ProgressBar(fraction: fraction, tone: .violet)
                Circle()
                    .fill(Tokens.Surface.buttonPrimary)
                    .overlay(Circle().strokeBorder(Tokens.Surface.card, lineWidth: 3))
                    .frame(width: knob, height: knob)
                    .offset(x: (width - knob) * fraction)
            }
            .frame(maxHeight: .infinity)
            .contentShape(Rectangle())
            .gesture(DragGesture(minimumDistance: 0).onChanged { drag in
                value = snapped(drag.location.x / max(width, 1))
            })
        }
        .frame(height: Tokens.Size.touchTargetMin)
        .accessibilityRepresentation {
            SwiftUI.Slider(
                value: continuous,
                in: Double(range.lowerBound) ... Double(range.upperBound),
                step: Double(step)
            ) {
                Text(label)
            }
        }
    }

    private var fraction: Double {
        Double(value - range.lowerBound) / Double(range.upperBound - range.lowerBound)
    }

    private var continuous: Binding<Double> {
        Binding(
            get: { Double(value) },
            set: { value = snapped(($0 - Double(range.lowerBound)) / Double(range.upperBound - range.lowerBound)) }
        )
    }

    /// The value nearest `fraction` of the range, on a step.
    func snapped(_ fraction: Double) -> Int {
        let raw = Double(range.lowerBound) + min(max(fraction, 0), 1) * Double(range.upperBound - range.lowerBound)
        return Int((raw / Double(step)).rounded()) * step
    }
}
