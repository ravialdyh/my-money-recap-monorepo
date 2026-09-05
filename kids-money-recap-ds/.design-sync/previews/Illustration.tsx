import { Card, CardTitle, Illustration, ILLUSTRATIONS } from '@kids-money-recap/ui';

const ground = (g: string) => ({ background: `var(--kmr-color-ground-${g})`, fontFamily: 'var(--kmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16, overflow: 'hidden' });

export function Cover() {
  return <div style={ground('recap')}><Illustration src={ILLUSTRATIONS.cover} width={260} /><Card underIllustration><CardTitle>Hangs into the card</CardTitle></Card></div>;
}
export function Lemonade() {
  return <div style={ground('earning')}><Illustration src={ILLUSTRATIONS.lemonade} width={200} /><Card underIllustration><CardTitle>Past recaps</CardTitle></Card></div>;
}
export function Vending() {
  return <div style={ground('goal')}><Illustration src={ILLUSTRATIONS.vending} width={168} /><Card underIllustration><CardTitle>Where £1,760 went</CardTitle></Card></div>;
}
export function Squirrel() {
  return <div style={ground('personality')}><Illustration src={ILLUSTRATIONS.squirrel} width={280} /><Card underIllustration><CardTitle>The Squirrel</CardTitle></Card></div>;
}
