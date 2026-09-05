import { Card, CardTitle, StreakBadge } from '@kids-money-recap/ui';

const ground = { background: 'var(--kmr-color-ground-earning)', fontFamily: 'var(--kmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Days() {
  return (
    <div style={ground}>
      <Card variant="accent">
        <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          <StreakBadge value={9} />
          <CardTitle>Longest no-spend streak</CardTitle>
        </div>
      </Card>
    </div>
  );
}
export function Weeks() {
  return (
    <div style={ground}>
      <Card variant="accent">
        <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          <StreakBadge value={3} unit="weeks" />
          <CardTitle>Saving streak</CardTitle>
        </div>
      </Card>
    </div>
  );
}
