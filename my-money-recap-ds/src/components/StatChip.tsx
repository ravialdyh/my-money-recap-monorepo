import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { ensureStyles } from '../styles/ensureStyles';
import type { DataTone, InteractiveState } from './types';

export interface StatChipProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'children'> {
  /** Data tone drives both the tint and the status dot. */
  tone: DataTone;
  state?: InteractiveState;
  children: ReactNode;
}
/** Tinted pill with an 8px status dot: "Money in £2,400", "Saved £640", "Spent £1,760". */
export function StatChip({ tone, state = 'default', children, className = '', ...rest }: StatChipProps) {
  ensureStyles();
  return (
    <button
      type="button"
      className={`mmr-stat-chip mmr-pressable ${className}`}
      data-tone={tone}
      data-state={state}
      disabled={state === 'disabled' || rest.disabled}
      {...rest}
    >
      <span className="mmr-stat-chip-dot" />
      {children}
    </button>
  );
}
export function StatChipRow({ children }: { children: ReactNode }) {
  ensureStyles();
  return <div className="mmr-stat-chip-row">{children}</div>;
}
