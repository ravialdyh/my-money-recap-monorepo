import { Chip, ChipRow, Icon, CATEGORIES, CATEGORY_LABEL, CATEGORY_TONE } from '@kids-money-recap/ui';

const ground = { background: 'var(--kmr-color-ground-recap)', fontFamily: 'var(--kmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function OnGround() { return <div style={ground}><Chip variant="onGround">September 2026</Chip></div>; }
export function Inverse() { return <div style={ground}><Chip variant="inverse">Earning</Chip></div>; }
export function Highlight() { return <div style={ground}><Chip variant="highlight">Maya's recap</Chip></div>; }
export function Pressed() { return <div style={ground}><Chip variant="onGround" state="pressed">Pressed</Chip></div>; }
export function Disabled() { return <div style={ground}><Chip variant="onGround" state="disabled">Disabled</Chip></div>; }
export function AsLabel() { return <div style={ground}><Chip variant="onGround" asLabel>Static label</Chip></div>; }
export function Tappable() { return <div style={ground}><Chip variant="onGround" tappable>Settings</Chip></div>; }
/** The eight Spend Categories as selection chips: each on its own tint, the chosen one ink. */
export function Categories() {
  return (
    <div style={{ ...ground, display: 'flex', flexWrap: 'wrap', gap: 8 }}>
      {CATEGORIES.map((c) => (
        <Chip key={c} tone={CATEGORY_TONE[c]} selected={c === 'snacks'} tappable>
          <Icon name={c} />{CATEGORY_LABEL[c]}
        </Chip>
      ))}
    </div>
  );
}
export function CategoryUnselected() {
  return <div style={ground}><Chip tone="violet" tappable><Icon name="games" />Games</Chip></div>;
}
export function CategorySelected() {
  return <div style={ground}><Chip tone="violet" selected tappable><Icon name="games" />Games</Chip></div>;
}
export function Row() {
  return <div style={ground}><ChipRow><Chip asLabel>September 2026</Chip><Chip asLabel variant="highlight">Maya's recap</Chip></ChipRow></div>;
}
