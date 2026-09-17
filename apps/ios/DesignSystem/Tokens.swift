// Generated from my-money-recap-ds/src/tokens/tokens.json by scripts/build-swift-tokens.mjs. Do not edit by hand.
import SwiftUI

public struct TextStyle {
    public let family: String
    public let weight: Font.Weight
    public let size: CGFloat
    /// Multiple of size.
    public let lineHeight: CGFloat
    /// In em.
    public let letterSpacing: CGFloat
    public let uppercase: Bool
}

public enum Tokens {
    public enum Palette {
        public static let violet100 = Color(.sRGB, red: 0xEF / 255, green: 0xE9 / 255, blue: 0xFF / 255, opacity: 1)
        public static let violet500 = Color(.sRGB, red: 0x5B / 255, green: 0x4F / 255, blue: 0xE8 / 255, opacity: 1)
        public static let yellow100 = Color(.sRGB, red: 0xFF / 255, green: 0xF3 / 255, blue: 0xCC / 255, opacity: 1)
        public static let yellow500 = Color(.sRGB, red: 0xF7 / 255, green: 0xC9 / 255, blue: 0x48 / 255, opacity: 1)
        public static let coral100 = Color(.sRGB, red: 0xFF / 255, green: 0xE3 / 255, blue: 0xE0 / 255, opacity: 1)
        public static let coral500 = Color(.sRGB, red: 0xF0 / 255, green: 0x66 / 255, blue: 0x5A / 255, opacity: 1)
        public static let green100 = Color(.sRGB, red: 0xC9 / 255, green: 0xF0 / 255, blue: 0xD6 / 255, opacity: 1)
        public static let green500 = Color(.sRGB, red: 0x3D / 255, green: 0xAA / 255, blue: 0x6A / 255, opacity: 1)
        public static let gold100 = Color(.sRGB, red: 0xFD / 255, green: 0xEF / 255, blue: 0xD0 / 255, opacity: 1)
        public static let gold500 = Color(.sRGB, red: 0xE8 / 255, green: 0xA3 / 255, blue: 0x17 / 255, opacity: 1)
        public static let rust100 = Color(.sRGB, red: 0xF7 / 255, green: 0xE0 / 255, blue: 0xD3 / 255, opacity: 1)
        public static let rust500 = Color(.sRGB, red: 0xC9 / 255, green: 0x6A / 255, blue: 0x3B / 255, opacity: 1)
        public static let sky100 = Color(.sRGB, red: 0xDF / 255, green: 0xF3 / 255, blue: 0xFF / 255, opacity: 1)
        public static let sky500 = Color(.sRGB, red: 0x3B / 255, green: 0xA6 / 255, blue: 0xF5 / 255, opacity: 1)
        public static let ink100 = Color(.sRGB, red: 0xEF / 255, green: 0xEC / 255, blue: 0xF7 / 255, opacity: 1)
        public static let ink500 = Color(.sRGB, red: 0x6B / 255, green: 0x68 / 255, blue: 0x90 / 255, opacity: 1)
        public static let ink900 = Color(.sRGB, red: 0x1E / 255, green: 0x1B / 255, blue: 0x3A / 255, opacity: 1)
        public static let white = Color(.sRGB, red: 0xFF / 255, green: 0xFF / 255, blue: 0xFF / 255, opacity: 1)
    }

    public enum Ground {
        public static let recap = Palette.violet500
        public static let earning = Palette.yellow500
        public static let goal = Palette.coral500
        public static let personality = Palette.green500
    }

    public enum Surface {
        public static let card = Palette.white
        public static let cardAccent = Palette.violet500
        public static let chipOnGround = Palette.white
        public static let chipInverse = Palette.ink900
        public static let chipHighlight = Palette.yellow500
        public static let buttonPrimary = Palette.ink900
        public static let buttonSecondary = Palette.white
        public static let wellViolet = Palette.violet100
        public static let wellSky = Palette.sky100
        public static let wellCoral = Palette.coral100
        public static let wellYellow = Palette.yellow100
        public static let wellGreen = Palette.green100
        public static let wellGold = Palette.gold100
        public static let wellRust = Palette.rust100
        public static let wellInk = Palette.ink100
        public static let trackCoral = Palette.coral100
        public static let trackViolet = Palette.violet100
        public static let trackYellow = Palette.yellow100
        public static let trackGreen = Palette.green100
        public static let trackSky = Palette.sky100
        public static let trackGold = Palette.gold100
        public static let trackRust = Palette.rust100
        public static let trackInk = Palette.ink100
        public static let sheet = Palette.white
        public static let scrim = Color(.sRGB, red: 30 / 255, green: 27 / 255, blue: 58 / 255, opacity: 0.4)
        public static let field = Palette.ink100
        public static let banner = Palette.coral500
        public static let bannerOffline = Palette.ink500
        public static let toggleOn = Palette.violet500
        public static let toggleOff = Palette.ink100
        public static let divider = Palette.ink100
    }

    public enum Text {
        public static let primary = Palette.ink900
        public static let secondary = Palette.ink500
        public static let onGround = Palette.white
        public static let onGroundMuted = Color(.sRGB, red: 255 / 255, green: 255 / 255, blue: 255 / 255, opacity: 0.85)
        public static let onGroundDark = Palette.ink900
        public static let onInverse = Palette.white
        public static let accentCoral = Palette.coral500
    }

    public enum Data {
        public static let earned = Palette.yellow500
        public static let saved = Palette.green500
        public static let spent = Palette.coral500
        public static let categorySnacks = Palette.coral500
        public static let categoryGames = Palette.violet500
        public static let categoryToys = Palette.yellow500
        public static let categoryRides = Palette.sky500
        public static let categoryClothes = Palette.green500
        public static let categoryBills = Palette.ink500
        public static let categoryGifts = Palette.gold500
        public static let categoryOther = Palette.rust500
        public static let streakActive = Palette.yellow500
        public static let streakInactive = Color(.sRGB, red: 255 / 255, green: 255 / 255, blue: 255 / 255, opacity: 0.28)
        public static let chipEarned = Palette.yellow500
        public static let chipSaved = Palette.green100
        public static let chipSpent = Palette.coral100
    }

    public enum Archetype {
        public static let squirrel = Palette.rust500
        public static let hustler = Palette.gold500
        public static let treatYourself = Palette.coral500
        public static let slowBurner = Palette.violet500
    }

    public enum Indicator {
        public static let pageActiveOnGround = Palette.white
        public static let pageInactiveOnGround = Color(.sRGB, red: 255 / 255, green: 255 / 255, blue: 255 / 255, opacity: 0.45)
        public static let pageActiveOnLight = Palette.ink900
        public static let pageInactiveOnLight = Color(.sRGB, red: 30 / 255, green: 27 / 255, blue: 58 / 255, opacity: 0.35)
    }

    public enum State {
        public static let selectedRing = Palette.white
        public static let focusRing = Palette.violet500
        public static let invalidRing = Palette.coral500
        public static let pressedLightnessDelta: Double = -0.08
        public static let disabledOpacity: Double = 0.4
        public static let unselectedOpacity: Double = 0.6
        public static let rowPressedWash = Color(.sRGB, red: 30 / 255, green: 27 / 255, blue: 58 / 255, opacity: 0.04)
    }

    public enum FontFamily {
        public static let display = "Fredoka"
        public static let text = "Nunito"
    }

    public enum TypeRamp {
        public static let displayXl = TextStyle(family: FontFamily.display, weight: .bold, size: 80, lineHeight: 1, letterSpacing: -0.02, uppercase: false)
        public static let displayL = TextStyle(family: FontFamily.display, weight: .bold, size: 68, lineHeight: 0.98, letterSpacing: -0.02, uppercase: false)
        public static let headlineL = TextStyle(family: FontFamily.display, weight: .bold, size: 48, lineHeight: 1.02, letterSpacing: -0.01, uppercase: false)
        public static let headlineM = TextStyle(family: FontFamily.display, weight: .bold, size: 40, lineHeight: 1.04, letterSpacing: -0.01, uppercase: false)
        public static let headlineMCompact = TextStyle(family: FontFamily.display, weight: .bold, size: 36, lineHeight: 1.04, letterSpacing: -0.01, uppercase: false)
        public static let statBadge = TextStyle(family: FontFamily.display, weight: .bold, size: 30, lineHeight: 1, letterSpacing: 0, uppercase: false)
        public static let titleM = TextStyle(family: FontFamily.display, weight: .semibold, size: 20, lineHeight: 1.1, letterSpacing: 0, uppercase: false)
        public static let titleS = TextStyle(family: FontFamily.display, weight: .semibold, size: 18, lineHeight: 1.1, letterSpacing: 0, uppercase: false)
        public static let bodyM = TextStyle(family: FontFamily.text, weight: .bold, size: 16, lineHeight: 1.4, letterSpacing: 0, uppercase: false)
        public static let bodyS = TextStyle(family: FontFamily.text, weight: .bold, size: 14, lineHeight: 1.4, letterSpacing: 0, uppercase: false)
        public static let labelM = TextStyle(family: FontFamily.text, weight: .heavy, size: 14, lineHeight: 1.2, letterSpacing: 0, uppercase: false)
        public static let labelS = TextStyle(family: FontFamily.text, weight: .heavy, size: 13, lineHeight: 1.2, letterSpacing: 0, uppercase: false)
        public static let captionUpper = TextStyle(family: FontFamily.text, weight: .heavy, size: 12, lineHeight: 1.2, letterSpacing: 0.08, uppercase: true)
        public static let captionPill = TextStyle(family: FontFamily.text, weight: .heavy, size: 10, lineHeight: 1.2, letterSpacing: 0.02, uppercase: false)
        public static let captionUnit = TextStyle(family: FontFamily.text, weight: .heavy, size: 9, lineHeight: 1, letterSpacing: 0.1, uppercase: true)
    }

    public enum Radius {
        public static let card: CGFloat = 28
        public static let inset: CGFloat = 16
        public static let pill: CGFloat = 999
    }

    public enum Space {
        public static let _1: CGFloat = 4
        public static let _2: CGFloat = 6
        public static let _3: CGFloat = 8
        public static let _4: CGFloat = 12
        public static let _5: CGFloat = 14
        public static let _6: CGFloat = 16
        public static let _7: CGFloat = 22
        public static let _8: CGFloat = 24
        public static let _9: CGFloat = 28
        public static let _10: CGFloat = 56
        public static let overlap: CGFloat = 44
        public static let cardInsetUnderIllustration: CGFloat = 56
    }

    public enum Size {
        public static let screenWidth: CGFloat = 390
        public static let statHeroM: CGFloat = 56
        public static let statHeroS: CGFloat = 40
        public static let amountL: CGFloat = 48
        public static let amountM: CGFloat = 36
        public static let amountS: CGFloat = 28
        public static let screenHeight: CGFloat = 844
        public static let buttonHeight: CGFloat = 52
        public static let streakBadge: CGFloat = 64
        public static let iconWell: CGFloat = 30
        public static let iconGlyph: CGFloat = 16
        public static let iconButton: CGFloat = 20
        public static let trackGoal: CGFloat = 22
        public static let trackCompare: CGFloat = 18
        public static let trackCategory: CGFloat = 16
        public static let streakDot: CGFloat = 12
        public static let pageDot: CGFloat = 6
        public static let pageDotActiveWidth: CGFloat = 22
        public static let statusDot: CGFloat = 8
        public static let touchTargetMin: CGFloat = 44
        public static let fieldHeight: CGFloat = 52
        public static let sheetHandleWidth: CGFloat = 36
        public static let sheetHandleHeight: CGFloat = 5
        public static let toggleWidth: CGFloat = 51
        public static let toggleHeight: CGFloat = 31
        public static let toggleKnob: CGFloat = 27
        public static let sliderKnob: CGFloat = 28
        public static let avatarGridCell: CGFloat = 68
        public static let avatarInline: CGFloat = 56
        public static let homeHeroCard: CGFloat = 191
        public static let homeTargetCard: CGFloat = 136
    }

    public enum Motion {
        public enum Easing {
            public static let standard = UnitCurve.bezier(startControlPoint: UnitPoint(x: 0.2, y: 0.8), endControlPoint: UnitPoint(x: 0.2, y: 1))
            public static let press = UnitCurve.bezier(startControlPoint: UnitPoint(x: 0, y: 0), endControlPoint: UnitPoint(x: 0.2, y: 1))
        }
        public enum Duration {
            public static let press: TimeInterval = 0.12
            public static let enterScreen: TimeInterval = 0.32
            public static let revealBar: TimeInterval = 0.6
            public static let revealNumber: TimeInterval = 0.8
            public static let revealNumberDelay: TimeInterval = 0.15
            public static let stagger: TimeInterval = 0.08
            public static let reduced: TimeInterval = 0.15
        }
        public static let pressScale: CGFloat = 0.97
        public static let enterOffset: CGFloat = 24
    }
}
