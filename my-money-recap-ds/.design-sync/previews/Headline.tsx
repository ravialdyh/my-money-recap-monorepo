import { Headline } from '@my-money-recap/ui';

const ground = (g: 'recap' | 'personality') => ({
  background: `var(--mmr-color-ground-${g})`, color: 'var(--mmr-color-text-on-ground)',
  fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16,
});

export function HeadlineL() { return <div style={ground('recap')}><Headline size="l">Your September in money</Headline></div>; }
export function HeadlineM() { return <div style={ground('recap')}><Headline size="m">133% of your saving target</Headline></div>; }
export function HeadlineCompact() { return <div style={ground('recap')}><Headline size="compact">You kept £640 of £2,400. That's 133% of your target.</Headline></div>; }
export function Display() { return <div style={ground('personality')}><Headline size="display">The Squirrel</Headline></div>; }
