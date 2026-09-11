import { Card, StatHero } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-recap)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Hero() { return <div style={ground}><Card><StatHero eyebrow="Money that came your way" value={2400} currency="GBP" locale="en-GB" /></Card></div>; }
export function CountUp() { return <div style={ground}><Card><StatHero eyebrow="Money that came your way" value={2400} currency="GBP" locale="en-GB" animate /></Card></div>; }
export function CustomFormat() { return <div style={ground}><Card><StatHero eyebrow="Longest no-spend streak" value={9} format={(n) => `${Math.round(n)} days`} /></Card></div>; }
/** PRD §11, D71: the value is drawn at 80px, or 56 / 40px when the full string does not fit the card — measured, not counted. */
export function FitsAt80() { return <div style={ground}><Card><StatHero eyebrow="Fits at 80px" value={0} currency="IDR" locale="id-ID" /></Card></div>; }
export function FitsAt56() { return <div style={ground}><Card><StatHero eyebrow="Too wide at 80, fits at 56px" value={520000} currency="IDR" locale="id-ID" /></Card></div>; }
export function FitsAt40() { return <div style={ground}><Card><StatHero eyebrow="Too wide at 56, drawn at 40px" value={15000000} currency="IDR" locale="id-ID" /></Card></div>; }
