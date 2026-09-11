import { useRef } from 'react';
import { useStyles } from '../styles/useStyles';
import { TextField, type FieldState } from './TextField';
import { useFitStep } from './fitStep';
import { currencySymbol, formatAmountNumber, parseAmount } from '../money';

export interface AmountFieldProps {
  /** The amount as an exact decimal string ("15000000", "12.40"), the same form it is stored in (PRD §7). */
  value: string;
  /** ISO 4217 code. The symbol and the decimals the field shows come from it, via the platform formatter. */
  currency: string;
  /** Viewer's language tag; omit for the device language. Sets the separators. */
  locale?: string;
  label?: string;
  helper?: string;
  state?: FieldState;
  /** Called with the typed amount as a decimal string; "" when the field is empty. */
  onChange?: (next: string) => void;
}
/**
 * Currency-aware amount entry: a TextField with a fixed symbol and the Currency's decimals, both from the
 * platform formatter. The value is Fredoka 48px, or 36 / 28px when the symbol plus value does not fit the
 * field (PRD §11, D71) — measured, not counted.
 */
export function AmountField({ value, currency, locale, label, helper, state = 'default', onChange }: AmountFieldProps) {
  useStyles();
  const symbol = currencySymbol(currency, locale);
  const formatted = value === '' ? '' : formatAmountNumber(value, currency, locale);
  const mirror = useRef<HTMLSpanElement>(null);
  const field = useRef<HTMLSpanElement>(null); // the padded row: what the symbol plus value must fit inside
  const step = useFitStep(mirror, field, `${symbol} ${formatted}`);
  return (
    <TextField
      fieldRef={field}
      className="mmr-amount"
      data-step={step}
      label={label}
      helper={helper}
      state={state}
      prefix={symbol}
      inputMode="decimal"
      value={formatted}
      onChange={(e) => onChange?.(parseAmount(e.currentTarget.value, locale))}
      aria-label={label ?? `Amount in ${currency}`}
      mirror={<span ref={mirror} className="mmr-amount-mirror" aria-hidden="true">{symbol} {formatted}</span>}
    />
  );
}
