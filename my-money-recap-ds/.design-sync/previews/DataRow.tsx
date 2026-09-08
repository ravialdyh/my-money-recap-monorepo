import { Card, DataRow, DataRowList, Eyebrow, Icon } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-earning)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Default() {
  return <div style={ground}><Card><DataRow icon={<Icon name="paw" />} iconTone="violet" label="September 2026" amount="Squirrel" /></Card></div>;
}
/**
 * Pressed is a 4% ink wash, no scale and no brightness shift — a row in a list must not shrink while
 * its neighbours hold still. Shown inside the list it lives in, because alone in a card the wash has
 * no siblings to read against and looks like a nested panel rather than a highlighted row.
 */
export function Pressed() {
  return (
    <div style={ground}>
      <Card style={{ gap: 12 }}>
        <Eyebrow>Past recaps</Eyebrow>
        <DataRowList>
          <DataRow icon={<Icon name="paw" />} iconTone="violet" label="September 2026" amount="Squirrel" />
          <DataRow icon={<Icon name="droplet" />} iconTone="sky" label="August 2026" amount="Hustler" state="pressed" />
          <DataRow icon={<Icon name="lemon" />} iconTone="coral" label="July 2026" amount="Slow Burner" />
        </DataRowList>
      </Card>
    </div>
  );
}
export function Disabled() {
  return <div style={ground}><Card><DataRow icon={<Icon name="lemon" />} iconTone="coral" label="July 2026" amount="Slow Burner" state="disabled" /></Card></div>;
}
export function List() {
  return (
    <div style={ground}>
      <Card style={{ gap: 12 }}>
        <Eyebrow>Past recaps</Eyebrow>
        <DataRowList>
          <DataRow icon={<Icon name="paw" />} iconTone="violet" label="September 2026" amount="Squirrel" />
          <DataRow icon={<Icon name="droplet" />} iconTone="sky" label="August 2026" amount="Hustler" />
          <DataRow icon={<Icon name="lemon" />} iconTone="coral" label="July 2026" amount="Slow Burner" />
        </DataRowList>
      </Card>
    </div>
  );
}
