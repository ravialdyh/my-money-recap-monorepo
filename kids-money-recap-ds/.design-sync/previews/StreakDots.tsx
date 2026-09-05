import { Card, CardTitle, StreakDots } from '@kids-money-recap/ui';

const ground = { background: 'var(--kmr-color-ground-earning)', fontFamily: 'var(--kmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Range() {
  return <div style={ground}><Card variant="accent"><CardTitle>12 to 20 September</CardTitle><StreakDots total={30} activeRange={[12, 20]} /></Card></div>;
}
export function ExplicitDays() {
  return <div style={ground}><Card variant="accent"><CardTitle>Seven no-spend days</CardTitle><StreakDots total={30} activeDays={[1, 2, 5, 9, 10, 11, 22]} /></Card></div>;
}
export function FullMonth() {
  return <div style={ground}><Card variant="accent"><CardTitle>Every day in October</CardTitle><StreakDots total={31} activeRange={[1, 31]} /></Card></div>;
}
