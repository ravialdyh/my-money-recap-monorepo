import { AvatarGrid, Card, CardTitle, ILLUSTRATIONS } from '@kids-money-recap/ui';

const ground = { background: 'var(--kmr-color-ground-recap)', fontFamily: 'var(--kmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };
const avatars = [
  { key: 'fox' as const, src: ILLUSTRATIONS['avatar-fox'] },
  { key: 'otter' as const, src: ILLUSTRATIONS['avatar-otter'] },
  { key: 'panda' as const, src: ILLUSTRATIONS['avatar-panda'] },
  { key: 'frog' as const, src: ILLUSTRATIONS['avatar-frog'] },
  { key: 'owl' as const, src: ILLUSTRATIONS['avatar-owl'] },
  { key: 'rabbit' as const, src: ILLUSTRATIONS['avatar-rabbit'] },
];

export function Selected() { return <div style={ground}><Card><CardTitle>Pick your avatar</CardTitle><AvatarGrid avatars={avatars} selected="otter" /></Card></div>; }
export function Empty() { return <div style={ground}><Card><CardTitle>Pick your avatar</CardTitle><AvatarGrid avatars={avatars} /></Card></div>; }
