import { Card, CardTitle, Eyebrow, StatHero } from '@kids-money-recap/ui';

const ground = (g: string) => ({ background: `var(--kmr-color-ground-${g})`, fontFamily: 'var(--kmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 });

export function Secondary() {
  return <div style={ground('recap')}><Card><Eyebrow>Where £1,760 went</Eyebrow><CardTitle>Mostly snacks</CardTitle></Card></div>;
}
export function AboveStat() {
  return <div style={ground('recap')}><Card><StatHero eyebrow="Money that came your way" value={36} /></Card></div>;
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
