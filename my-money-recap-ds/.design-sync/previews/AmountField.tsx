import { AmountField, Card } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-recap)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Pounds() { return <div style={ground}><Card><AmountField label="Your monthly salary" currency="GBP" locale="en-GB" value="2400" /></Card></div>; }
export function Rupiah() { return <div style={ground}><Card><AmountField label="Your monthly salary" currency="IDR" locale="id-ID" value="3000000" helper="Separators from the locale, decimals from the currency data." /></Card></div>; }
export function Dinar() { return <div style={ground}><Card><AmountField label="Your monthly salary" currency="BHD" locale="en-GB" value="1200" helper="Bahraini dinar has three decimals." /></Card></div>; }
export function Empty() { return <div style={ground}><Card><AmountField label="Amount" currency="GBP" locale="en-GB" value="" helper="Enter an amount above zero." /></Card></div>; }
export function Invalid() { return <div style={ground}><Card><AmountField label="Amount" currency="GBP" locale="en-GB" value="" state="invalid" helper="Enter an amount above zero." /></Card></div>; }

/** PRD §11, D71: 48px, then 36 and 28 when the symbol plus value does not fit the field, measured. The 11-digit cap (D57) lands at the smallest step. */
export function RupiahSalary() { return <div style={ground}><Card><AmountField label="Your monthly salary" currency="IDR" locale="id-ID" value="150000000" /></Card></div>; }
/** Shown at the phone's 390pt width: at 28px the capped value fits the real field (241 of 266px); the 342px preview card would clip it. */
export function ElevenDigitCap() { return <div style={{ ...ground, maxWidth: 390 }}><Card><AmountField label="Amount" currency="GBP" locale="en-GB" value="99999999999" state="invalid" helper="That's more than the app can store." /></Card></div>; }
