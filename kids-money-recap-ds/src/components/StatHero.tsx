import { useEffect, useState } from 'react';
import { useStyles } from '../styles/useStyles';
import { Eyebrow } from './Headline';

export interface StatHeroProps {
  eyebrow: string;
  /** Numeric value, formatted with `format`. */
  value: number;
  format?: (n: number) => string;
  /** Count up over motion.duration.reveal-number. Respects prefers-reduced-motion. */
  animate?: boolean;
}
const defaultFormat = (n: number) => `£${Math.round(n)}`;

/** The one big number on a screen: eyebrow + 80px Fredoka value. */
export function StatHero({ eyebrow, value, format = defaultFormat, animate = false }: StatHeroProps) {
  useStyles();
  const shown = useCountUp(value, animate);
  return (
    <div className="kmr-stat-hero">
      <Eyebrow>{eyebrow}</Eyebrow>
      <div className="kmr-stat-hero-value">{format(shown)}</div>
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
