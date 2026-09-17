import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { AvatarGrid } from './AvatarGrid';
import { Banner } from './Banner';
import { DataRow, DataRowList, IconWell } from './DataRow';
import { InlineError } from './InlineError';
import { ListRow } from './ListRow';
import { Sheet } from './Sheet';

afterEach(cleanup);

const AVATARS = [
  { key: 'fox', src: 'fox.webp' },
  { key: 'otter', src: 'otter.webp' },
] as const;

describe('Banner (PRD §6.8, D58, D63)', () => {
  it('announces itself as a status without stealing focus', () => {
    render(<Banner>You are close to your target.</Banner>);
    expect(screen.getByRole('status').textContent).toContain('You are close to your target.');
  });

  it('is a warning unless told otherwise', () => {
    render(<Banner>Careful</Banner>);
    expect(screen.getByRole('status').getAttribute('data-tone')).toBe('warning');
  });

  it('takes the tone it is given', () => {
    render(<Banner tone="offline">No connection. You can look, but not log.</Banner>);
    expect(screen.getByRole('status').getAttribute('data-tone')).toBe('offline');
  });

  it('offers no dismiss control unless there is somewhere to send it', () => {
    render(<Banner>Careful</Banner>);
    expect(screen.queryByLabelText('Dismiss')).toBeNull();
  });

  it('dismisses for the session when asked', () => {
    const onDismiss = vi.fn();
    render(<Banner onDismiss={onDismiss}>Careful</Banner>);
    fireEvent.click(screen.getByLabelText('Dismiss'));
    expect(onDismiss).toHaveBeenCalled();
  });
});

describe('InlineError (PRD §6.1, §6.19, D67)', () => {
  it('speaks up as an alert, because it replaced the toast', () => {
    render(<InlineError>We could not reach the rate.</InlineError>);
    expect(screen.getByRole('alert').textContent).toContain('We could not reach the rate.');
  });

  it('carries a trailing action when one is offered', () => {
    render(<InlineError action={<button type="button">Try again</button>}>Rate unavailable</InlineError>);
    expect(screen.getByText('Try again')).toBeTruthy();
  });

  it('leaves the action slot out by default', () => {
    render(<InlineError>Rate unavailable</InlineError>);
    expect(document.querySelector('.mmr-inline-error-action')).toBeNull();
  });

  it('hides its decorative dot from the screen reader', () => {
    render(<InlineError>Rate unavailable</InlineError>);
    expect(document.querySelector('.mmr-inline-error-dot')?.getAttribute('aria-hidden')).toBe('true');
  });
});

describe('Sheet', () => {
  it('is a dialog named by its title', () => {
    render(<Sheet title="Saving target">body</Sheet>);
    expect(screen.getByRole('dialog').getAttribute('aria-label')).toBe('Saving target');
  });

  it('shows the drag handle on a bottom sheet', () => {
    render(<Sheet title="Saving target">body</Sheet>);
    expect(document.querySelector('.mmr-sheet-handle')).toBeTruthy();
  });

  it('drops the handle on a full sheet, which does not drag', () => {
    render(
      <Sheet title="Settings" variant="full">
        body
      </Sheet>,
    );
    expect(document.querySelector('.mmr-sheet-handle')).toBeNull();
  });

  it('stands alone unless a scrim is asked for', () => {
    const { rerender } = render(<Sheet title="Saving target">body</Sheet>);
    expect(document.querySelector('.mmr-sheet-scrim')).toBeNull();
    rerender(
      <Sheet title="Saving target" scrim>
        body
      </Sheet>,
    );
    expect(document.querySelector('.mmr-sheet-scrim')?.querySelector('.mmr-sheet')).toBeTruthy();
  });
});

describe('ListRow', () => {
  it('shows its label and its current value', () => {
    render(<ListRow label="Currency" value="GBP" />);
    expect(screen.getByText('Currency')).toBeTruthy();
    expect(screen.getByText('GBP')).toBeTruthy();
  });

  it('opens a screen with a chevron by default', () => {
    render(<ListRow label="Currency" />);
    expect(document.querySelector('.mmr-list-row-chevron')).toBeTruthy();
  });

  it('marks a selection as pressed rather than drawing a chevron', () => {
    render(<ListRow label="GBP" trailing="check" selected />);
    expect(screen.getByRole('button').getAttribute('aria-pressed')).toBe('true');
    expect(document.querySelector('.mmr-list-row-check')).toBeTruthy();
  });

  it('leaves the check off an unselected row', () => {
    render(<ListRow label="IDR" trailing="check" selected={false} />);
    expect(screen.getByRole('button').getAttribute('aria-pressed')).toBe('false');
    expect(document.querySelector('.mmr-list-row-check')).toBeNull();
  });

  it('draws a plain row with no trailing affordance', () => {
    render(<ListRow label="Version" value="1.0" trailing="none" />);
    expect(document.querySelector('.mmr-list-row-chevron')).toBeNull();
    expect(screen.getByRole('button').hasAttribute('aria-pressed')).toBe(false);
  });

  it('carries a destructive tone for a row that deletes', () => {
    render(<ListRow label="Delete account" tone="destructive" />);
    expect(screen.getByRole('button').getAttribute('data-tone')).toBe('destructive');
  });

  it('disables the button when the state says disabled', () => {
    render(<ListRow label="Currency" state="disabled" />);
    expect((screen.getByRole('button') as HTMLButtonElement).disabled).toBe(true);
  });
});

describe('DataRow', () => {
  it('lines up the icon, the label and the amount', () => {
    render(<DataRow icon={<span>x</span>} iconTone="violet" label="Snacks" amount="£24" />);
    expect(screen.getByText('Snacks')).toBeTruthy();
    expect(screen.getByText('£24')).toBeTruthy();
    expect(document.querySelector('.mmr-well')?.getAttribute('data-tone')).toBe('violet');
  });

  it('reports a tap so the row can open its Spend', () => {
    const onClick = vi.fn();
    render(<DataRow icon={<span>x</span>} iconTone="violet" label="Snacks" amount="£24" onClick={onClick} />);
    fireEvent.click(screen.getByRole('button'));
    expect(onClick).toHaveBeenCalled();
  });

  it('goes quiet when disabled', () => {
    const onClick = vi.fn();
    render(
      <DataRow
        icon={<span>x</span>}
        iconTone="violet"
        label="Snacks"
        amount="£24"
        state="disabled"
        onClick={onClick}
      />,
    );
    fireEvent.click(screen.getByRole('button'));
    expect(onClick).not.toHaveBeenCalled();
  });

  it('sizes the well only when asked, leaving the token to do it otherwise', () => {
    const { rerender } = render(
      <IconWell tone="violet">
        <span>x</span>
      </IconWell>,
    );
    expect(document.querySelector<HTMLElement>('.mmr-well')?.style.width).toBe('');
    rerender(
      <IconWell tone="violet" size={44} glyphSize={20}>
        <span>x</span>
      </IconWell>,
    );
    const well = document.querySelector<HTMLElement>('.mmr-well');
    expect(well?.style.width).toBe('44px');
    expect(well?.style.getPropertyValue('--mmr-well-glyph-size')).toBe('20px');
  });

  it('wraps its rows in a list', () => {
    render(
      <DataRowList>
        <DataRow icon={<span>x</span>} iconTone="violet" label="Snacks" amount="£24" />
      </DataRowList>,
    );
    expect(document.querySelector('.mmr-row-list')?.querySelector('.mmr-row')).toBeTruthy();
  });
});

describe('AvatarGrid (PRD D36)', () => {
  it('is one radio group with a named option per avatar', () => {
    render(<AvatarGrid avatars={[...AVATARS]} />);
    expect(screen.getByRole('radiogroup').getAttribute('aria-label')).toBe('Pick your avatar');
    expect(screen.getByRole('radio', { name: 'Fox' })).toBeTruthy();
    expect(screen.getByRole('radio', { name: 'Otter' })).toBeTruthy();
  });

  it('checks nothing while the Member has not picked', () => {
    render(<AvatarGrid avatars={[...AVATARS]} />);
    expect(screen.getAllByRole('radio').filter((el) => el.getAttribute('aria-checked') === 'true')).toHaveLength(0);
  });

  it('checks exactly the chosen avatar', () => {
    render(<AvatarGrid avatars={[...AVATARS]} selected="otter" />);
    expect(screen.getByRole('radio', { name: 'Otter' }).getAttribute('aria-checked')).toBe('true');
    expect(screen.getByRole('radio', { name: 'Fox' }).getAttribute('aria-checked')).toBe('false');
  });

  it('reports the avatar that was picked', () => {
    const onSelect = vi.fn();
    render(<AvatarGrid avatars={[...AVATARS]} onSelect={onSelect} />);
    fireEvent.click(screen.getByRole('radio', { name: 'Fox' }));
    expect(onSelect).toHaveBeenCalledWith('fox');
  });

  it('hides the artwork from the screen reader, which reads the label instead', () => {
    render(<AvatarGrid avatars={[...AVATARS]} />);
    for (const img of Array.from(document.querySelectorAll('img'))) {
      expect(img.getAttribute('alt')).toBe('');
    }
  });

  it('takes a custom cell size for the smaller inline avatar', () => {
    render(<AvatarGrid avatars={[...AVATARS]} size={56} />);
    expect(document.querySelector<HTMLElement>('.mmr-avatar')?.style.width).toBe('56px');
  });
});
