import { Button, Chip, ListRow, Sheet, TextField } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-recap)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };
const chips = { display: 'flex', flexWrap: 'wrap' as const, gap: 8 };

export function LogASpend() {
  return (
    <div style={{ ...ground, padding: 0, paddingTop: 24 }}>
      <Sheet title="Log a spend">
        <TextField label="Amount" prefix="£" defaultValue="4.50" />
        <div style={chips}>
          <Chip tappable variant="inverse">Snacks</Chip><Chip tappable>Games</Chip><Chip tappable>Toys</Chip><Chip tappable>Rides</Chip>
        </div>
        <Button>Save</Button>
      </Sheet>
    </div>
  );
}
export function Untitled() {
  return <div style={{ ...ground, padding: 0, paddingTop: 24 }}><Sheet><TextField label="Note" placeholder="Optional" /><Button>Save</Button></Sheet></div>;
}
export function FullSurface() {
  return (
    <div style={{ ...ground, padding: 12 }}>
      <Sheet variant="full" title="Settings">
        <ListRow label="Profile" value="Maya" />
        <ListRow label="Currency" value="GBP" />
        <ListRow label="Log out" trailing="none" />
      </Sheet>
    </div>
  );
}
