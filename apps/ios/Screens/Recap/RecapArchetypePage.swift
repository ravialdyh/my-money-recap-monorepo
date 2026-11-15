import SwiftUI

struct RecapArchetypePage: View {
    @Bindable var model: RecapViewModel
    var active = true

    var body: some View {
        Screen(ground: Tokens.Ground.personality, scrolls: true) {
            StatChipRow {
                Chip(text: "Your money personality", asLabel: true)
            }
            .padding(.leading, RecapLayout.chipShift)
            .padding(.top, Tokens.Space._3)
            if model.partial {
                closing
            } else if let archetype = model.archetype {
                personality(archetype)
            } else {
                PromptCard(
                    title: "Set a saving target to get an archetype",
                    helper: "Your archetype comes from what you kept against what you meant to keep.",
                    height: Tokens.Size.homeHeroCard,
                    action: { Task { await model.editTarget() } }
                )
                .padding(.top, Tokens.Space._6)
                Spacer(minLength: Tokens.Space._6)
                Card { bars }
                pills(nil)
            }
        }
    }

    @ViewBuilder private func personality(_ archetype: ArchetypePill.Kind) -> some View {
        let recap = model.answer?.recap
        Text(model.personalityLead)
            .textStyle(RecapLayout.lead)
            .foregroundStyle(Tokens.Text.onGroundMuted)
            .padding(.top, Tokens.Space._6)
        Headline(text: recap?.archetypeTitle ?? "", size: .display)
            .padding(.top, 2)
        SubCopy(text: recap?.archetypeDescription ?? "")
            .padding(.top, Tokens.Space._4)
        Illustration(name: Self.mascot(archetype), width: RecapLayout.mascot, maxHeight: RecapLayout.mascotCap)
        Card(underIllustration: true) {
            bars
            MMRButton(title: model.share, icon: "square.and.arrow.up") { model.sharePoster() }
        }
        pills(archetype)
            .sheet(item: $model.shared) { shared in
                ShareSheet(image: shared.image) { model.shared = nil }
            }
    }

    @ViewBuilder private var closing: some View {
        Headline(text: model.answer?.recap.closingLine ?? "")
            .padding(.top, Tokens.Space._6)
        Spacer(minLength: Tokens.Space._6)
        VStack(alignment: .leading, spacing: 10) {
            Text("Archetypes")
                .textStyle(Tokens.TypeRamp.captionUpper)
                .foregroundStyle(Tokens.Text.onGroundMuted)
            ArchetypeRow().opacity(Tokens.State.disabledOpacity)
        }
        .padding(.bottom, Tokens.Space._3)
        PageDots(count: model.pages.count, active: $model.page)
            .padding(.top, 18)
    }

    private var bars: some View {
        CompareBars(rows: model.savedAgainstSpent, shown: active)
    }

    private func pills(_ archetype: ArchetypePill.Kind?) -> some View {
        ArchetypeRow(selected: archetype)
            .opacity(archetype == nil ? Tokens.State.disabledOpacity : 1)
            .padding(.top, Tokens.Space._4)
            .padding(.leading, RecapLayout.pillsInset)
    }

    static func mascot(_ archetype: ArchetypePill.Kind) -> String {
        switch archetype {
        case .squirrel: "ill-squirrel"
        case .hustler: "ill-hustler"
        case .treatYourself: "ill-treat"
        case .slowBurner: "ill-slowburner"
        }
    }
}

extension RecapLayout {
    static let lead = TextStyle(
        family: Tokens.FontFamily.text,
        weight: .heavy,
        size: 15,
        lineHeight: 1.2,
        letterSpacing: 0,
        uppercase: false
    )
    static let mascot: CGFloat = 340
    static let mascotCap: CGFloat = 300
    static let pillsInset: CGFloat = 18
}

extension RecapViewModel {
    var archetype: ArchetypePill.Kind? {
        switch answer?.recap.archetype {
        case .squirrel: .squirrel
        case .hustler: .hustler
        case .treatYourself: .treatYourself
        case .slowBurner: .slowBurner
        default: nil
        }
    }

    var personalityLead: String {
        "This month, \(answer?.displayName ?? ""), you were"
    }

    var share: String {
        "Share your \(answer?.recap.archetypeName ?? "")"
    }

    var poster: SharePoster? {
        guard let archetype, let answer else { return nil }
        return SharePoster(
            archetype: .init(kind: archetype, title: answer.recap.archetypeTitle),
            month: monthChip,
            displayName: answer.displayName,
            saved: money(answer.recap.saved),
            spent: money(answer.recap.spent)
        )
    }

    func sharePoster() {
        shared = poster.map { SharedPicture(image: $0.picture()) }
    }

    var savedAgainstSpent: [CompareBars.Row] {
        guard let recap = answer?.recap else { return [] }
        return [
            .init(label: "Saved", value: Double(recap.saved.amount) ?? 0, display: money(recap.saved), tone: .green),
            .init(label: "Spent", value: Double(recap.spent.amount) ?? 0, display: money(recap.spent), tone: .coral),
        ]
    }
}
