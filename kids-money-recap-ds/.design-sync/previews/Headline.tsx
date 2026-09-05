import { Headline } from '@kids-money-recap/ui';

const ground = (g: 'recap' | 'personality') => ({
  background: `var(--kmr-color-ground-${g})`, color: 'var(--kmr-color-text-on-ground)',
  fontFamily: 'var(--kmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16,
});

export function HeadlineL() { return <div style={ground('recap')}><Headline size="l">Your September in money</Headline></div>; }
export function HeadlineM() { return <div style={ground('recap')}><Headline size="m">68% of the way to the skateboard</Headline></div>; }
export function HeadlineCompact() { return <div style={ground('recap')}><Headline size="compact">You earned £24. That's 12 dog walks' worth.</Headline></div>; }
export function Display() { return <div style={ground('personality')}><Headline size="display">The Squirrel</Headline></div>; }
