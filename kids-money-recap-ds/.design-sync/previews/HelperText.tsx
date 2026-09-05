import { Card, CardTitle, HelperText, StatChip, StatChipRow } from '@kids-money-recap/ui';

const ground = { background: 'var(--kmr-color-ground-recap)', fontFamily: 'var(--kmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function InCard() {
  return (
    <div style={ground}>
      <Card>
        <StatChipRow><StatChip tone="earned">Money in £2,400</StatChip><StatChip tone="saved">Saved £640</StatChip></StatChipRow>
        <HelperText>Swipe to see where it all went.</HelperText>
      </Card>
    </div>
  );
}
export function GoalHelper() {
  return (
    <div style={{ ...ground, background: 'var(--kmr-color-ground-goal)' }}>
      <Card>
        <CardTitle>Saving target</CardTitle>
        <HelperText>Target met. £160 to spare.</HelperText>
      </Card>
    </div>
  );
}
