import { Card, StatHero } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-recap)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Hero() { return <div style={ground}><Card><StatHero eyebrow="Money that came your way" value={36} /></Card></div>; }
export function CountUp() { return <div style={ground}><Card><StatHero eyebrow="Money that came your way" value={36} animate /></Card></div>; }
export function CustomFormat() { return <div style={ground}><Card><StatHero eyebrow="Longest no-spend streak" value={9} format={(n) => `${Math.round(n)} days`} /></Card></div>; }
