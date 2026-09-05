import { useStyles } from '../styles/useStyles';

export interface StreakBadgeProps { value: number; unit?: string; }
/** 64px yellow circle with a Fredoka number and a tiny uppercase unit. */
export function StreakBadge({ value, unit = 'days' }: StreakBadgeProps) {
  useStyles();
  return (
    <div className="mmr-badge" role="img" aria-label={`${value} ${unit}`}>
      <div className="mmr-badge-value">{value}</div>
      <div className="mmr-badge-unit">{unit}</div>
    </div>
  );
}

export interface StreakDotsProps {
  /** Days in the month (dots rendered). */
  total?: number;
  /** 1-based inclusive day range that is active, e.g. [12, 20]. */
  activeRange?: [number, number];
  /** Or an explicit set of active day numbers. */
  activeDays?: number[];
}
/** 15-column grid of 12px dots. Two rows for a 30-day month. */
export function StreakDots({ total = 30, activeRange, activeDays }: StreakDotsProps) {
  useStyles();
  const active = new Set<number>(activeDays ?? []);
  if (activeRange) for (let d = activeRange[0]; d <= activeRange[1]; d++) active.add(d);
  return (
    <div className="mmr-dots" role="img" aria-label={`${active.size} active days of ${total}`}>
      {Array.from({ length: total }, (_, i) => (
        <span key={i} className="mmr-dot" data-active={active.has(i + 1) ? 'true' : 'false'} />
      ))}
    </div>
  );
}
