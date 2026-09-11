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
}
/** 15-column grid of 12px dots. Two rows for a 30-day month. */
export function StreakDots({ total = 30, activeRange }: StreakDotsProps) {
  useStyles();
  const [from, to] = activeRange ?? [0, -1];
  const isActive = (day: number) => day >= from && day <= to;
  const count = activeRange ? to - from + 1 : 0;
  return (
    <div className="mmr-dots" role="img" aria-label={`${count} active days of ${total}`}>
      {Array.from({ length: total }, (_, i) => (
        <span key={i} className="mmr-dot" data-active={isActive(i + 1) ? 'true' : 'false'} />
      ))}
    </div>
  );
}
