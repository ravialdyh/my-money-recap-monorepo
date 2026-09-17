// Renders the four Recap screens and the component boards to static HTML for a visual check. Not part of the package API.

import { writeFileSync } from 'node:fs';
import { renderToStaticMarkup } from 'react-dom/server';
import { ILLUSTRATIONS } from '../src/assets/illustrations';
import { AmountField } from '../src/components/AmountField';
import { ArchetypePill } from '../src/components/ArchetypePill';
import { AvatarGrid } from '../src/components/AvatarGrid';
import { Banner } from '../src/components/Banner';
import { CategoryBar } from '../src/components/Bars';
import { Button } from '../src/components/Button';
import { Card, CardTitle } from '../src/components/Card';
import { Chip } from '../src/components/Chip';
import { DataRowList, IconWell } from '../src/components/DataRow';
import { ListRow } from '../src/components/ListRow';
import { Sheet } from '../src/components/Sheet';
import { Slider } from '../src/components/Slider';
import { StatChip } from '../src/components/StatChip';
import { TextField } from '../src/components/TextField';
import { Toggle } from '../src/components/Toggle';
import { CATEGORIES, CATEGORY_LABEL, CATEGORY_TONE } from '../src/components/types';
import { ICON_NAMES, Icon } from '../src/icons';
import { RecapCoverScreen, RecapEarningScreen, RecapGoalScreen, RecapPersonalityScreen } from '../src/screens';
import { MMR_CSS } from '../src/styles/css.generated';

const screens = [<RecapCoverScreen />, <RecapEarningScreen />, <RecapGoalScreen />, <RecapPersonalityScreen />].map(
  renderToStaticMarkup,
);
const states = renderToStaticMarkup(
  <div style={{ display: 'grid', gap: 16, padding: 24, background: '#5B4FE8', width: 390, boxSizing: 'border-box' }}>
    <div style={{ display: 'flex', gap: 8 }}>
      <Chip>Default</Chip>
      <Chip state="pressed">Pressed</Chip>
      <Chip state="disabled">Disabled</Chip>
    </div>
    <div style={{ display: 'flex', gap: 8 }}>
      <StatChip tone="saved">Default</StatChip>
      <StatChip tone="saved" state="pressed">
        Pressed
      </StatChip>
      <StatChip tone="saved" state="disabled">
        Disabled
      </StatChip>
    </div>
    <div style={{ display: 'grid', gap: 8 }}>
      <Button icon={<Icon name="share" />}>Default</Button>
      <Button icon={<Icon name="share" />} state="pressed">
        Pressed
      </Button>
      <Button icon={<Icon name="share" />} state="disabled">
        Disabled
      </Button>
    </div>
    <div style={{ display: 'flex', gap: 6 }}>
      <ArchetypePill archetype="squirrel" selected />
      <ArchetypePill archetype="hustler" />
      <ArchetypePill archetype="treatYourself" state="pressed" />
      <ArchetypePill archetype="slowBurner" state="disabled" />
    </div>
  </div>,
);
const avatars = (['fox', 'otter', 'panda', 'frog', 'owl', 'rabbit'] as const).map((key) => ({
  key,
  src: ILLUSTRATIONS[`avatar-${key}`],
}));
const additions = renderToStaticMarkup(
  <div
    style={{
      display: 'grid',
      gap: 16,
      padding: 24,
      background: '#5B4FE8',
      width: 390,
      boxSizing: 'border-box',
      fontFamily: 'Nunito, sans-serif',
    }}
  >
    <Card style={{ gap: 12 }}>
      <CardTitle>Glyphs</CardTitle>
      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
        {CATEGORIES.map((c) => (
          <IconWell key={c} tone={CATEGORY_TONE[c]}>
            <Icon name={c} />
          </IconWell>
        ))}
      </div>
      <div style={{ display: 'flex', gap: 12, color: '#1E1B3A', alignItems: 'center' }}>
        <Icon name="settings" />
        <Icon name="chevron" />
        <Icon name="plus" />
        <Icon name="piggy" />
      </div>
      <div style={{ fontSize: 11, color: '#6B6890' }}>
        {ICON_NAMES.length} glyphs: {ICON_NAMES.join(' · ')}
      </div>
    </Card>
    <Card style={{ gap: 12 }}>
      <CardTitle>Eight categories</CardTitle>
      {CATEGORIES.map((c, i) => (
        <CategoryBar
          key={c}
          label={CATEGORY_LABEL[c]}
          percent={[35, 26, 14, 10, 7, 5, 2, 1][i]}
          tone={CATEGORY_TONE[c]}
        />
      ))}
    </Card>
    <Card style={{ gap: 14 }}>
      <CardTitle>Fields</CardTitle>
      <TextField label="First name" defaultValue="Maya" />
      <TextField
        label="Email address"
        defaultValue="maya@example"
        state="invalid"
        helper="That doesn't look like an email address."
      />
      <AmountField label="Your monthly salary" currency="GBP" locale="en-GB" value="2400" />
      <Slider value={20} helper="£480 of £2,400 each month" aria-label="Saving target" />
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          fontWeight: 800,
          fontSize: 14,
        }}
      >
        <span>Recap ready</span>
        <Toggle checked aria-label="Recap ready" />
      </div>
    </Card>
    <Banner onDismiss={() => {}}>You've passed your saving line. £185 over.</Banner>
    <Card>
      <DataRowList>
        <ListRow label="Profile" value="Maya" />
        <ListRow label="Currency" value="GBP" />
        <ListRow label="GBP · British Pound" trailing="check" selected />
        <ListRow label="Delete account" tone="destructive" />
      </DataRowList>
    </Card>
    <Card>
      <CardTitle>Pick your avatar</CardTitle>
      <AvatarGrid avatars={avatars} selected="otter" />
    </Card>
    <Sheet title="Log a spend">
      <TextField label="Amount" prefix="£" defaultValue="4.50" />
      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
        <Chip tappable variant="inverse">
          Snacks
        </Chip>
        <Chip tappable>Games</Chip>
        <Chip tappable>Rides</Chip>
      </div>
      <Button>Save</Button>
    </Sheet>
  </div>,
);
const html = `<!doctype html><html><head><meta charset="utf-8"><style>${MMR_CSS}</style><style>body{margin:0;background:#ddd;display:flex;gap:20px;padding:20px;align-items:flex-start;width:max-content} body>*{flex-shrink:0}</style></head><body>${screens.join('')}${states}${additions}</body></html>`;
writeFileSync(new URL('./index.html', import.meta.url), html);
console.log('rendered', screens.length, 'screens + states board + additions board');
