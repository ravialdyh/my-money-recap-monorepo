import { Card, CardTitle, Eyebrow, CategoryBar, StatHero, StreakBadge, Illustration, ILLUSTRATIONS } from '@my-money-recap/ui';

const ground = (g: string) => ({ background: `var(--mmr-color-ground-${g})`, fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 });

export function Surface() {
  return <div style={ground('recap')}><Card><Eyebrow>Where £1,760 went</Eyebrow><CardTitle>Mostly snacks</CardTitle></Card></div>;
}
export function Accent() {
  return (
    <div style={ground('earning')}>
      <Card variant="accent">
        <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          <StreakBadge value={9} />
          <div style={{ display: 'grid', gap: 2 }}>
            <CardTitle>Longest no-spend streak</CardTitle>
            <div style={{ fontSize: 13, fontWeight: 700, opacity: 0.8, fontFamily: 'var(--mmr-font-family-text)' }}>12 to 20 September. New badge unlocked.</div>
          </div>
        </div>
      </Card>
    </div>
  );
}
export function Split() {
  return (
    <div style={ground('goal')}>
      <Card variant="split" style={{ marginTop: 40 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14, flexGrow: 1, minWidth: 0 }}>
          <div><Eyebrow>Where £1,760 went</Eyebrow><CardTitle>Mostly snacks</CardTitle></div>
          <div style={{ display: 'grid', gap: 10 }}>
            <CategoryBar label="Snacks" percent={45} tone="coral" />
            <CategoryBar label="Games" percent={36} tone="violet" />
            <CategoryBar label="Toys" percent={19} tone="yellow" />
          </div>
        </div>
        <div style={{ width: 120, flexShrink: 0, display: 'flex', alignItems: 'flex-end', justifyContent: 'center', marginTop: -40 }}>
          <img src={ILLUSTRATIONS.vending} alt="" style={{ width: 120, display: 'block' }} />
        </div>
      </Card>
    </div>
  );
}
export function UnderIllustration() {
  return (
    <div style={ground('recap')}>
      <Illustration src={ILLUSTRATIONS.cover} width={220} />
      <Card underIllustration>
        <StatHero eyebrow="Money that came your way" value={36} currency="GBP" />
      </Card>
    </div>
  );
}
