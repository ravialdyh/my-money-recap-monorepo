import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { ensureStyles } from '../styles/ensureStyles';
import { ARCHETYPE_LABEL, type Archetype, type InteractiveState } from './types';

export interface ArchetypePillProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'children'> {
  archetype: Archetype;
  selected?: boolean;
  state?: InteractiveState;
  /** Lifts the hit area to 44pt without changing the drawn pill. Set this wherever the pill is a control. */
  tappable?: boolean;
  children?: ReactNode;
}
/** Small coloured pill per money personality. Unselected sits at 60% opacity; selected is full opacity with a 2px white ring. ~24px tall, so set `tappable` when it is a control. */
export function ArchetypePill({
  archetype,
  selected = false,
  state = 'default',
  tappable = false,
  children,
  className = '',
  ...rest
}: ArchetypePillProps) {
  ensureStyles();
  const isSelected = selected || state === 'selected';
  return (
    <button
      type="button"
      className={`mmr-archetype mmr-pressable ${className}`}
      data-archetype={archetype}
      data-selected={isSelected ? 'true' : 'false'}
      data-state={state}
      data-tappable={tappable ? 'true' : undefined}
      aria-pressed={isSelected}
      disabled={state === 'disabled' || rest.disabled}
      {...rest}
    >
      {children ?? ARCHETYPE_LABEL[archetype]}
    </button>
  );
}
export function ArchetypeRow({ children }: { children: ReactNode }) {
  ensureStyles();
  return <div className="mmr-archetype-row">{children}</div>;
}
