import { type RefObject, useEffect, useLayoutEffect, useState } from 'react';

/**
 * Fit-to-width sizing (PRD §11, D71): `l` → `m` → `s`, each mapped to a token by the CSS. Three steps,
 * so the type ramp stays discrete. README "Design rules" has the sizes.
 */
export type FitStep = 'l' | 'm' | 's';
const STEPS: readonly FitStep[] = ['l', 'm', 's'];

const useIsoLayoutEffect = typeof window === 'undefined' ? useEffect : useLayoutEffect;

/**
 * Measures `text` against `box` (default: the text element) and returns the step to draw, setting
 * `data-step` on both so the CSS may key off either. Re-measures on `key`, resize and `fonts.ready`;
 * before the first measurement the step is `l`.
 */
export function useFitStep(
  text: RefObject<HTMLElement | null>,
  box: RefObject<HTMLElement | null> | null,
  key: string,
): FitStep {
  const [step, setStep] = useState<FitStep>('l');
  useIsoLayoutEffect(() => {
    const el = text.current;
    if (!el) return;
    const measure = () => {
      const container = box?.current ?? el;
      const fits = (s: FitStep) => {
        el.dataset.step = s;
        container.dataset.step = s;
        const available = container === el ? el.clientWidth : innerWidth(container);
        return el.scrollWidth <= available + 0.5;
      };
      setStep(STEPS.find(fits) ?? 's');
    };
    measure();
    const ro = typeof ResizeObserver === 'undefined' ? null : new ResizeObserver(measure);
    ro?.observe(el);
    if (box?.current && box.current !== el) ro?.observe(box.current);
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
