import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { RecapCoverScreen } from './RecapCoverScreen';
import { RecapEarningScreen } from './RecapEarningScreen';
import { RecapGoalScreen } from './RecapGoalScreen';
import { RecapPersonalityScreen } from './RecapPersonalityScreen';
import { ARCHETYPE_BLURB } from './shared';

afterEach(cleanup);

/** The money formatter joins a symbol to its number with U+00A0; flatten it so assertions stay readable. */
function text() {
  return (document.body.textContent ?? '').replace(/ /g, ' ');
}

function bars() {
  return document.querySelectorAll('.mmr-category');
}

describe('RecapCoverScreen (PRD §8, D2)', () => {
  it('draws the sample month without being handed any props', () => {
    render(<RecapCoverScreen />);
    expect(text()).toContain('September 2026');
    expect(text()).toContain("Maya's recap");
  });

  it('leads with Money In as the one big number', () => {
    render(<RecapCoverScreen />);
    expect(document.querySelector('.mmr-stat-hero-value')?.textContent).toBe('£2,400');
  });

  it('shows saved and spent beside it', () => {
    render(<RecapCoverScreen />);
    expect(text()).toContain('Saved £640');
    expect(text()).toContain('Spent £1,760');
  });

  it('formats every figure in the Member currency and the viewer locale', () => {
    render(<RecapCoverScreen currency="IDR" locale="id-ID" total={15000000} saved={5000000} spent={10000000} />);
    expect(text()).toContain('Rp 15.000.000');
    expect(text()).toContain('Saved Rp 5.000.000');
  });

  it('marks its place as the first of the four screen run', () => {
    render(<RecapCoverScreen />);
    expect(document.querySelectorAll('.mmr-pagedot')).toHaveLength(4);
    expect(screen.getByRole('img', { name: 'Page 1 of 4' })).toBeTruthy();
  });
});

describe('RecapEarningScreen (PRD §6.12, §8)', () => {
  it('states the share of the target that was kept', () => {
    render(<RecapEarningScreen />);
    expect(text()).toContain("You kept £640 of £2,400. That's 133% of your target.");
  });

  it('drops the percentage for a Member who never set a target', () => {
    render(<RecapEarningScreen targetAmount={null} />);
    expect(text()).toContain('You kept £640 of £2,400.');
    expect(text()).not.toContain('% of your target');
  });

  it('treats a zero target as no target rather than dividing by it', () => {
    render(<RecapEarningScreen targetAmount={0} />);
    expect(text()).not.toContain('% of your target');
    expect(text()).not.toContain('Infinity');
  });

  it('caps a runaway percentage rather than printing it in full', () => {
    render(<RecapEarningScreen saved={1000000} targetAmount={1} />);
    expect(text()).toContain('999%');
  });

  it('shows the streak the month earned', () => {
    render(<RecapEarningScreen />);
    expect(text()).toContain('12 to 20 September. New badge unlocked.');
  });
});

describe('RecapGoalScreen (PRD §6.13, §8.7, SCREEN-STATES C3)', () => {
  it('leads with the share of the target and how much was spare', () => {
    render(<RecapGoalScreen />);
    expect(text()).toContain('133% of your saving target');
    expect(text()).toContain('Target met. £160 to spare.');
  });

  it('swaps the headline and offers a prompt when there is no target', () => {
    render(<RecapGoalScreen target={null} />);
    expect(text()).toContain('Where £1,760 went');
    expect(text()).toContain('Set a saving target');
  });

  it('opens settings from the no-target prompt', () => {
    const onSetTarget = vi.fn();
    render(<RecapGoalScreen target={null} onSetTarget={onSetTarget} />);
    fireEvent.click(screen.getByText('Set a saving target'));
    expect(onSetTarget).toHaveBeenCalled();
  });

  it('says so plainly when the month has no spends at all', () => {
    render(<RecapGoalScreen spentTotal={0} />);
    expect(text()).toContain('Nothing spent this month');
  });

  it('names the biggest category in the card title', () => {
    render(<RecapGoalScreen />);
    expect(text()).toContain('Mostly bills');
  });

  it('draws only as many bars as the card holds', () => {
    render(<RecapGoalScreen />);
    expect(bars()).toHaveLength(3);
  });

  it('rolls the categories that did not fit into one line with their total', () => {
    render(<RecapGoalScreen />);
    expect(text()).toContain('· 25%');
  });

  it('leaves the tail line out when every category fits', () => {
    render(<RecapGoalScreen maxCategoryBars={8} />);
    expect(bars()).toHaveLength(8);
    expect(text()).not.toContain('· 25%');
  });

  it('hides the splurge line when the caller passes none', () => {
    render(<RecapGoalScreen splurge={null} />);
    expect(text()).not.toContain('biggest single buy');
  });
});

describe('RecapPersonalityScreen (PRD §8.6)', () => {
  it('gives the Archetype its verbatim blurb', () => {
    render(<RecapPersonalityScreen />);
    expect(text()).toContain(ARCHETYPE_BLURB.squirrel);
  });

  it('follows the Archetype it is given', () => {
    render(<RecapPersonalityScreen archetype="hustler" />);
    expect(text()).toContain(ARCHETYPE_BLURB.hustler);
    expect(text()).not.toContain(ARCHETYPE_BLURB.squirrel);
  });

  it('lets a caller override the blurb and the label', () => {
    render(<RecapPersonalityScreen archetypeLabel="The Magpie" blurb="You liked shiny things." />);
    expect(text()).toContain('The Magpie');
    expect(text()).toContain('You liked shiny things.');
  });

  it('addresses the Member by name', () => {
    render(<RecapPersonalityScreen childName="Ravi" />);
    expect(text()).toContain('This month, Ravi, you were');
  });

  it('compares saved against spent in the Member currency', () => {
    render(<RecapPersonalityScreen />);
    expect(text()).toContain('£640');
    expect(text()).toContain('£1,760');
  });
});
