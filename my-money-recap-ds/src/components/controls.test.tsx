import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { DayPicker } from './DayPicker';
import { Slider } from './Slider';
import { Toggle } from './Toggle';

const PILL = 44;
const ROW = 300;
const restores: Array<() => void> = [];

afterEach(() => {
  cleanup();
  while (restores.length) restores.pop()?.();
});

/** Replaces one layout property jsdom cannot compute, and remembers how to put the original back. */
function stubLayout(target: object, name: string, get: (self: HTMLElement) => number) {
  const original = Object.getOwnPropertyDescriptor(target, name);
  Object.defineProperty(target, name, {
    configurable: true,
    get(this: HTMLElement) {
      return get(this);
    },
  });
  restores.push(() => {
    if (original) Object.defineProperty(target, name, original);
    else Reflect.deleteProperty(target, name);
  });
}

/** Lays the pills out end to end so the centring arithmetic has real numbers to work on. */
function stubPillRow() {
  stubLayout(HTMLElement.prototype, 'offsetWidth', () => PILL);
  stubLayout(HTMLElement.prototype, 'offsetLeft', (self) => {
    const siblings = self.parentElement ? Array.from(self.parentElement.children) : [];
    return siblings.indexOf(self) * PILL;
  });
  stubLayout(HTMLElement.prototype, 'clientWidth', () => ROW);
}

describe('Toggle', () => {
  it('reports the state the Member just switched to', () => {
    const onChange = vi.fn();
    render(<Toggle checked={false} onChange={onChange} aria-label="Roll over" />);
    fireEvent.click(screen.getByRole('switch'));
    expect(onChange).toHaveBeenCalledWith(true);
  });

  it('reports false when switching back off', () => {
    const onChange = vi.fn();
    render(<Toggle checked onChange={onChange} aria-label="Roll over" />);
    fireEvent.click(screen.getByRole('switch'));
    expect(onChange).toHaveBeenCalledWith(false);
  });

  it('marks itself disabled, which is what keeps the tap from ever landing', () => {
    render(<Toggle checked={false} disabled onChange={() => {}} aria-label="Roll over" />);
    const box = screen.getByRole('switch') as HTMLInputElement;
    fireEvent.click(box);
    expect(box.disabled).toBe(true);
    expect(box.checked).toBe(false);
  });

  it('carries its label and its checked state for a screen reader', () => {
    render(<Toggle checked onChange={() => {}} aria-label="Roll over" />);
    const box = screen.getByRole('switch') as HTMLInputElement;
    expect(box.checked).toBe(true);
    expect(box.getAttribute('aria-label')).toBe('Roll over');
  });

  it('leaves the checked state to the native checkbox rather than aria-checked', () => {
    render(<Toggle checked onChange={() => {}} aria-label="Roll over" />);
    expect(screen.getByRole('switch').hasAttribute('aria-checked')).toBe(false);
  });
});

describe('Slider (PRD §6.7)', () => {
  it('reports a number, not the string the input holds', () => {
    const onChange = vi.fn();
    render(<Slider value={20} onChange={onChange} aria-label="Saving target" />);
    fireEvent.change(screen.getByRole('slider'), { target: { value: '25' } });
    expect(onChange).toHaveBeenCalledWith(25);
  });

  it('moves in steps of five by default', () => {
    render(<Slider value={20} aria-label="Saving target" />);
    expect(screen.getByRole('slider').getAttribute('step')).toBe('5');
  });

  it('fills the track in proportion to the value', () => {
    render(<Slider value={40} aria-label="Saving target" />);
    expect(document.querySelector<HTMLElement>('.mmr-fill')?.style.width).toBe('40%');
  });

  it('clamps the fill when the value runs past the ends', () => {
    const { rerender } = render(<Slider value={150} aria-label="Saving target" />);
    expect(document.querySelector<HTMLElement>('.mmr-fill')?.style.width).toBe('100%');
    rerender(<Slider value={-20} aria-label="Saving target" />);
    expect(document.querySelector<HTMLElement>('.mmr-fill')?.style.width).toBe('0%');
  });

  it('places the knob where the fill ends', () => {
    render(<Slider value={40} aria-label="Saving target" />);
    expect(document.querySelector<HTMLElement>('.mmr-slider-knob')?.style.left).toBe('40%');
  });

  it('scales the fill to a custom range', () => {
    render(<Slider value={30} min={20} max={40} aria-label="Saving target" />);
    expect(document.querySelector<HTMLElement>('.mmr-fill')?.style.width).toBe('50%');
  });

  it('marks itself disabled so the whole control reads as off', () => {
    render(<Slider value={20} disabled aria-label="Saving target" />);
    expect((screen.getByRole('slider') as HTMLInputElement).disabled).toBe(true);
    expect(document.querySelector('.mmr-slider')?.getAttribute('data-disabled')).toBe('true');
  });

  it('draws the helper line when one is given', () => {
    render(<Slider value={20} helper="£480 of £2,400 each month" aria-label="Saving target" />);
    expect(screen.getByText('£480 of £2,400 each month')).toBeTruthy();
  });
});

describe('DayPicker (PRD §6.6, §8.2, D51)', () => {
  it('offers every day of the month plus the last day', () => {
    render(<DayPicker value={1} />);
    const options = screen.getAllByRole('radio');
    expect(options).toHaveLength(32);
    expect(options[30].textContent).toBe('31');
    expect(options[31].textContent).toBe('Last day');
  });

  it('marks the chosen day for a screen reader', () => {
    render(<DayPicker value={10} />);
    const checked = screen.getAllByRole('radio').filter((el) => el.getAttribute('aria-checked') === 'true');
    expect(checked).toHaveLength(1);
    expect(checked[0].textContent).toBe('10');
  });

  it('reports the day that was tapped', () => {
    const onChange = vi.fn();
    render(<DayPicker value={1} onChange={onChange} />);
    fireEvent.click(screen.getByText('10'));
    expect(onChange).toHaveBeenCalledWith(10);
  });

  it('reports the last day as a word, not a number', () => {
    const onChange = vi.fn();
    render(<DayPicker value={1} onChange={onChange} />);
    fireEvent.click(screen.getByText('Last day'));
    expect(onChange).toHaveBeenCalledWith('last');
  });

  it('groups the pills as one radio group named Payday', () => {
    render(<DayPicker value={1} />);
    expect(screen.getByRole('radiogroup').getAttribute('aria-label')).toBe('Payday');
  });

  it('goes quiet and unclickable when disabled', () => {
    const onChange = vi.fn();
    render(<DayPicker value={1} onChange={onChange} state="disabled" />);
    fireEvent.click(screen.getByText('10'));
    expect(onChange).not.toHaveBeenCalled();
  });

  it('lists the paydays that follow, which the caller works out', () => {
    render(<DayPicker value={30} nextPaydays={['30 Nov', '31 Dec', '31 Jan']} />);
    expect(screen.getByText(/30 Nov · 31 Dec · 31 Jan/)).toBeTruthy();
  });

  it('leaves the line out when there are no paydays to show', () => {
    render(<DayPicker value={30} nextPaydays={[]} />);
    expect(document.querySelector('.mmr-day-picker-helper')).toBeNull();
  });

  it('scrolls the chosen pill to the middle when it opens', () => {
    stubPillRow();
    render(<DayPicker value={10} />);
    const row = document.querySelector<HTMLElement>('.mmr-day-picker-row');
    expect(row?.scrollLeft).toBe(9 * PILL - (ROW - PILL) / 2);
  });

  it('does not chase the selection once the Member starts tapping', () => {
    stubPillRow();
    const { rerender } = render(<DayPicker value={10} />);
    const row = document.querySelector<HTMLElement>('.mmr-day-picker-row');
    const centred = row?.scrollLeft;
    rerender(<DayPicker value={28} />);
    expect(row?.scrollLeft).toBe(centred);
  });
});
