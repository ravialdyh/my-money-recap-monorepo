/**
 * An exact decimal plus an ISO 4217 code, printed by `Intl.NumberFormat` (PRD §7, §11, D71). Nothing
 * here knows a currency's decimals, symbol or separators. README "Design rules" has the reasoning.
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
 * Turns typed text into a decimal string using the locale's own separators: "2.400,50" in `id-ID` and
 * "2,400.50" in `en-GB` both give "2400.50". Input with no digits gives "". D57's cap is the caller's.
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
