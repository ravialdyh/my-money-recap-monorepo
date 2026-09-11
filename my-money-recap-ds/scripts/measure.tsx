// Render check, not part of the package: the screens whose fit is fragile at 844pt, plus the long-number
// states (PRD §11, D45, D71). Build with esbuild (see `npm run measure`), open scripts/preview/measure.html;
// the red dashed line is 844pt.
import { renderToStaticMarkup } from 'react-dom/server';
import { writeFileSync } from 'node:fs';
import { MMR_CSS } from '../src/styles/css.generated';
import { RecapCoverScreen, RecapGoalScreen, RecapEarningScreen, RecapPersonalityScreen } from '../src/screens/RecapScreens';
import { Card } from '../src/components/Card';
import { StatHero } from '../src/components/StatHero';
import { StatChip, StatChipRow } from '../src/components/StatChip';
import { PromptCard } from '../src/components/PromptCard';
import { AmountField } from '../src/components/AmountField';
import { Banner } from '../src/components/Banner';
import { PageDots } from '../src/components/PageDots';
import { Screen } from '../src/components/Screen';

const idr = (n: number) => `Rp ${Math.round(n).toLocaleString('id-ID')}`;

const cases: Array<[string, JSX.Element]> = [
  ['C3 goal · 5 bars + splurge', <RecapGoalScreen />],
  ['C3 goal · no spends', <RecapGoalScreen saved={2400} target={480} helper="Target met. £1,920 to spare." subCopy="Not a single spend logged." spentTotal={0} categories={[]} splurge={null} />],
  ['C2 earning · no target (null)', <RecapEarningScreen month="October" moneyIn={2400} saved={310} targetAmount={null} streakDays={4} streakRange={[8,11]} daysInMonth={31} />],
  ['C4 squirrel', <RecapPersonalityScreen archetype="squirrel" />],
  ['C4 hustler (bee)', <RecapPersonalityScreen archetype="hustler" saved={520} spent={1880} />],
  ['C4 treat (cat)', <RecapPersonalityScreen archetype="treatYourself" saved={210} spent={2190} />],
  ['C4 slowburner (tortoise)', <RecapPersonalityScreen archetype="slowBurner" saved={180} spent={2220} />],
  ['C1 cover · rupiah (X4)', <RecapCoverScreen month="September 2026" childName="Budi" total={15000000} saved={5600000} spent={9400000} format={idr} />],
  ['C3 goal · no target', <RecapGoalScreen target={null} spentTotal={1760} />],
  ['C3 goal · rupiah', <RecapGoalScreen format={idr} saved={5600000} target={3000000} helper="Target met. Rp 2.600.000 to spare." spentTotal={9400000} splurge="Rp 2.500.000 on a flight home. Your biggest single buy." />],
  ['Long-number pieces · rupiah', <Screen ground="recap">
    <Card><StatHero eyebrow="Spent so far" value={9400000} format={idr} /><StatChipRow><StatChip tone="earned">Money in Rp 15.000.000</StatChip><StatChip tone="saved">Saved Rp 5.600.000</StatChip></StatChipRow></Card>
    <div style={{ marginTop: 12 }}><PromptCard size="hero" height={233} title="Set your monthly salary" helper="We'll show what you kept once we know what came in." /></div>
    <div style={{ marginTop: 12 }}><Card><StatHero eyebrow="Up to 7 glyphs · 80" value={0} format={idr} /><StatHero eyebrow="8 to 10 · 56" value={520000} format={idr} /><StatHero eyebrow="11 and up · 40" value={15000000} format={idr} /></Card></div>
    <div style={{ marginTop: 12 }}><Banner tone="offline">No connection. You can look, but not log.</Banner></div>
    <div style={{ marginTop: 12 }}><Card><AmountField label="Salary" currency="IDR" valueMinor={15000000} /><AmountField label="Amount" currency="GBP" valueMinor={99999999999900} state="invalid" helper="That's more than the app can store." /></Card></div>
    <div style={{ marginTop: 12 }}><PageDots count={4} active={1} onSelect={() => {}} /></div>
  </Screen>],
];
const html = `<!doctype html><html><head><meta charset="utf-8"><style>${MMR_CSS}</style>
<style>body{margin:0;background:#ddd;display:flex;gap:24px;padding:24px;align-items:flex-start;width:max-content}
figure{margin:0} figcaption{font:12px sans-serif;padding:6px 0}
.wrap{position:relative} .wrap::after{content:'';position:absolute;left:0;top:844px;width:390px;border-top:2px dashed #f00;z-index:99}</style>
</head><body>${cases.map(([n, el]) =>
  `<figure><figcaption>${n}</figcaption><div class="wrap" data-name="${n}">${renderToStaticMarkup(el)}</div></figure>`
).join('')}</body></html>`;
writeFileSync(new URL('./measure.html', import.meta.url), html);
console.log('wrote measure.html with', cases.length, 'cases');
