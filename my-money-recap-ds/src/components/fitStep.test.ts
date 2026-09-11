import { describe, expect, it } from 'vitest';
import { FIT_STEPS, pickStep } from './fitStep';

describe('pickStep (PRD §11, D71: the largest size that fits)', () => {
  const widthsAt = { l: 289, m: 202, s: 144 }; // "Rp 15.000.000" at 48 / 36 / 28px, measured in Fredoka 700
  const fitsIn = (available: number) => (step: keyof typeof widthsAt) => widthsAt[step] <= available;

  it('keeps the largest size when it fits', () => {
    expect(pickStep(FIT_STEPS, fitsIn(298))).toBe('l');
  });
  it('steps down once when the largest overflows', () => {
    expect(pickStep(FIT_STEPS, fitsIn(250))).toBe('m');
  });
  it('steps down twice when the middle overflows', () => {
    expect(pickStep(FIT_STEPS, fitsIn(180))).toBe('s');
  });
  it('returns the smallest size when nothing fits, never throws', () => {
    expect(pickStep(FIT_STEPS, fitsIn(100))).toBe('s');
  });
  it('tries the steps in order and stops at the first fit', () => {
    const tried: string[] = [];
    pickStep(FIT_STEPS, (s) => { tried.push(s); return s === 'm'; });
    expect(tried).toEqual(['l', 'm']);
  });
});
