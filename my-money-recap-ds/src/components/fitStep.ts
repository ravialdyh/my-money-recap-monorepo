import { useEffect, useLayoutEffect, useState, type RefObject } from 'react';

/**
 * Fit-to-width sizing (PRD §11, D71): a number is drawn at the largest of a few fixed sizes that fits
 * its container, measured, never guessed from a character count. The steps are `l` → `m` → `s`; the
 * CSS maps each to a token (StatHero 80 → 56 → 40px, AmountField 48 → 36 → 28px). This is the web
 * equivalent of SwiftUI's `.lineLimit(1).minimumScaleFactor(…)`, limited to three sizes so the
 * type ramp stays discrete.
 */
export type FitStep = 'l' | 'm' | 's';
export const FIT_STEPS: readonly FitStep[] = ['l', 'm', 's'];

/** Pure: the first step that fits, or the smallest. `fits` is called in order and may have side effects (setting the step before measuring). */
export function pickStep<T>(steps: readonly T[], fits: (step: T) => boolean): T {
  for (const step of steps) if (fits(step)) return step;
  return steps[steps.length - 1]!;
}

const useIsoLayoutEffect = typeof window === 'undefined' ? useEffect : useLayoutEffect;

/**
 * Measures `text` against `box` (the width it must fit; defaults to the text element itself) and
 * returns the step to draw. During measurement `data-step` is set on both elements, so the CSS may key
 * off either. Re-measures when `key` changes, when the box resizes, and once web fonts have loaded.
 * Before the first measurement, and on the server, the step is `l`.
 */
export function useFitStep(text: RefObject<HTMLElement>, box: RefObject<HTMLElement> | null, key: string): FitStep {
  const [step, setStep] = useState<FitStep>('l');
  useIsoLayoutEffect(() => {
    const el = text.current;
    if (!el) return;
    const measure = () => {
      const container = box?.current ?? el;
      const chosen = pickStep(FIT_STEPS, (s) => {
        el.dataset.step = s;
        container.dataset.step = s;
        const available = container === el ? el.clientWidth : innerWidth(container);
        return el.scrollWidth <= available + 0.5;
      });
      setStep(chosen);
    };
    measure();
    const observed = box?.current ?? el;
    const ro = typeof ResizeObserver === 'undefined' ? null : new ResizeObserver(measure);
    ro?.observe(observed);
    const fonts = typeof document !== 'undefined' ? document.fonts : undefined;
    fonts?.ready.then(measure).catch(() => {});
    return () => ro?.disconnect();
  }, [text, box, key]);
  return step;
}

function innerWidth(el: HTMLElement): number {
  const cs = getComputedStyle(el);
  return el.clientWidth - parseFloat(cs.paddingLeft) - parseFloat(cs.paddingRight);
}
