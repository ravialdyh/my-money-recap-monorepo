import { Card, CompareBars } from '@kids-money-recap/ui';

const ground = { background: 'var(--kmr-color-ground-personality)', fontFamily: 'var(--kmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function SavedVsSpent() {
  return (
    <div style={ground}>
      <Card>
        <CompareBars rows={[{ label: 'Saved', value: 25, display: '£25', tone: 'green' }, { label: 'Spent', value: 11, display: '£11', tone: 'coral' }]} />
      </Card>
    </div>
  );
}
export function ThreeWay() {
  return (
    <div style={ground}>
      <Card>
        <CompareBars rows={[
          { label: 'Earned', value: 24, display: '£24', tone: 'yellow' },
          { label: 'Saved', value: 25, display: '£25', tone: 'green' },
          { label: 'Spent', value: 11, display: '£11', tone: 'coral' },
        ]} />
      </Card>
    </div>
  );
}
