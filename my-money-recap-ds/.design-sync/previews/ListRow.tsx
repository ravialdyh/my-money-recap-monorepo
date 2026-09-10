import { Card, DataRowList, ListRow } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-recap)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Default() { return <div style={ground}><Card><ListRow label="Salary and payday" value="£2,400 · 25th" /></Card></div>; }
export function Settings() {
  return (
    <div style={ground}><Card><DataRowList>
      <ListRow label="Profile" value="Maya" />
      <ListRow label="Salary and payday" value="£2,400 · 25th" />
      <ListRow label="Saving target" value="20%" />
      <ListRow label="Currency" value="GBP" />
    </DataRowList></Card></div>
  );
}
export function CurrencyList() {
  return (
    <div style={ground}><Card><DataRowList>
      <ListRow label="GBP · British Pound" trailing="check" selected />
      <ListRow label="IDR · Indonesian Rupiah" trailing="check" />
      <ListRow label="USD · US Dollar" trailing="check" />
    </DataRowList></Card></div>
  );
}
export function Destructive() { return <div style={ground}><Card><ListRow label="Delete account" tone="destructive" /></Card></div>; }
/** Same wash as `DataRow`, shown among its siblings — a lone pressed row reads as a nested panel. */
export function Pressed() {
  return (
    <div style={ground}><Card><DataRowList>
      <ListRow label="Profile" value="Maya" />
      <ListRow label="Saving target" value="20%" state="pressed" />
      <ListRow label="Currency" value="GBP" />
    </DataRowList></Card></div>
  );
}
