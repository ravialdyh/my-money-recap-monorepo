import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { Icon } from '../icons';
import { ensureStyles } from '../styles/ensureStyles';
import type { InteractiveState } from './types';

export interface ListRowProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'children' | 'value'> {
  label: string;
  /** Right-aligned secondary value, e.g. the current Currency or Payday. */
  value?: ReactNode;
  /** Trailing affordance. `chevron` opens a screen, `check` marks a selection, `none` for a plain row. */
  trailing?: 'chevron' | 'check' | 'none';
  /** Selected rows (the Currency list) show the violet check. */
  selected?: boolean;
  /** Destructive rows (Delete account) render the label in coral. */
  tone?: 'default' | 'destructive';
  state?: InteractiveState;
}
/** Settings and Currency list row: a DataRow with no IconWell and a chevron or check instead of an amount. */
export function ListRow({
  label,
  value,
  trailing = 'chevron',
  selected = false,
  tone = 'default',
  state = 'default',
  className = '',
  ...rest
}: ListRowProps) {
  ensureStyles();
  return (
    <button
      type="button"
      className={`mmr-row mmr-list-row mmr-pressable ${className}`}
      data-state={state}
      data-tone={tone}
      aria-pressed={trailing === 'check' ? selected : undefined}
      disabled={state === 'disabled' || rest.disabled}
      {...rest}
    >
      <span className="mmr-row-label">{label}</span>
      {value ? <span className="mmr-list-row-value">{value}</span> : null}
      {trailing === 'chevron' ? (
        <span className="mmr-list-row-chevron">
          <Icon name="chevron" size={18} />
        </span>
      ) : null}
      {trailing === 'check' && selected ? <span className="mmr-list-row-check" aria-hidden="true" /> : null}
    </button>
  );
}
