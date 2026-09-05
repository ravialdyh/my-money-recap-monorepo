import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { useStyles } from '../styles/useStyles';
import type { Hue, InteractiveState } from './types';

export interface ChipProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'children'> {
  /** `onGround` white on colour (default), `inverse` ink, `highlight` yellow (identifies the child). */
  variant?: 'onGround' | 'inverse' | 'highlight';
  /**
   * Selection chip tinted by hue: the chip sits on that hue's `*.100` tint with ink text, and flips
   * to the ink inverse fill when `selected`. This is how the eight Spend Categories are chosen —
   * pair it with `CATEGORY_TONE[category]` and the matching `Icon`.
   */
  tone?: Hue;
  /** Selected state for a `tone` chip. Ignored by the header variants. */
  selected?: boolean;
  /** Forced state for previews. Real presses and `disabled` work without it. */
  state?: InteractiveState;
  /** Render as a static label rather than a button. Use when the chip is not tappable. */
  asLabel?: boolean;
  /** Lifts the hit area to 44pt without changing the drawn pill. Set this on any chip that navigates. */
  tappable?: boolean;
  children: ReactNode;
}
/** Uppercase caption pill used in screen headers. ~30px tall, so set `tappable` when it is a control. */
export function Chip({ variant = 'onGround', tone, selected = false, state = 'default', asLabel = false, tappable = false, children, className = '', ...rest }: ChipProps) {
  useStyles();
  const cls = `kmr-chip ${asLabel ? '' : 'kmr-pressable'} ${className}`;
  if (asLabel) return <span className={cls} data-variant={variant} data-tone={tone}>{children}</span>;
  return (
    <button
      type="button" className={cls} data-variant={variant} data-tone={tone}
      data-selected={tone ? (selected ? 'true' : 'false') : undefined}
      data-state={state} data-tappable={tappable ? 'true' : undefined}
      aria-pressed={tone ? selected : undefined}
      disabled={state === 'disabled' || rest.disabled} {...rest}
    >
      {children}
    </button>
  );
}
export function ChipRow({ children }: { children: ReactNode }) {
  useStyles();
  return <div className="kmr-chip-row">{children}</div>;
}
