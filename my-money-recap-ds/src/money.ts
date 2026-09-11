/**
 * Money, the standard way (PRD §7, §11, D71): an amount is an exact decimal plus an ISO 4217 code, and
 * printing it is the platform formatter's job (`Intl.NumberFormat` here, `Decimal.FormatStyle.Currency`
 * on iOS — both read the same CLDR data). Nothing in this file, or anywhere else in the system, knows how
 * many decimals a currency has, what its symbol is, or which separators it uses; those come from the
 * currency code and the viewer's locale at the moment of printing.
 *
 * - `amount` is a decimal as a string ("15000000", "12.40") or a number for values that are exact in
 *   binary floating point (whole units and cents up to 2^53 are). Strings are preferred at the storage
 *   boundary; every supported runtime's formatter accepts a decimal string without loss (ES2023).
 * - `locale` is the viewer's language tag. Omit it for the runtime default, which on a phone is the
 *   device language: an Indonesian phone prints "Rp 15.000.000", a British phone printing rupiah gets
 *   "Rp 15,000,000" — same value, that phone's conventions.
 */

/** A formatter bound to one currency, for components that take one as a prop. */
export type MoneyFormat = (amount: number | string) => string;

const formatters = new Map<string, Intl.NumberFormat>();
/**
 * `display` formatters drop a zero fraction (the standard `trailingZeroDisplay: 'stripIfInteger'`): £2,400 and
 * £12.40, Rp 15.000.000, BHD 12. `entry` formatters keep the currency's full decimals, for a field being typed into.
 */
function formatter(currency: string, locale: string | undefined, kind: 'display' | 'entry'): Intl.NumberFormat {
  const key = `${kind}|${locale ?? ''}|${currency}`;
  let f = formatters.get(key);
  if (!f) {
    f = new Intl.NumberFormat(locale, {
      style: 'currency', currency, currencyDisplay: 'narrowSymbol',
      ...(kind === 'display' ? { trailingZeroDisplay: 'stripIfInteger' } : {}),
    } as Intl.NumberFormatOptions);
    formatters.set(key, f);
  }
  return f;
}

/** For display: formatMoney("15000000", "IDR", "id-ID") -> "Rp 15.000.000"; formatMoney(2400, "GBP", "en-GB") -> "£2,400"; formatMoney("12.40", "GBP", "en-GB") -> "£12.40". */
export function formatMoney(amount: number | string, currency: string, locale?: string): string {
  return formatter(currency, locale, 'display').format(amount as number);
}

/** The symbol the formatter would print for this currency in this locale ("£", "Rp", "BHD"). */
export function currencySymbol(currency: string, locale?: string): string {
  const part = formatter(currency, locale, 'display').formatToParts(1).find((p) => p.type === 'currency');
  return part ? part.value : currency;
}

/** For an entry field that draws the symbol separately, with the currency's full decimals: formatAmountNumber("2400", "GBP", "en-GB") -> "2,400.00". */
export function formatAmountNumber(amount: number | string, currency: string, locale?: string): string {
  return formatter(currency, locale, 'entry')
    .formatToParts(amount as number)
    .filter((p) => p.type !== 'currency' && p.type !== 'literal')
    .map((p) => p.value)
    .join('');
}

/** A `MoneyFormat` bound to one currency and locale. */
export function moneyFormat(currency: string, locale?: string): MoneyFormat {
  return (amount) => formatMoney(amount, currency, locale);
}

/**
 * Turns what a Member typed into a decimal string, using the locale's own separators to tell grouping from
 * the decimal point: parseAmount("2.400,50", "id-ID") -> "2400.50"; parseAmount("2,400.50", "en-GB") -> "2400.50".
 * Returns "" for input with no digits. The 11-digit cap (D57) is the caller's rule.
 */
export function parseAmount(typed: string, locale?: string): string {
  const parts = new Intl.NumberFormat(locale).formatToParts(1234.5);
  const group = parts.find((p) => p.type === 'group')?.value ?? ',';
  const decimal = parts.find((p) => p.type === 'decimal')?.value ?? '.';
  const cleaned = typed.split(group).join('').replace(decimal, '.').replace(/[^0-9.]/g, '');
  if (!/\d/.test(cleaned)) return '';
  const [whole, ...rest] = cleaned.split('.');
  return rest.length ? `${whole || '0'}.${rest.join('')}` : whole;
}
