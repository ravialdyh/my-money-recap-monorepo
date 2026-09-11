import { Card, CardTitle, StreakDots } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-earning)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Range() {
  return <div style={ground}><Card variant="accent"><CardTitle>12 to 20 September</CardTitle><StreakDots total={30} activeRange={[12, 20]} /></Card></div>;
}
export function FullMonth() {
  return <div style={ground}><Card variant="accent"><CardTitle>Every day in October</CardTitle><StreakDots total={31} activeRange={[1, 31]} /></Card></div>;
}
