import { Headline, SubCopy } from '@kids-money-recap/ui';

const ground = (g: 'goal' | 'personality') => ({
  background: `var(--kmr-color-ground-${g})`, color: 'var(--kmr-color-text-on-ground)', fontFamily: 'var(--kmr-font-family-text)',
  padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16,
});

export function UnderHeadline() {
  return (
    <div style={ground('goal')}>
      <Headline size="m">68% of the way to the skateboard</Headline>
      <SubCopy style={{ marginTop: 10 }}>About 3 weeks to go at this rate.</SubCopy>
    </div>
  );
}
export function Blurb() {
  return (
    <div style={ground('personality')}>
      <Headline size="display">The Squirrel</Headline>
      <SubCopy style={{ marginTop: 12, opacity: 0.9 }}>You stashed more than you spent, and you didn't crack once.</SubCopy>
    </div>
  );
}
