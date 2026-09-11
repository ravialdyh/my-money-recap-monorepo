import { Card, StatHero } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-recap)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Hero() { return <div style={ground}><Card><StatHero eyebrow="Money that came your way" value={36} /></Card></div>; }
export function CountUp() { return <div style={ground}><Card><StatHero eyebrow="Money that came your way" value={36} animate /></Card></div>; }
export function CustomFormat() { return <div style={ground}><Card><StatHero eyebrow="Longest no-spend streak" value={9} format={(n) => `${Math.round(n)} days`} /></Card></div>; }

const idr = (n: number) => `Rp ${Math.round(n).toLocaleString('id-ID')}`;
/** PRD §11: 80px up to 7 glyphs, 56px to 10, 40px beyond — counted on the formatted string. */
export function StepDown80() { return <div style={ground}><Card><StatHero eyebrow="Up to 7 glyphs · 80px" value={0} format={idr} /></Card></div>; }
export function StepDown56() { return <div style={ground}><Card><StatHero eyebrow="8 to 10 glyphs · 56px" value={520000} format={idr} /></Card></div>; }
export function StepDown40() { return <div style={ground}><Card><StatHero eyebrow="11 glyphs and up · 40px" value={15000000} format={idr} /></Card></div>; }
