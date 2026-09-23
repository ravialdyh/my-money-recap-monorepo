import type { HTMLAttributes } from 'react';
import { useEffect, useRef } from 'react';
import { ensureStyles } from '../styles/ensureStyles';
import { HelperText } from './Headline';
import type { InteractiveState } from './types';

/** A Payday: a day of the month, or the month's last day (PRD D51). */
export type PaydayValue = number | 'last';
/** 1…31 then `last`, in the order the picker draws them. */
const PAYDAY_VALUES: readonly PaydayValue[] = [...Array.from({ length: 31 }, (_, i) => i + 1), 'last'] as const;

export interface DayPickerProps extends Omit<HTMLAttributes<HTMLDivElement>, 'onChange'> {
  /** The chosen Payday. The row scrolls to centre it when the picker opens. */
  value: PaydayValue;
  onChange?: (value: PaydayValue) => void;
  /**
   * The three formatted dates that follow `value`, as `['30 Nov', '31 Dec', '31 Jan']`. This component
   * does no calendar arithmetic: `min(Payday, days in month)` is the caller's (PRD §8.2, D51).
   * `null`, the default, removes the line.
   */
  nextPaydays?: readonly string[] | null;
  state?: InteractiveState;
}
/**
 * Centres the selected pill. Called once, when the picker opens: re-centring on every tap would
 * drag the row out from under the Member's finger.
 */
function centreSelectionOnOpen(row: HTMLDivElement | null) {
  const pill = row?.querySelector<HTMLElement>('[data-selected="true"]');
  if (!row || !pill) return;
  row.scrollLeft = pill.offsetLeft - (row.clientWidth - pill.offsetWidth) / 2;
}
/**
 * Payday picker: a horizontally scrolling row of 44pt pills, 1–31 plus `Last day`, the selection
 * drawn as the ink pill. Onboarding step 3 and the Settings salary screen (PRD §6.6, D51).
 */
export function DayPicker({
  value,
  onChange,
  nextPaydays = null,
  state = 'default',
  className = '',
  ...rest
}: DayPickerProps) {
  ensureStyles();
  const row = useRef<HTMLDivElement>(null);
  useEffect(() => centreSelectionOnOpen(row.current), []);
  const disabled = state === 'disabled';
  return (
    <div className={`mmr-day-picker ${className}`} data-state={state} {...rest}>
      <div className="mmr-day-picker-row" ref={row} role="radiogroup" aria-label="Payday">
        {PAYDAY_VALUES.map((day) => (
          <button
            key={String(day)}
            type="button"
            className="mmr-day-pill mmr-pressable"
            role="radio"
            aria-checked={day === value}
            data-selected={day === value ? 'true' : 'false'}
            data-wide={day === 'last' ? 'true' : undefined}
            disabled={disabled}
            onClick={onChange ? () => onChange(day) : undefined}
          >
            {day === 'last' ? 'Last day' : day}
          </button>
        ))}
      </div>
      {nextPaydays && nextPaydays.length > 0 ? (
        <HelperText className="mmr-day-picker-helper">Next paydays: {nextPaydays.join(' · ')}</HelperText>
      ) : null}
    </div>
  );
}
