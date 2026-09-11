import { AmountField, Card } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-recap)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Pounds() { return <div style={ground}><Card><AmountField label="Your monthly salary" currency="GBP" valueMinor={240000} /></Card></div>; }
export function Rupiah() { return <div style={ground}><Card><AmountField label="Your monthly salary" currency="IDR" valueMinor={3000000} helper="Rupiah has no minor units." /></Card></div>; }
export function Dinar() { return <div style={ground}><Card><AmountField label="Your monthly salary" currency="BHD" valueMinor={1200000} helper="Bahraini dinar has three decimals." /></Card></div>; }
export function Empty() { return <div style={ground}><Card><AmountField label="Amount" currency="GBP" valueMinor={0} helper="Enter an amount above zero." /></Card></div>; }
export function Invalid() { return <div style={ground}><Card><AmountField label="Amount" currency="GBP" valueMinor={0} state="invalid" helper="Enter an amount above zero." /></Card></div>; }

/** PRD §11, D57: 48px up to 7 glyphs, 36px to 10, 28px beyond; the 12-digit cap sits at the smallest step. */
export function RupiahSalary() { return <div style={ground}><Card><AmountField label="Your monthly salary" currency="IDR" valueMinor={15000000} /></Card></div>; }
export function TwelveDigitCap() { return <div style={ground}><Card><AmountField label="Amount" currency="GBP" valueMinor={99999999999900} state="invalid" helper="That's more than the app can store." /></Card></div>; }
