import { useEffect, useRef, useState } from 'react';
import { useStyles } from '../styles/useStyles';
import { Eyebrow } from './Headline';
import { useFitStep } from './fitStep';
import { formatMoney, type MoneyFormat } from '../money';

export interface StatHeroProps {
  eyebrow: string;
  /** Numeric value in major units, formatted with `format`, or as money in `currency`. */
  value: number;
  /** ISO 4217 code the value is formatted in when `format` is not given. */
  currency?: string;
  /** Viewer's language tag; omit for the device language. */
  locale?: string;
  /** Custom formatter for values that are not money (e.g. "9 days"). Overrides `currency`. */
  format?: MoneyFormat;
  /** Count up over motion.duration.reveal-number. Respects prefers-reduced-motion. */
  animate?: boolean;
}

/**
 * The one big number on a screen: eyebrow + Fredoka value at 80px, or 56 / 40px when the full
 * string does not fit the card (PRD §11, D71) — measured on the final value, so a count-up never
 * resizes mid-animation.
 */
export function StatHero({ eyebrow, value, currency = 'GBP', locale, format, animate = false }: StatHeroProps) {
  useStyles();
  const fmt: MoneyFormat = format ?? ((n) => formatMoney(n, currency, locale));
  const shown = useCountUp(value, animate);
  const ref = useRef<HTMLDivElement>(null);
  const final = fmt(value);
  const step = useFitStep(ref, null, final);
  // Count-up frames are whole units so an intermediate frame never grows a fraction the final value lacks.
  const text = shown === value ? final : fmt(Math.round(shown));
  return (
    <div className="mmr-stat-hero">
      <Eyebrow>{eyebrow}</Eyebrow>
      <div ref={ref} className="mmr-stat-hero-value" data-step={step}>{text}</div>
    </div>
  );
}

function useCountUp(target: number, enabled: boolean, durationMs = 800, delayMs = 150): number {
  const [v, setV] = useState(enabled ? 0 : target);
  useEffect(() => {
    if (!enabled) { setV(target); return; }
    const reduced = typeof window !== 'undefined' && window.matchMedia?.('(prefers-reduced-motion: reduce)').matches;
    if (reduced) { setV(target); return; }
    let raf = 0; let start = 0;
    const ease = (t: number) => 1 - Math.pow(1 - t, 3);
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
