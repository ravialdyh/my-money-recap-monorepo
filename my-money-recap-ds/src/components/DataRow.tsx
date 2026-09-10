import type { ButtonHTMLAttributes, CSSProperties, ReactNode } from 'react';
import { useStyles } from '../styles/useStyles';
import type { InteractiveState, WellTone } from './types';

export interface IconWellProps {
  tone: WellTone;
  /** Circle diameter in px. Defaults to the 30px `size.icon-well` token; `PromptCard`'s chevron well is 44. */
  size?: number;
  /** Glyph size in px. Defaults to the 16px `size.icon-glyph` token; pass 20 for a stroked glyph in a 44px well. */
  glyphSize?: number;
  children: ReactNode;
}
/** Tinted circle holding a glyph. 30px around 16px by default. */
export function IconWell({ tone, size, glyphSize, children }: IconWellProps) {
  useStyles();
  const style: CSSProperties = {};
  if (size !== undefined) { style.width = size; style.height = size; }
  if (glyphSize !== undefined) (style as Record<string, string>)['--mmr-well-glyph-size'] = `${glyphSize}px`;
  return <span className="mmr-well" data-tone={tone} style={size === undefined && glyphSize === undefined ? undefined : style}>{children}</span>;
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
    <button type="button" className={`mmr-row mmr-pressable ${className}`} data-state={state} disabled={state === 'disabled' || rest.disabled} {...rest}>
      <IconWell tone={iconTone}>{icon}</IconWell>
      <span className="mmr-row-label">{label}</span>
      <span className="mmr-row-amount">{amount}</span>
    </button>
  );
}
export function DataRowList({ children }: { children: ReactNode }) {
  useStyles();
  return <div className="mmr-row-list">{children}</div>;
}
