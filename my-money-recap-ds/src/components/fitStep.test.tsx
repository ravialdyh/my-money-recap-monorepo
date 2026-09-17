import { cleanup, render } from '@testing-library/react';
import { act, useRef } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useFitStep } from './fitStep';

afterEach(() => {
  cleanup();
  Reflect.deleteProperty(document, 'fonts');
});

type Widths = Record<string, number>;

/**
 * Gives an element a fixed clientWidth and a scrollWidth that varies with its own `data-step`, which is
 * how the real stylesheet behaves: the step picks a font size, and the font size decides the text width.
 */
function measurable(el: HTMLElement, available: number, widths: Widths) {
  Object.defineProperty(el, 'clientWidth', { configurable: true, get: () => available });
  Object.defineProperty(el, 'scrollWidth', {
    configurable: true,
    get: () => widths[el.dataset.step ?? 'l'] ?? 0,
  });
}

interface FakeObserver {
  fire: () => void;
  observed: Element[];
  disconnected: boolean;
}

/** Installs a ResizeObserver jsdom does not provide, and hands back the instances the hook creates. */
function stubResizeObserver(): FakeObserver[] {
  const made: FakeObserver[] = [];
  vi.stubGlobal(
    'ResizeObserver',
    class {
      observed: Element[] = [];
      disconnected = false;
      constructor(callback: () => void) {
        const self = this;
        made.push({
          fire: () => act(() => callback()),
          get observed() {
            return self.observed;
          },
          get disconnected() {
            return self.disconnected;
          },
        } as FakeObserver);
      }
      observe(el: Element) {
        this.observed.push(el);
      }
      unobserve() {}
      disconnect() {
        this.disconnected = true;
      }
    },
  );
  return made;
}

function Probe({ available, widths, padding }: { available: number; widths: Widths; padding?: number }) {
  const text = useRef<HTMLSpanElement>(null);
  const row = useRef<HTMLDivElement>(null);
  const step = useFitStep(text, padding === undefined ? null : row, 'key');
  return (
    <div
      ref={(el) => {
        row.current = el;
        if (el && padding !== undefined) {
          el.style.paddingLeft = `${padding}px`;
          el.style.paddingRight = `${padding}px`;
          Object.defineProperty(el, 'clientWidth', { configurable: true, get: () => available });
        }
      }}
    >
      <span
        data-testid="text"
        ref={(el) => {
          text.current = el;
          if (el) measurable(el, available, widths);
        }}
      >
        {step}
      </span>
    </div>
  );
}

describe('useFitStep (PRD §11, D71)', () => {
  it('keeps the largest step when the text already fits', () => {
    render(<Probe available={200} widths={{ l: 100, m: 60, s: 40 }} />);
    expect(document.querySelector('[data-testid="text"]')?.textContent).toBe('l');
  });

  it('steps down to the first size that fits', () => {
    render(<Probe available={100} widths={{ l: 300, m: 90, s: 40 }} />);
    expect(document.querySelector('[data-testid="text"]')?.textContent).toBe('m');
  });

  it('settles on the smallest step when nothing fits', () => {
    render(<Probe available={10} widths={{ l: 300, m: 200, s: 100 }} />);
    expect(document.querySelector('[data-testid="text"]')?.textContent).toBe('s');
  });

  it('leaves the chosen step on the element for the stylesheet to read', () => {
    render(<Probe available={100} widths={{ l: 300, m: 90, s: 40 }} />);
    expect(document.querySelector('[data-testid="text"]')?.getAttribute('data-step')).toBe('m');
  });

  it('allows a half pixel, so a value that lands exactly on the edge still fits', () => {
    render(<Probe available={100} widths={{ l: 100.5, m: 60, s: 40 }} />);
    expect(document.querySelector('[data-testid="text"]')?.textContent).toBe('l');
  });

  it('measures against the box inside its padding when a box is given', () => {
    render(<Probe available={140} widths={{ l: 130, m: 90, s: 40 }} padding={20} />);
    expect(document.querySelector('[data-testid="text"]')?.textContent).toBe('m');
  });

  it('re-measures when the observer reports a resize', () => {
    const observers = stubResizeObserver();
    let available = 200;
    const widths = { l: 100, m: 60, s: 40 };
    function Resizing() {
      const text = useRef<HTMLSpanElement>(null);
      const step = useFitStep(text, null, 'key');
      return (
        <span
          data-testid="text"
          ref={(el) => {
            text.current = el;
            if (el) {
              Object.defineProperty(el, 'clientWidth', { configurable: true, get: () => available });
              Object.defineProperty(el, 'scrollWidth', {
                configurable: true,
                get: () => widths[(el.dataset.step ?? 'l') as keyof typeof widths],
              });
            }
          }}
        >
          {step}
        </span>
      );
    }
    render(<Resizing />);
    expect(document.querySelector('[data-testid="text"]')?.textContent).toBe('l');
    available = 70;
    observers[0].fire();
    expect(document.querySelector('[data-testid="text"]')?.textContent).toBe('m');
  });

  it('watches the text element, and the box too when they differ', () => {
    const observers = stubResizeObserver();
    render(<Probe available={140} widths={{ l: 130, m: 90, s: 40 }} padding={20} />);
    expect(observers[0].observed).toHaveLength(2);
  });

  it('disconnects the observer when the component goes away', () => {
    const observers = stubResizeObserver();
    const view = render(<Probe available={200} widths={{ l: 100, m: 60, s: 40 }} />);
    view.unmount();
    expect(observers[0].disconnected).toBe(true);
  });

  it('survives an environment without ResizeObserver, which is the server', () => {
    expect(() => render(<Probe available={200} widths={{ l: 100, m: 60, s: 40 }} />)).not.toThrow();
  });

  it('re-measures once the webfonts have loaded, because they change the width', async () => {
    let available = 200;
    let resolveFonts = () => {};
    const ready = new Promise<void>((resolve) => {
      resolveFonts = resolve;
    });
    Object.defineProperty(document, 'fonts', { configurable: true, value: { ready } });
    const widths = { l: 100, m: 60, s: 40 };
    function Fonted() {
      const text = useRef<HTMLSpanElement>(null);
      const step = useFitStep(text, null, 'key');
      return (
        <span
          data-testid="fonted"
          ref={(el) => {
            text.current = el;
            if (el) {
              Object.defineProperty(el, 'clientWidth', { configurable: true, get: () => available });
              Object.defineProperty(el, 'scrollWidth', {
                configurable: true,
                get: () => widths[(el.dataset.step ?? 'l') as keyof typeof widths],
              });
            }
          }}
        >
          {step}
        </span>
      );
    }
    render(<Fonted />);
    expect(document.querySelector('[data-testid="fonted"]')?.textContent).toBe('l');
    available = 70;
    await act(async () => {
      resolveFonts();
      await ready;
    });
    expect(document.querySelector('[data-testid="fonted"]')?.textContent).toBe('m');
  });
});
