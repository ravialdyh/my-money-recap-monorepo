import { useStyles } from '../styles/useStyles';
import { TextField, type FieldState } from './TextField';
import { stepDownFor } from './stepDown';

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
/**
 * Currency-aware amount entry: a TextField with a fixed symbol and minor-unit formatting. The symbol is
 * never editable. The value is Fredoka 48px up to 7 glyphs, 36px to 10, 28px beyond (PRD §11), counted on
 * the symbol plus the integer part — the fraction is fixed per Currency, so "£ 2,400" is 7 and stays at 48
 * while "Rp 15,000,000" is 13 and drops to 28.
 */
export function AmountField({ valueMinor, currency, label, helper, state = 'default', onChange }: AmountFieldProps) {
  useStyles();
  const digits = MINOR_UNITS[currency] ?? 2;
  const symbol = CURRENCY_SYMBOL[currency] ?? currency;
  const formatted = formatMinor(valueMinor, currency);
  const integerPart = digits ? formatted.slice(0, formatted.length - digits - 1) : formatted;
  return (
    <TextField
      className="mmr-amount"
      data-step={stepDownFor(`${symbol} ${integerPart}`)}
      label={label}
      helper={helper}
      state={state}
      prefix={symbol}
      inputMode="decimal"
      value={formatted}
      onChange={(e) => {
        const raw = Number(e.currentTarget.value.replace(/[^0-9]/g, ''));
        onChange?.(Number.isFinite(raw) ? raw : 0);
      }}
      aria-label={label ?? `Amount in ${currency}`}
      data-minor-digits={digits}
    />
  );
}
