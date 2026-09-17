// Render check, not a test: the fragile-fit screens and long-number states in a real browser, so the
// fit-to-width pass runs (PRD §11, D45, D71). `npm run measure`; README "Rebuilding generated files".
import { createRoot } from 'react-dom/client';
import { AmountField } from '../src/components/AmountField';
import { Banner } from '../src/components/Banner';
import { Card } from '../src/components/Card';
import { PageDots } from '../src/components/PageDots';
import { PromptCard } from '../src/components/PromptCard';
import { Screen } from '../src/components/Screen';
import { StatChip, StatChipRow } from '../src/components/StatChip';
import { StatHero } from '../src/components/StatHero';
import { RecapCoverScreen, RecapEarningScreen, RecapGoalScreen, RecapPersonalityScreen } from '../src/screens';
import { MMR_CSS } from '../src/styles/css.generated';

const cases: Array<[string, JSX.Element]> = [
  ['C3 goal · default', <RecapGoalScreen />],
  [
    'C3 goal · no spends',
    <RecapGoalScreen
      saved={2400}
      target={480}
      helper="Target met. £1,920.00 to spare."
      subCopy="Not a single spend logged."
      spentTotal={0}
      categories={[]}
      splurge={null}
    />,
  ],
  [
    'C2 earning · no target (null)',
    <RecapEarningScreen
      month="October"
      moneyIn={2400}
      saved={310}
      targetAmount={null}
      streakDays={4}
      streakRange={[8, 11]}
      daysInMonth={31}
    />,
  ],
  ['C4 squirrel', <RecapPersonalityScreen archetype="squirrel" />],
  ['C4 hustler (bee)', <RecapPersonalityScreen archetype="hustler" saved={520} spent={1880} />],
  ['C4 treat (cat)', <RecapPersonalityScreen archetype="treatYourself" saved={210} spent={2190} />],
  ['C4 slowburner (tortoise)', <RecapPersonalityScreen archetype="slowBurner" saved={180} spent={2220} />],
  [
    'C1 cover · rupiah (X4)',
    <RecapCoverScreen
      month="September 2026"
      childName="Budi"
      total={15000000}
      saved={5600000}
      spent={9400000}
      currency="IDR"
      locale="id-ID"
    />,
  ],
  ['C3 goal · no target', <RecapGoalScreen target={null} spentTotal={1760} />],
  [
    'C3 goal · rupiah',
    <RecapGoalScreen
      currency="IDR"
      locale="id-ID"
      saved={5600000}
      target={3000000}
      spentTotal={9400000}
      splurge="Rp 2.500.000 on a flight home. Your biggest single buy."
    />,
  ],
  [
    'The cap (D57) · 11 digits at 390pt',
    <Screen ground="recap">
      <Card>
        <StatHero eyebrow="11 digits, pounds" value={99999999999} currency="GBP" locale="en-GB" />
      </Card>
      <div style={{ marginTop: 12 }}>
        <Card>
          <StatHero
            eyebrow="11 digits, rupiah (297 of 298pt at 40pt)"
            value={99999999999}
            currency="IDR"
            locale="id-ID"
          />
        </Card>
      </div>
      <div style={{ marginTop: 12 }}>
        <Card>
          <AmountField label="11 digits" currency="GBP" locale="en-GB" value="99999999999" />
          <AmountField label="11 digits rupiah" currency="IDR" locale="id-ID" value="99999999999" />
        </Card>
      </div>
    </Screen>,
  ],
  [
    'Long-number pieces · rupiah',
    <Screen ground="recap">
      <Card>
        <StatHero eyebrow="Spent so far" value={9400000} currency="IDR" locale="id-ID" />
        <StatChipRow>
          <StatChip tone="earned">Money in Rp 15.000.000</StatChip>
          <StatChip tone="saved">Saved Rp 5.600.000</StatChip>
        </StatChipRow>
      </Card>
      <div style={{ marginTop: 12 }}>
        <PromptCard
          height={233}
          title="Set your monthly salary"
          helper="We'll show what you kept once we know what came in."
        />
      </div>
      <div style={{ marginTop: 12 }}>
        <Card>
          <StatHero eyebrow="Steps to 56" value={520000} currency="IDR" locale="id-ID" />
          <StatHero eyebrow="Steps to 40" value={15000000} currency="IDR" locale="id-ID" />
          <StatHero eyebrow="The cap at 40" value={99999999999} currency="GBP" locale="en-GB" />
        </Card>
      </div>
      <div style={{ marginTop: 12 }}>
        <Banner tone="offline">No connection. You can look, but not log.</Banner>
      </div>
      <div style={{ marginTop: 12 }}>
        <Card>
          <AmountField label="Salary" currency="IDR" locale="id-ID" value="15000000" />
          <AmountField
            label="Amount"
            currency="GBP"
            locale="en-GB"
            value="99999999999"
            state="invalid"
            helper="That's more than the app can store."
          />
          <AmountField label="Amount" currency="GBP" locale="en-GB" value="2400" />
        </Card>
      </div>
      <div style={{ marginTop: 12 }}>
        <PageDots count={4} active={1} onSelect={() => {}} />
      </div>
    </Screen>,
  ],
];

const style = document.createElement('style');
style.textContent = `${MMR_CSS}
body{margin:0;background:#ddd;display:flex;gap:24px;padding:24px;align-items:flex-start;width:max-content}
figure{margin:0} figcaption{font:12px sans-serif;padding:6px 0}
.wrap{position:relative} .wrap::after{content:'';position:absolute;left:0;top:844px;width:390px;border-top:2px dashed #f00;z-index:99}`;
document.head.appendChild(style);
createRoot(document.getElementById('root')!).render(
  cases.map(([name, el]) => (
    <figure key={name}>
      <figcaption>{name}</figcaption>
      <div className="wrap" data-name={name}>
        {el}
      </div>
    </figure>
  )),
);
