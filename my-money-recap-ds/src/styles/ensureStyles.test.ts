import { beforeEach, describe, expect, it } from 'vitest';
import { MMR_CSS } from './css.generated';
import { ensureStyles } from './ensureStyles';

const STYLE_ID = 'mmr-design-system-styles';

beforeEach(() => {
  document.getElementById(STYLE_ID)?.remove();
});

describe('ensureStyles', () => {
  it('puts the stylesheet in the head', () => {
    ensureStyles();
    const el = document.getElementById(STYLE_ID);
    expect(el?.tagName).toBe('STYLE');
    expect(el?.parentElement).toBe(document.head);
  });

  it('carries the generated stylesheet, not a placeholder', () => {
    ensureStyles();
    expect(document.getElementById(STYLE_ID)?.textContent).toBe(MMR_CSS);
  });

  it('injects once however many components ask for it', () => {
    ensureStyles();
    ensureStyles();
    ensureStyles();
    expect(document.querySelectorAll(`#${STYLE_ID}`)).toHaveLength(1);
  });

  it('leaves an existing sheet alone rather than replacing it', () => {
    ensureStyles();
    const first = document.getElementById(STYLE_ID);
    ensureStyles();
    expect(document.getElementById(STYLE_ID)).toBe(first);
  });
});
