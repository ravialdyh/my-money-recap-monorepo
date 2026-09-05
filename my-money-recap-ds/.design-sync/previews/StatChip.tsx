import { Card, StatChip, StatChipRow } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-recap)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Earned() { return <div style={ground}><Card><StatChipRow><StatChip tone="earned">Money in £2,400</StatChip></StatChipRow></Card></div>; }
export function Saved() { return <div style={ground}><Card><StatChipRow><StatChip tone="saved">Saved £640</StatChip></StatChipRow></Card></div>; }
export function Spent() { return <div style={ground}><Card><StatChipRow><StatChip tone="spent">Spent £1,760</StatChip></StatChipRow></Card></div>; }
export function Pressed() { return <div style={ground}><Card><StatChipRow><StatChip tone="saved" state="pressed">Pressed</StatChip></StatChipRow></Card></div>; }
export function Disabled() { return <div style={ground}><Card><StatChipRow><StatChip tone="saved" state="disabled">Disabled</StatChip></StatChipRow></Card></div>; }
export function Row() {
  return (
    <div style={ground}>
      <Card>
        <StatChipRow>
          <StatChip tone="earned">Money in £2,400</StatChip><StatChip tone="saved">Saved £640</StatChip><StatChip tone="spent">Spent £1,760</StatChip>
        </StatChipRow>
      </Card>
    </div>
  );
}
