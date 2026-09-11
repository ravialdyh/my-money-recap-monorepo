import { Card, CardTitle, Eyebrow, StatHero } from '@my-money-recap/ui';

const ground = (g: string) => ({ background: `var(--mmr-color-ground-${g})`, fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 });

export function Secondary() {
  return <div style={ground('recap')}><Card><Eyebrow>Where £1,760 went</Eyebrow><CardTitle>Mostly snacks</CardTitle></Card></div>;
}
export function AboveStat() {
  return <div style={ground('recap')}><Card><StatHero eyebrow="Money that came your way" value={36} currency="GBP" /></Card></div>;
}
export function Coral() {
  return (
    <div style={ground('goal')}>
      <Card style={{ gap: 4 }}>
        <Eyebrow tone="coral">The splurge</Eyebrow>
        <div style={{ fontSize: 15, fontWeight: 800, lineHeight: 1.3 }}>£4 on a new game. Your biggest single buy.</div>
      </Card>
    </div>
  );
}
