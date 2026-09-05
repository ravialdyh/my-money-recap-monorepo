import { AmountField, Card } from '@kids-money-recap/ui';

const ground = { background: 'var(--kmr-color-ground-recap)', fontFamily: 'var(--kmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Pounds() { return <div style={ground}><Card><AmountField label="Your monthly salary" currency="GBP" valueMinor={240000} /></Card></div>; }
export function Rupiah() { return <div style={ground}><Card><AmountField label="Your monthly salary" currency="IDR" valueMinor={3000000} helper="Rupiah has no minor units." /></Card></div>; }
export function Dinar() { return <div style={ground}><Card><AmountField label="Your monthly salary" currency="BHD" valueMinor={1200000} helper="Bahraini dinar has three decimals." /></Card></div>; }
export function Empty() { return <div style={ground}><Card><AmountField label="Amount" currency="GBP" valueMinor={0} helper="Enter an amount above zero." /></Card></div>; }
export function Invalid() { return <div style={ground}><Card><AmountField label="Amount" currency="GBP" valueMinor={0} state="invalid" helper="Enter an amount above zero." /></Card></div>; }
