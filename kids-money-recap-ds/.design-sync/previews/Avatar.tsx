import { Avatar, ILLUSTRATIONS } from '@kids-money-recap/ui';

const ground = { background: 'var(--kmr-color-ground-recap)', fontFamily: 'var(--kmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Greeting() {
  return (
    <div style={{ ...ground, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
      <span style={{ fontFamily: 'var(--kmr-font-family-display)', fontSize: 40, fontWeight: 700, color: '#fff' }}>Hi, Maya</span>
      <Avatar src={ILLUSTRATIONS['avatar-otter']} label="Otter" />
    </div>
  );
}
export function Small() { return <div style={ground}><Avatar src={ILLUSTRATIONS['avatar-fox']} label="Fox" size={40} /></div>; }
