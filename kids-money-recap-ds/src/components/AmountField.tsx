import { useStyles } from '../styles/useStyles';
import { TextField, type FieldState } from './TextField';

/** Decimal places per ISO 4217 minor unit. IDR 0, GBP 2, BHD 3 (PRD §7). */
export const MINOR_UNITS: Record<string, number> = { GBP: 2, USD: 2, EUR: 2, SGD: 2, IDR: 0, BHD: 3 };
export const CURRENCY_SYMBOL: Record<string, string> = { GBP: '£', USD: '$', EUR: '€', SGD: 'S$', IDR: 'Rp', BHD: 'BD' };

/** Formats an integer minor-unit amount for display: 240000 GBP -> "2,400.00", 3000000 IDR -> "3,000,000". */
export function formatMinor(amountMinor: number, currency: string): string {
  const digits = MINOR_UNITS[currency] ?? 2;
  const value = amountMinor / 10 ** digits;
  return value.toLocaleString('en-GB', { minimumFractionDigits: digits, maximumFractionDigits: digits });
}

export interface AmountFieldProps {
  /** Amount in the Currency's minor units. Integer, always. */
  valueMinor: number;
  /** ISO 4217 code. Drives both the fixed symbol and the number of decimals. */
  currency: string;
  label?: string;
  helper?: string;
  state?: FieldState;
  onChange?: (next: number) => void;
}
/** Currency-aware amount entry: a TextField with a fixed symbol and minor-unit formatting. The symbol is never editable. */
export function AmountField({ valueMinor, currency, label, helper, state = 'default', onChange }: AmountFieldProps) {
  useStyles();
  const digits = MINOR_UNITS[currency] ?? 2;
  return (
    <TextField
      className="kmr-amount"
      label={label}
      helper={helper}
      state={state}
      prefix={CURRENCY_SYMBOL[currency] ?? currency}
      inputMode="decimal"
      value={formatMinor(valueMinor, currency)}
      onChange={(e) => {
        const raw = Number(e.currentTarget.value.replace(/[^0-9]/g, ''));
        onChange?.(Number.isFinite(raw) ? raw : 0);
      }}
      aria-label={label ?? `Amount in ${currency}`}
      data-minor-digits={digits}
    />
  );
}
