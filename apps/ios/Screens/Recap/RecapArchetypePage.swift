import SwiftUI

/// Page 4, C4 or P4 (PRD §6.14, §6.17), green: C4's Archetype, mascot, Saved against Spent and Share, or with none the
/// target card and no Share; P4's closing line. The pills end each, at 40% with no Archetype. Scrolls held sideways.
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
                    height: Tokens.Size.homeHeroCard
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
            MMRButton(title: model.share, icon: "square.and.arrow.up") {}
        }
        pills(archetype)
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

    /// C4's pills under the card, the month's picked out, or all at 40% without an Archetype.
    private func pills(_ archetype: ArchetypePill.Kind?) -> some View {
        ArchetypeRow(selected: archetype)
            .opacity(archetype == nil ? Tokens.State.disabledOpacity : 1)
            .padding(.top, Tokens.Space._4)
            .padding(.leading, RecapLayout.pillsInset)
    }

    /// The Archetype's mascot in the asset catalog (`ARCHETYPE_ILLUSTRATION` in RecapPersonalityScreen.tsx).
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
    /// "This month, {Display Name}, you were": 15pt heavy Nunito (`RecapPersonalityScreen`).
    static let lead = TextStyle(
        family: Tokens.FontFamily.text,
        weight: .heavy,
        size: 15,
        lineHeight: 1.2,
        letterSpacing: 0,
        uppercase: false
    )
    /// The mascot's slot width and the height it is capped at, so the tallest, the Hustler's, clears the pills.
    static let mascot: CGFloat = 340
    static let mascotCap: CGFloat = 300
    /// The pill row's inset under C4's card (`marginLeft: 18`).
    static let pillsInset: CGFloat = 18
}

/// Page 4's words (spec #151 copy table): the title, description, name and closing line are ledger's (A2).
extension RecapViewModel {
    /// The month's Archetype, nil on a Partial Recap and without a Saving Target.
    var archetype: ArchetypePill.Kind? {
        switch answer?.recap.archetype {
        case .squirrel: .squirrel
        case .hustler: .hustler
        case .treatYourself: .treatYourself
        case .slowBurner: .slowBurner
        default: nil
        }
    }

    /// "This month, {Display Name}, you were", the current name.
    var personalityLead: String {
        "This month, \(answer?.displayName ?? ""), you were"
    }

    /// "Share your {Archetype name}", the short name ledger sends ("Share your Squirrel"); the action is #157's.
    var share: String {
        "Share your \(answer?.recap.archetypeName ?? "")"
    }

    /// Saved against Spent as shares of their sum, a negative Saved printed as it is and drawn empty.
    var savedAgainstSpent: [CompareBars.Row] {
        guard let recap = answer?.recap else { return [] }
        return [
            .init(label: "Saved", value: Double(recap.saved.amount) ?? 0, display: money(recap.saved), tone: .green),
            .init(label: "Spent", value: Double(recap.spent.amount) ?? 0, display: money(recap.spent), tone: .coral),
        ]
    }
}
