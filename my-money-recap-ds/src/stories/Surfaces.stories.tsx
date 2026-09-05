import type { Meta, StoryObj } from './_csf';
import { Banner, type BannerProps } from '../components/Banner';
import { Sheet } from '../components/Sheet';
import { ListRow } from '../components/ListRow';
import { DataRowList } from '../components/DataRow';
import { AvatarGrid, Avatar } from '../components/AvatarGrid';
import { ILLUSTRATIONS } from '../assets/illustrations';

const meta: Meta<BannerProps> = { title: 'Components/Surfaces', component: Banner, parameters: { backgrounds: { default: 'violet' } } };
export default meta;
type S = StoryObj<BannerProps>;
const avatars = [
  { key: 'fox' as const, src: ILLUSTRATIONS['avatar-fox'] },
  { key: 'otter' as const, src: ILLUSTRATIONS['avatar-otter'] },
  { key: 'panda' as const, src: ILLUSTRATIONS['avatar-panda'] },
  { key: 'frog' as const, src: ILLUSTRATIONS['avatar-frog'] },
  { key: 'owl' as const, src: ILLUSTRATIONS['avatar-owl'] },
  { key: 'rabbit' as const, src: ILLUSTRATIONS['avatar-rabbit'] },
];
export const TargetWarning: S = { args: { children: "You've passed your saving line. £185 over." } };
export const Dismissable: S = { args: { children: "You've passed your saving line. £185 over.", onDismiss: () => {} } };
export const BottomSheet: S = { render: () => <div style={{ width: 390 }}><Sheet title="Log a spend">Sheet body</Sheet></div> };
export const SettingsRows: S = { render: () => (
  <div style={{ width: 342 }}><DataRowList>
    <ListRow label="Profile" value="Maya" />
    <ListRow label="Currency" value="GBP" />
    <ListRow label="Delete account" tone="destructive" />
  </DataRowList></div>
) };
export const Avatars: S = { render: () => <div style={{ width: 300 }}><AvatarGrid avatars={avatars} selected="otter" /></div> };
export const SingleAvatar: S = { render: () => <Avatar src={ILLUSTRATIONS['avatar-otter']} label="Otter" /> };
