import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { useStyles } from '../styles/useStyles';
import type { InteractiveState, WellTone } from './types';

export interface IconWellProps { tone: WellTone; children: ReactNode; }
/** 30px tinted circle holding a 16px glyph. */
export function IconWell({ tone, children }: IconWellProps) {
  useStyles();
  return <span className="kmr-well" data-tone={tone}>{children}</span>;
}

export interface DataRowProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'children'> {
  icon: ReactNode;
  iconTone: WellTone;
  label: string;
  amount: string;
  state?: InteractiveState;
}
/** Itemised list row: icon well + label + Fredoka amount, hairline divider below. Pressed state is a 4% ink wash. */
export function DataRow({ icon, iconTone, label, amount, state = 'default', className = '', ...rest }: DataRowProps) {
  useStyles();
  return (
    <button type="button" className={`kmr-row kmr-pressable ${className}`} data-state={state} disabled={state === 'disabled' || rest.disabled} {...rest}>
      <IconWell tone={iconTone}>{icon}</IconWell>
      <span className="kmr-row-label">{label}</span>
      <span className="kmr-row-amount">{amount}</span>
    </button>
  );
}
export function DataRowList({ children }: { children: ReactNode }) {
  useStyles();
  return <div className="kmr-row-list">{children}</div>;
}
