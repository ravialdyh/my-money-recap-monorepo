import { Card, CardTitle, CategoryBar, Eyebrow } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-goal)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

/** All eight Categories, sorted descending by share of Spent. */
export function AllCategories() {
  return (
    <div style={ground}>
      <Card style={{ gap: 14 }}>
        <div><Eyebrow>Where £1,760 went</Eyebrow><CardTitle>Mostly bills</CardTitle></div>
        <div style={{ display: 'grid', gap: 10 }}>
          <CategoryBar label="Bills" percent={35} tone="ink" />
          <CategoryBar label="Rides" percent={26} tone="sky" />
          <CategoryBar label="Snacks" percent={14} tone="coral" />
          <CategoryBar label="Clothes" percent={10} tone="green" />
          <CategoryBar label="Gifts" percent={7} tone="gold" />
          <CategoryBar label="Games" percent={5} tone="violet" />
          <CategoryBar label="Toys" percent={2} tone="yellow" />
          <CategoryBar label="Other" percent={1} tone="rust" />
        </div>
      </Card>
    </div>
  );
}
export function SpendingSplit() {
  return (
    <div style={ground}>
      <Card style={{ gap: 14 }}>
        <div><Eyebrow>Where £1,760 went</Eyebrow><CardTitle>Mostly snacks</CardTitle></div>
        <div style={{ display: 'grid', gap: 10 }}>
          <CategoryBar label="Snacks" percent={45} tone="coral" />
          <CategoryBar label="Games" percent={36} tone="violet" />
          <CategoryBar label="Toys" percent={19} tone="yellow" />
        </div>
      </Card>
    </div>
  );
}
export function Single() {
  return <div style={ground}><Card><CategoryBar label="Saved" percent={62} tone="green" /></Card></div>;
}
