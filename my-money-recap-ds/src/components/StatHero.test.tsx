import { cleanup, render } from '@testing-library/react';
import { act } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { StatHero } from './StatHero';

afterEach(cleanup);

interface Frames {
  step: (now: number) => void;
  pending: () => number;
  cancelled: () => number[];
}

/** Replaces the frame loop with one this test drives by hand, so a count-up is checked at exact times. */
function controllableFrames(): Frames {
  let queue: Array<(now: number) => void> = [];
  const cancelled: number[] = [];
  let id = 0;
  vi.stubGlobal('requestAnimationFrame', (cb: (now: number) => void) => {
    queue.push(cb);
    id += 1;
    return id;
  });
  vi.stubGlobal('cancelAnimationFrame', (handle: number) => {
    cancelled.push(handle);
  });
  return {
    step: (now) => {
      const due = queue;
      queue = [];
      act(() => {
        for (const cb of due) cb(now);
      });
    },
    pending: () => queue.length,
    cancelled: () => cancelled,
  };
}

function shown() {
  return document.querySelector('.mmr-stat-hero-value')?.textContent;
}

describe('StatHero (PRD §11, D71)', () => {
  it('draws the formatted money for its currency and locale', () => {
    render(<StatHero eyebrow="You earned" value={2400} currency="GBP" locale="en-GB" />);
    expect(shown()).toBe('£2,400');
  });

  it('draws the eyebrow above the number', () => {
    render(<StatHero eyebrow="You earned" value={2400} currency="GBP" locale="en-GB" />);
    expect(document.querySelector('.mmr-stat-hero')?.textContent).toContain('You earned');
  });

  it('uses a custom formatter for values that are not money', () => {
    render(<StatHero eyebrow="Streak" value={9} format={(n) => `${n} days`} />);
    expect(shown()).toBe('9 days');
  });

  it('shows the final value at once when it is not asked to animate', () => {
    const frames = controllableFrames();
    render(<StatHero eyebrow="You earned" value={2400} currency="GBP" locale="en-GB" />);
    expect(shown()).toBe('£2,400');
    expect(frames.pending()).toBe(0);
  });

  it('counts up from zero and lands exactly on the final value', () => {
    const frames = controllableFrames();
    render(<StatHero eyebrow="You earned" value={2400} currency="GBP" locale="en-GB" animate />);
    frames.step(0);
    expect(shown()).toBe('£0');
    frames.step(550);
    const midway = shown();
    expect(midway).not.toBe('£0');
    expect(midway).not.toBe('£2,400');
    frames.step(950);
    expect(shown()).toBe('£2,400');
  });

  it('stops asking for frames once it has arrived', () => {
    const frames = controllableFrames();
    render(<StatHero eyebrow="You earned" value={2400} currency="GBP" locale="en-GB" animate />);
    frames.step(0);
    expect(frames.pending()).toBe(1);
    frames.step(950);
    expect(frames.pending()).toBe(0);
  });

  it('holds at zero through the delay before the count-up starts', () => {
    const frames = controllableFrames();
    render(<StatHero eyebrow="You earned" value={2400} currency="GBP" locale="en-GB" animate />);
    frames.step(0);
    frames.step(100);
    expect(shown()).toBe('£0');
  });

  it('skips the animation for a viewer who asked for less motion', () => {
    const frames = controllableFrames();
    vi.stubGlobal('matchMedia', (query: string) => ({ matches: query.includes('reduce') }));
    render(<StatHero eyebrow="You earned" value={2400} currency="GBP" locale="en-GB" animate />);
    expect(shown()).toBe('£2,400');
    expect(frames.pending()).toBe(0);
  });

  it('cancels the pending frame when it unmounts mid-count', () => {
    const frames = controllableFrames();
    const view = render(<StatHero eyebrow="You earned" value={2400} currency="GBP" locale="en-GB" animate />);
    frames.step(0);
    view.unmount();
    expect(frames.cancelled()).not.toHaveLength(0);
  });

  it('counts to the new target when the value changes', () => {
    const frames = controllableFrames();
    const view = render(<StatHero eyebrow="You earned" value={2400} currency="GBP" locale="en-GB" animate />);
    frames.step(0);
    frames.step(950);
    expect(shown()).toBe('£2,400');
    view.rerender(<StatHero eyebrow="You earned" value={3600} currency="GBP" locale="en-GB" animate />);
    frames.step(1000);
    frames.step(2000);
    expect(shown()).toBe('£3,600');
  });
});
