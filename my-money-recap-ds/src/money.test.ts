import { describe, expect, it } from 'vitest';
import { currencySymbol, formatAmountNumber, formatMoney, parseAmount } from './money';

// The formatter separates a symbol from the number with a non-breaking space (U+00A0), so the two never split across lines.
const NB = ' ';

describe('money formatting (PRD §7, §11, D71)', () => {
  it('takes the symbol from the currency and locale', () => {
    expect(currencySymbol('GBP', 'en-GB')).toBe('£');
    expect(currencySymbol('IDR', 'id-ID')).toBe('Rp');
    expect(currencySymbol('IDR', 'en-GB')).toBe('Rp');
    expect(currencySymbol('BHD', 'en-GB')).toBe('BHD');
  });

  it('follows the viewer locale for separators and drops a zero fraction on display', () => {
    expect(formatMoney(2400, 'GBP', 'en-GB')).toBe('£2,400');
    expect(formatMoney(2400, 'GBP', 'id-ID')).toBe('£2.400');
    expect(formatMoney('12.40', 'GBP', 'en-GB')).toBe('£12.40');
    expect(formatMoney('12.345', 'BHD', 'en-GB')).toBe(`BHD${NB}12.345`);
    expect(formatMoney(12, 'BHD', 'en-GB')).toBe(`BHD${NB}12`);
    expect(formatMoney(2400, 'JPY', 'ja-JP')).toBe('￥2,400');
  });

  it('prints rupiah the same on every platform for whole amounts; a fraction follows the platform data (CLDR 47: 2 decimals, 48+: 0)', () => {
    expect(formatMoney('15000000', 'IDR', 'id-ID').replace(NB, ' ')).toBe('Rp 15.000.000');
    expect(['Rp 15.000.000,50', 'Rp 15.000.001']).toContain(formatMoney('15000000.5', 'IDR', 'id-ID').replace(NB, ' '));
  });

  it('accepts exact decimal strings, the storage form, without loss', () => {
    expect(formatMoney('12.40', 'GBP', 'en-GB')).toBe('£12.40');
    expect(formatMoney('99999999999', 'GBP', 'en-GB')).toBe('£99,999,999,999');
  });

  it('keeps the full decimals in an entry field that draws its own symbol', () => {
    expect(formatAmountNumber('2400', 'GBP', 'en-GB')).toBe('2,400.00');
    expect(formatAmountNumber('12', 'BHD', 'en-GB')).toBe('12.000');
  });

  it('parses typed input using the locale separators back to a decimal string', () => {
    expect(parseAmount('2,400.50', 'en-GB')).toBe('2400.50');
    expect(parseAmount('2.400,50', 'id-ID')).toBe('2400.50');
    expect(parseAmount('12', 'en-GB')).toBe('12');
    expect(parseAmount('', 'en-GB')).toBe('');
    expect(parseAmount('£', 'en-GB')).toBe('');
  });

  it('works for any ISO 4217 code without configuration', () => {
    for (const code of ['USD', 'EUR', 'SGD', 'KWD', 'VND', 'CHF', 'COP', 'HUF']) {
      expect(() => formatMoney(1, code, 'en-GB')).not.toThrow();
    }
  });
});
