import { Icon, IconWell } from '@my-money-recap/ui';

const row = { display: 'flex', gap: 12, alignItems: 'center', padding: 8 };

export function Violet() { return <div style={row}><IconWell tone="violet"><Icon name="paw" /></IconWell></div>; }
export function Sky() { return <div style={row}><IconWell tone="sky"><Icon name="droplet" /></IconWell></div>; }
export function Coral() { return <div style={row}><IconWell tone="coral"><Icon name="lemon" /></IconWell></div>; }
export function AllTones() {
  return (
    <div style={row}>
      <IconWell tone="violet"><Icon name="paw" /></IconWell>
      <IconWell tone="sky"><Icon name="droplet" /></IconWell>
      <IconWell tone="coral"><Icon name="lemon" /></IconWell>
    </div>
  );
}
