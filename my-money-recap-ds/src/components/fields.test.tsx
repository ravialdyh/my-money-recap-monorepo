import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { useRef } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { AmountField } from './AmountField';
import { TextField } from './TextField';

afterEach(cleanup);

function field() {
  return document.querySelector<HTMLElement>('.mmr-field');
}

describe('TextField', () => {
  it('draws the label, the helper and both affixes', () => {
    render(<TextField label="Salary" helper="After tax" prefix="£" suffix="/mo" />);
    expect(screen.getByText('Salary')).toBeTruthy();
    expect(screen.getByText('After tax')).toBeTruthy();
    expect(screen.getByText('£')).toBeTruthy();
    expect(screen.getByText('/mo')).toBeTruthy();
  });

  it('leaves the label and helper out when it has none', () => {
    render(<TextField />);
    expect(document.querySelector('.mmr-field-label')).toBeNull();
    expect(document.querySelector('.mmr-field-helper')).toBeNull();
  });

  it('puts the state on the field row so the ring can follow it', () => {
    render(<TextField state="invalid" helper="Too big" />);
    expect(field()?.getAttribute('data-state')).toBe('invalid');
    expect(document.querySelector('.mmr-field-helper')?.getAttribute('data-state')).toBe('invalid');
  });

  it('disables the input when the state says disabled, without being told twice', () => {
    render(<TextField state="disabled" />);
    expect((document.querySelector('.mmr-field-input') as HTMLInputElement).disabled).toBe(true);
  });

  it('hands fieldRef the padded row, which is the width a value must fit', () => {
    function Harness() {
      const row = useRef<HTMLSpanElement>(null);
      return (
        <>
          <TextField fieldRef={row} />
          <button type="button" onClick={() => row.current?.setAttribute('data-found', 'true')}>
            probe
          </button>
        </>
      );
    }
    render(<Harness />);
    fireEvent.click(screen.getByText('probe'));
    expect(field()?.getAttribute('data-found')).toBe('true');
  });

  it('puts its forwarded ref on the wrapping label, not the input', () => {
    function Harness() {
      const wrap = useRef<HTMLLabelElement>(null);
      return (
        <>
          <TextField ref={wrap} />
          <button type="button" onClick={() => wrap.current?.setAttribute('data-found', 'true')}>
            probe
          </button>
        </>
      );
    }
    render(<Harness />);
    expect(document.querySelector('label.mmr-field-wrap')).toBeTruthy();
    fireEvent.click(screen.getByText('probe'));
    expect(document.querySelector('label.mmr-field-wrap')?.getAttribute('data-found')).toBe('true');
  });

  it('places the measuring copy inside the field row', () => {
    render(<TextField mirror={<span data-testid="mirror">£ 2,400</span>} />);
    expect(field()?.querySelector('[data-testid="mirror"]')).toBeTruthy();
  });

  it('passes the rest of its props through to the input', () => {
    render(<TextField inputMode="decimal" value="12" onChange={() => {}} />);
    const input = document.querySelector('.mmr-field-input') as HTMLInputElement;
    expect(input.getAttribute('inputmode')).toBe('decimal');
    expect(input.value).toBe('12');
  });
});

describe('AmountField (PRD §7, §11, D71)', () => {
  it('shows the amount with the currency decimals and the locale separators', () => {
    render(<AmountField value="2400" currency="GBP" locale="en-GB" />);
    expect((document.querySelector('.mmr-field-input') as HTMLInputElement).value).toBe('2,400.00');
  });

  it('draws the currency symbol as a fixed affix, never part of the value', () => {
    render(<AmountField value="2400" currency="GBP" locale="en-GB" />);
    expect(document.querySelector('.mmr-field-affix')?.textContent).toBe('£');
  });

  it('stays empty rather than showing a zero before anything is typed', () => {
    render(<AmountField value="" currency="GBP" locale="en-GB" />);
    expect((document.querySelector('.mmr-field-input') as HTMLInputElement).value).toBe('');
  });

  it('reports what was typed as an exact decimal string, the form it is stored in', () => {
    const onChange = vi.fn();
    render(<AmountField value="2400" currency="GBP" locale="en-GB" onChange={onChange} />);
    fireEvent.change(document.querySelector('.mmr-field-input') as HTMLInputElement, {
      target: { value: '2,400.50' },
    });
    expect(onChange).toHaveBeenCalledWith('2400.50');
  });

  it('reads the separators from the locale, not from the currency', () => {
    const onChange = vi.fn();
    render(<AmountField value="2400" currency="GBP" locale="id-ID" onChange={onChange} />);
    fireEvent.change(document.querySelector('.mmr-field-input') as HTMLInputElement, {
      target: { value: '2.400,50' },
    });
    expect(onChange).toHaveBeenCalledWith('2400.50');
  });

  it('keeps a hidden copy of the symbol and value for the fit measurement', () => {
    render(<AmountField value="2400" currency="GBP" locale="en-GB" />);
    const mirror = document.querySelector('.mmr-amount-mirror');
    expect(mirror?.getAttribute('aria-hidden')).toBe('true');
    expect(mirror?.textContent).toBe('£ 2,400.00');
  });

  it('names itself after the currency when no label is given', () => {
    render(<AmountField value="2400" currency="IDR" locale="id-ID" />);
    expect(screen.getByLabelText('Amount in IDR')).toBeTruthy();
  });

  it('prefers the label for its accessible name when there is one', () => {
    render(<AmountField value="2400" currency="GBP" locale="en-GB" label="Salary" />);
    expect(screen.getByLabelText('Salary')).toBeTruthy();
  });

  it('follows the currency decimals, so a three decimal currency keeps all three', () => {
    render(<AmountField value="12" currency="BHD" locale="en-GB" />);
    expect((document.querySelector('.mmr-field-input') as HTMLInputElement).value).toBe('12.000');
  });
});
