import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { useStyles } from '../styles/useStyles';
import type { DataTone, InteractiveState } from './types';

export interface StatChipProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'children'> {
  /** Data tone drives both the tint and the status dot. */
  tone: DataTone;
  state?: InteractiveState;
  children: ReactNode;
}
/** Tinted pill with an 8px status dot: "Earned £24", "Saved £25", "Spent £11". */
export function StatChip({ tone, state = 'default', children, className = '', ...rest }: StatChipProps) {
  useStyles();
  return (
    <button type="button" className={`kmr-stat-chip kmr-pressable ${className}`} data-tone={tone} data-state={state} disabled={state === 'disabled' || rest.disabled} {...rest}>
      <span className="kmr-stat-chip-dot" />
      {children}
    </button>
  );
}
export function StatChipRow({ children }: { children: ReactNode }) {
  useStyles();
  return <div className="kmr-stat-chip-row">{children}</div>;
}
