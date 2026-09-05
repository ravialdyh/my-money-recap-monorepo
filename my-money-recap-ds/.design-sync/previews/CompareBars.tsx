import { Card, CompareBars } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-personality)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function SavedVsSpent() {
  return (
    <div style={ground}>
      <Card>
        <CompareBars rows={[{ label: 'Saved', value: 640, display: '£640', tone: 'green' }, { label: 'Spent', value: 1760, display: '£1,760', tone: 'coral' }]} />
      </Card>
    </div>
  );
}
export function ThreeWay() {
  return (
    <div style={ground}>
      <Card>
        <CompareBars rows={[
          { label: 'Money in', value: 2400, display: '£2,400', tone: 'yellow' },
          { label: 'Saved', value: 640, display: '£640', tone: 'green' },
          { label: 'Spent', value: 1760, display: '£1,760', tone: 'coral' },
        ]} />
      </Card>
    </div>
  );
}
