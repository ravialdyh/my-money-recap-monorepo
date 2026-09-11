/**
 * PRD §11 (D45, D57): a formatted amount steps down in size as it grows, so "Rp 15.000.000" fits
 * where "£2,400" did. The boundaries are counted on the *formatted string* — symbol, separators and
 * spaces included — because that is what has to fit: "Rp 520.000" is 10 glyphs, "£520,000" is 8.
 *
 *   up to 7 glyphs   → 'l'  (StatHero 80px, AmountField 48px)
 *   8 to 10 glyphs   → 'm'  (56px, 36px)
 *   11 and more      → 's'  (40px, 28px)
 */
export type StepDown = 'l' | 'm' | 's';
export function stepDownFor(formatted: string): StepDown {
  const glyphs = Array.from(formatted).length;
  return glyphs > 10 ? 's' : glyphs > 7 ? 'm' : 'l';
}
