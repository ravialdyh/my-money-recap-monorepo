// Throwaway: renders the screens Claude Design flagged and measures overflow past 844.
import { renderToStaticMarkup } from 'react-dom/server';
import { writeFileSync } from 'node:fs';
import { KMR_CSS } from '../src/styles/css.generated';
import { RecapGoalScreen, RecapEarningScreen, RecapPersonalityScreen } from '../src/screens/RecapScreens';

const cases: Array<[string, JSX.Element]> = [
  ['C3 goal · 5 bars + splurge', <RecapGoalScreen />],
  ['C3 goal · no spends', <RecapGoalScreen saved={2400} target={480} helper="Target met. £1,920 to spare." subCopy="Not a single spend logged." spentTotal={0} categories={[]} splurge={null} />],
  ['C2 earning · no target (null)', <RecapEarningScreen month="October" moneyIn={2400} saved={310} targetAmount={null} streakDays={4} streakRange={[8,11]} daysInMonth={31} />],
  ['C4 squirrel', <RecapPersonalityScreen archetype="squirrel" />],
  ['C4 hustler (bee)', <RecapPersonalityScreen archetype="hustler" saved={520} spent={1880} />],
  ['C4 treat (cat)', <RecapPersonalityScreen archetype="treatYourself" saved={210} spent={2190} />],
  ['C4 slowburner (tortoise)', <RecapPersonalityScreen archetype="slowBurner" saved={180} spent={2220} />],
];
const html = `<!doctype html><html><head><meta charset="utf-8"><style>${KMR_CSS}</style>
<style>body{margin:0;background:#ddd;display:flex;gap:24px;padding:24px;align-items:flex-start;width:max-content}
figure{margin:0} figcaption{font:12px sans-serif;padding:6px 0}
.wrap{position:relative} .wrap::after{content:'';position:absolute;left:0;top:844px;width:390px;border-top:2px dashed #f00;z-index:99}</style>
</head><body>${cases.map(([n, el]) =>
  `<figure><figcaption>${n}</figcaption><div class="wrap" data-name="${n}">${renderToStaticMarkup(el)}</div></figure>`
).join('')}</body></html>`;
writeFileSync(new URL('./measure.html', import.meta.url), html);
console.log('wrote measure.html with', cases.length, 'cases');
