import { useStyles } from '../styles/useStyles';
import type { BarTone } from './types';

export interface SliderProps {
  value: number;
  min?: number;
  max?: number;
  /** Saving Target moves in steps of 5 (PRD §6.7). */
  step?: number;
  tone?: BarTone;
  /** Rendered under the track, e.g. "£480 of £2,400 each month". */
  helper?: string;
  disabled?: boolean;
  onChange?: (next: number) => void;
  'aria-label'?: string;
}
/** The ProgressBar track with a draggable knob. Used for the Saving Target percentage. */
export function Slider({ value, min = 0, max = 100, step = 5, tone = 'violet', helper, disabled = false, onChange, ...rest }: SliderProps) {
  useStyles();
  const pct = Math.max(0, Math.min(100, ((value - min) / (max - min)) * 100));
  return (
    <div className="kmr-slider" data-disabled={disabled ? 'true' : undefined}>
      <div className="kmr-track" data-tone={tone} data-size="goal">
        <div className="kmr-fill" style={{ width: `${pct}%` }} />
        <span className="kmr-slider-knob" style={{ left: `${pct}%` }} />
      </div>
      <input
        className="kmr-slider-input" type="range" min={min} max={max} step={step} value={value} disabled={disabled}
        onChange={(e) => onChange?.(Number(e.currentTarget.value))} {...rest}
      />
      {helper ? <span className="kmr-field-helper">{helper}</span> : null}
    </div>
  );
}
