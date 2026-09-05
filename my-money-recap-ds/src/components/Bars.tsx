import type { ReactNode } from 'react';
import { useStyles } from '../styles/useStyles';
import type { BarTone } from './types';

interface TrackProps { tone: BarTone; size: 'goal' | 'compare' | 'category'; percent: number; animate?: boolean; marker?: ReactNode; }
function Track({ tone, size, percent, animate = false, marker }: TrackProps) {
  const pct = Math.max(0, Math.min(100, percent));
  return (
    <div className="mmr-track" data-tone={tone} data-size={size} role="progressbar" aria-valuenow={Math.round(pct)} aria-valuemin={0} aria-valuemax={100}>
      <div className="mmr-fill" data-animate={animate ? 'true' : undefined} style={{ width: `${pct}%` }} />
      {marker ? <div className="mmr-track-marker" style={{ left: `${pct}%` }}>{marker}</div> : null}
    </div>
  );
}

export interface ProgressBarProps {
  value: number; max: number; tone?: BarTone; animate?: boolean;
  /** Illustrated marker that rides the end of the fill (e.g. PiggyMarker). Needs ~30px clearance above. */
  marker?: ReactNode;
}
/** 22px goal progress bar with an optional illustrated marker. */
export function ProgressBar({ value, max, tone = 'coral', animate, marker }: ProgressBarProps) {
  useStyles();
  return <Track tone={tone} size="goal" percent={(value / max) * 100} animate={animate} marker={marker} />;
}

export interface CategoryBarProps { label: string; percent: number; tone: BarTone; animate?: boolean; }
/** Label row (name / %) above a 16px bar. Stack three for a spending split. */
export function CategoryBar({ label, percent, tone, animate }: CategoryBarProps) {
  useStyles();
  return (
    <div className="mmr-category">
      <div className="mmr-bar-labels"><span>{label}</span><span>{Math.round(percent)}%</span></div>
      <Track tone={tone} size="category" percent={percent} animate={animate} />
    </div>
  );
}

export interface CompareRow { label: string; value: number; display: string; tone: BarTone; }
export interface CompareBarsProps {
  rows: CompareRow[];
  /**
   * How the bars are scaled. `total` (default) reads each row as a share of the whole, which is
   * right when the rows partition one amount ("Saved" and "Spent" both come out of Money In).
   * `max` scales against the largest row, which is right when one row is a benchmark the other is
   * measured against ("Kept" against "Target").
   */
  scale?: 'total' | 'max';
  animate?: boolean;
}
/** Two or more 18px bars: "Saved £640" vs "Spent £1,760". See `scale` for how the widths are derived. */
export function CompareBars({ rows, scale = 'total', animate }: CompareBarsProps) {
  useStyles();
  const basis = scale === 'max'
    ? Math.max(...rows.map((r) => r.value), 0) || 1
    : rows.reduce((s, r) => s + r.value, 0) || 1;
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      {rows.map((r) => (
        <div key={r.label} className="mmr-compare-row">
          <span className="mmr-compare-label">{r.label}</span>
          <Track tone={r.tone} size="compare" percent={(r.value / basis) * 100} animate={animate} />
          <span className="mmr-compare-value">{r.display}</span>
        </div>
      ))}
    </div>
  );
}
