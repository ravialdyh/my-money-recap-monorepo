import { useEffect, useRef, useState } from 'react';
import { formatMoney, type MoneyFormat } from '../money';
import { ensureStyles } from '../styles/ensureStyles';
import tokens from '../tokens/tokens.json';
import { useFitStep } from './fitStep';
import { Eyebrow } from './Headline';

interface StatHeroBase {
  eyebrow: string;
  /** Numeric value in major units, formatted with `format`, or as money in `currency`. */
  value: number;
  /** Count up over motion.duration.reveal-number. Respects prefers-reduced-motion. */
  animate?: boolean;
}
/**
 * Money needs its currency: there is no default, so a screen that forgets it fails to type-check
 * instead of silently drawing pounds. Values that are not money pass `format` instead ("9 days").
 */
export type StatHeroProps = StatHeroBase &
  (
    | {
        /** ISO 4217 code the value is formatted in. */
        currency: string;
        /** Viewer's language tag; omit for the device language. */
        locale?: string;
        format?: undefined;
      }
    | {
        /** Custom formatter for values that are not money. */ format: MoneyFormat;
        currency?: undefined;
        locale?: undefined;
      }
  );

/**
 * The one big number on a screen: eyebrow + Fredoka value at 80px, or 56 / 40px when the full
 * string does not fit the card (PRD §11, D71) — measured on the final value, so a count-up never
 * resizes mid-animation.
 */
export function StatHero({ eyebrow, value, currency, locale, format, animate = false }: StatHeroProps) {
  ensureStyles();
  const fmt: MoneyFormat = format ?? ((n) => formatMoney(n, currency as string, locale));
  const shown = useCountUp(value, animate);
  const ref = useRef<HTMLDivElement>(null);
  const final = fmt(value);
  const step = useFitStep(ref, null, final);
  const text = shown === value ? final : fmt(Math.round(shown));
  return (
    <div className="mmr-stat-hero">
      <Eyebrow>{eyebrow}</Eyebrow>
      <div ref={ref} className="mmr-stat-hero-value" data-step={step}>
        {text}
      </div>
    </div>
  );
}

const REVEAL_MS = parseFloat(tokens.motion.duration['reveal-number'].$value);
const REVEAL_DELAY_MS = parseFloat(tokens.motion.duration['reveal-number-delay'].$value);

function useCountUp(target: number, enabled: boolean, durationMs = REVEAL_MS, delayMs = REVEAL_DELAY_MS): number {
  const [v, setV] = useState(enabled ? 0 : target);
  useEffect(() => {
    if (!enabled) {
      setV(target);
      return;
    }
    const reduced = typeof window !== 'undefined' && window.matchMedia?.('(prefers-reduced-motion: reduce)').matches;
    if (reduced) {
      setV(target);
      return;
    }
    let raf = 0;
    let start = 0;
    const ease = (t: number) => 1 - (1 - t) ** 3;
    const tick = (now: number) => {
      if (!start) start = now + delayMs;
      const t = Math.min(1, Math.max(0, (now - start) / durationMs));
      setV(target * ease(t));
      if (t < 1) raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, [target, enabled, durationMs, delayMs]);
  return v;
}
