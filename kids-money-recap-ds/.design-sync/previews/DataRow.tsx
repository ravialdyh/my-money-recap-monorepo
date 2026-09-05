import { Card, DataRow, DataRowList, Eyebrow, Icon } from '@kids-money-recap/ui';

const ground = { background: 'var(--kmr-color-ground-earning)', fontFamily: 'var(--kmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Default() {
  return <div style={ground}><Card><DataRow icon={<Icon name="paw" />} iconTone="violet" label="8 dog walks" amount="£16" /></Card></div>;
}
export function Pressed() {
  return <div style={ground}><Card><DataRow icon={<Icon name="droplet" />} iconTone="sky" label="1 car wash" amount="£5" state="pressed" /></Card></div>;
}
export function Disabled() {
  return <div style={ground}><Card><DataRow icon={<Icon name="lemon" />} iconTone="coral" label="Lemonade stand" amount="£3" state="disabled" /></Card></div>;
}
export function List() {
  return (
    <div style={ground}>
      <Card style={{ gap: 12 }}>
        <Eyebrow>How you earned it</Eyebrow>
        <DataRowList>
          <DataRow icon={<Icon name="paw" />} iconTone="violet" label="8 dog walks" amount="£16" />
          <DataRow icon={<Icon name="droplet" />} iconTone="sky" label="1 car wash" amount="£5" />
          <DataRow icon={<Icon name="lemon" />} iconTone="coral" label="Lemonade stand" amount="£3" />
        </DataRowList>
      </Card>
    </div>
  );
}
