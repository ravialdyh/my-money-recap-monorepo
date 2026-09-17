// Generates apps/ios/DesignSystem/Tokens.swift from src/tokens/tokens.json (docs/IOS.md I5).
// Colours, radii, spacing, sizes, the type ramp and motion. Aliases like "{palette.violet.500}" become Palette.violet500.
import { mkdirSync, writeFileSync } from 'node:fs';
import { resolve, dirname } from 'node:path';
import { here, tokens, walk } from './lib.mjs';

const out = resolve(here, '../../apps/ios/DesignSystem/Tokens.swift');

const camel = (s) => s.replace(/-([a-z0-9])/g, (_, c) => c.toUpperCase());
const pascal = (s) => camel(s).replace(/^[a-z]/, (c) => c.toUpperCase());
// ["violet", "500"] -> violet500; ["card-accent"] -> cardAccent
const ident = (segs) => camel(segs.map((s, i) => (i === 0 ? s : pascal(s))).join('')).replace(/^(\d)/, '_$1');
// "{palette.violet.500}" -> Palette.violet500; "{font.family.display}" -> FontFamily.display
const alias = (v) => {
  const segs = v.slice(1, -1).split('.');
  if (segs[0] === 'font') return `${pascal(segs[0])}${pascal(segs[1])}.${ident(segs.slice(2))}`;
  return `${pascal(segs[0])}.${ident(segs.slice(1))}`;
};
const num = (n) => String(n);
const px = (s) => num(parseFloat(s)); // "28px" -> 28
const ms = (s) => num(parseFloat(s) / 1000); // "120ms" -> 0.12

function color(v) {
  if (v.startsWith('{')) return alias(v);
  let m = v.match(/^#([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})$/i);
  if (m) return `Color(.sRGB, red: 0x${m[1]} / 255, green: 0x${m[2]} / 255, blue: 0x${m[3]} / 255, opacity: 1)`;
  m = v.match(/^rgba\((\d+),\s*(\d+),\s*(\d+),\s*([\d.]+)\)$/);
  if (m) return `Color(.sRGB, red: ${m[1]} / 255, green: ${m[2]} / 255, blue: ${m[3]} / 255, opacity: ${m[4]})`;
  return null;
}

// One `enum` of `static let`s per group: every colour and number under it.
function colorEnum(node) {
  const lines = [];
  walk(node, (p, { $value: raw }) => {
    if (typeof raw === 'number') lines.push(`        public static let ${ident(p)}: Double = ${num(raw)}`);
    else if (typeof raw === 'string' && /^-?[\d.]+%$/.test(raw)) lines.push(`        public static let ${ident(p)}: Double = ${num(parseFloat(raw) / 100)}`);
    else if (typeof raw === 'string' && color(raw)) lines.push(`        public static let ${ident(p)} = ${color(raw)}`);
  });
  return lines;
}
const block = (name, lines) => `    public enum ${name} {\n${lines.join('\n')}\n    }`;

const weights = { 500: '.medium', 600: '.semibold', 700: '.bold', 800: '.heavy' };
const typeRamp = Object.entries(tokens.type)
  .filter(([k]) => !k.startsWith('$'))
  .map(([k, { $value: t }]) => {
    const family = alias(t.fontFamily);
    const ls = t.letterSpacing ? parseFloat(t.letterSpacing) : 0;
    return `        public static let ${ident([k])} = TextStyle(family: ${family}, weight: ${weights[t.fontWeight]}, size: ${px(t.fontSize)}, lineHeight: ${num(t.lineHeight)}, letterSpacing: ${num(ls)}, uppercase: ${t.textTransform === 'uppercase'})`;
  });

const easing = Object.entries(tokens.motion.easing)
  .filter(([k]) => !k.startsWith('$'))
  .map(([k, { $value: [x1, y1, x2, y2] }]) => `            public static let ${ident([k])} = UnitCurve.bezier(startControlPoint: UnitPoint(x: ${num(x1)}, y: ${num(y1)}), endControlPoint: UnitPoint(x: ${num(x2)}, y: ${num(y2)}))`);
const duration = Object.entries(tokens.motion.duration)
  .filter(([k]) => !k.startsWith('$'))
  .map(([k, { $value: d }]) => `            public static let ${ident([k])}: TimeInterval = ${ms(d)}`);

const swift = `// Generated from my-money-recap-ds/src/tokens/tokens.json by scripts/build-swift-tokens.mjs. Do not edit by hand.
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
${block('Palette', colorEnum(tokens.palette))}

${Object.entries(tokens.color)
  .filter(([k]) => !k.startsWith('$'))
  .map(([k, v]) => block(pascal(k), colorEnum(v)))
  .join('\n\n')}

    public enum FontFamily {
${Object.entries(tokens.font.family)
  .filter(([k]) => !k.startsWith('$'))
  .map(([k, { $value: fam }]) => `        public static let ${ident([k])} = ${JSON.stringify(fam[0])}`)
  .join('\n')}
    }

${block('TypeRamp', typeRamp)}

${['radius', 'space', 'size'].map((g) => block(pascal(g), Object.entries(tokens[g])
  .filter(([k]) => !k.startsWith('$'))
  .map(([k, { $value: r }]) => `        public static let ${ident([k])}: CGFloat = ${px(r)}`))).join('\n\n')}

    public enum Motion {
        public enum Easing {
${easing.join('\n')}
        }
        public enum Duration {
${duration.join('\n')}
        }
        public static let pressScale: CGFloat = ${num(tokens.motion['press-scale'].$value)}
        public static let enterOffset: CGFloat = ${px(tokens.motion['enter-offset'].$value)}
    }
}
`;

mkdirSync(dirname(out), { recursive: true });
writeFileSync(out, swift);
console.log(`Tokens.swift: ${(swift.match(/static let/g) || []).length} tokens`);
