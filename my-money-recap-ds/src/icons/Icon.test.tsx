import { cleanup, render } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { ICON_NAMES, Icon } from './Icon';

afterEach(cleanup);

const STROKED = ['share', 'settings', 'chevron', 'plus'];

function svg() {
  return document.querySelector('svg');
}

describe('Icon', () => {
  it('draws every glyph in the set', () => {
    expect(ICON_NAMES.length).toBeGreaterThan(0);
    for (const name of ICON_NAMES) {
      const view = render(<Icon name={name} />);
      expect(svg()).toBeTruthy();
      view.unmount();
    }
  });

  it('gives a filled glyph the sixteen pixel default', () => {
    render(<Icon name="snacks" />);
    expect(svg()?.getAttribute('width')).toBe('16');
    expect(svg()?.getAttribute('height')).toBe('16');
  });

  it('gives a stroked glyph the twenty pixel default', () => {
    for (const name of STROKED) {
      const view = render(<Icon name={name as (typeof ICON_NAMES)[number]} />);
      expect(svg()?.getAttribute('width')).toBe('20');
      view.unmount();
    }
  });

  it('takes an explicit size over either default', () => {
    render(<Icon name="snacks" size={32} />);
    expect(svg()?.getAttribute('width')).toBe('32');
    expect(svg()?.getAttribute('height')).toBe('32');
  });

  it('leaves the saving-target marker at its own dimensions', () => {
    render(<Icon name="piggy" size={99} />);
    expect(svg()?.getAttribute('width')).not.toBe('99');
  });

  it('passes styling props through to the svg', () => {
    render(<Icon name="snacks" className="tinted" aria-hidden="true" />);
    expect(svg()?.getAttribute('class')).toBe('tinted');
    expect(svg()?.getAttribute('aria-hidden')).toBe('true');
  });

  it('takes its colour from the text around it', () => {
    render(<Icon name="chevron" />);
    const markup = svg()?.outerHTML ?? '';
    expect(markup).toContain('currentColor');
  });
});
