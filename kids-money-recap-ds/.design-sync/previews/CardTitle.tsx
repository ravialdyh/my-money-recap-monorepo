import { Card, CardTitle, Eyebrow, HelperText } from '@kids-money-recap/ui';

const ground = { background: 'var(--kmr-color-ground-goal)', fontFamily: 'var(--kmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Default() {
  return <div style={ground}><Card><CardTitle>Saving target</CardTitle><HelperText>Target met. £160 to spare.</HelperText></Card></div>;
}
export function WithEyebrow() {
  return <div style={ground}><Card><div><Eyebrow>Where £1,760 went</Eyebrow><CardTitle>Mostly snacks</CardTitle></div></Card></div>;
}
