import { useStyles } from '../styles/useStyles';

/** The six onboarding Avatars (PRD D36). Keys match `ILLUSTRATIONS['avatar-*']`. */
export type AvatarKey = 'fox' | 'otter' | 'panda' | 'frog' | 'owl' | 'rabbit';
export const AVATAR_KEYS: readonly AvatarKey[] = ['fox', 'otter', 'panda', 'frog', 'owl', 'rabbit'] as const;
export const AVATAR_LABEL: Record<AvatarKey, string> = {
  fox: 'Fox', otter: 'Otter', panda: 'Panda', frog: 'Frog', owl: 'Owl', rabbit: 'Rabbit',
};

export interface AvatarGridProps {
  /** One entry per Avatar, in grid order. `src` is the transparent WebP data URI. */
  avatars: Array<{ key: AvatarKey; src: string }>;
  /** The chosen Avatar, or undefined for the empty state. */
  selected?: AvatarKey;
  /** Cell size. 68 in the onboarding grid; use `Avatar` for the 56pt Home greeting. */
  size?: number;
  onSelect?: (key: AvatarKey) => void;
}
/** 3x2 grid of round Avatars. The selected one takes the ArchetypePill ring: 2px white on a violet halo. */
export function AvatarGrid({ avatars, selected, size = 68, onSelect }: AvatarGridProps) {
  useStyles();
  return (
    <div className="mmr-avatar-grid" role="radiogroup" aria-label="Pick your avatar">
      {avatars.map(({ key, src }) => (
        <button
          key={key} type="button" role="radio" aria-checked={key === selected} aria-label={AVATAR_LABEL[key]}
          className="mmr-avatar mmr-pressable" data-selected={key === selected ? 'true' : 'false'}
          style={{ width: size, height: size }} onClick={() => onSelect?.(key)}
        >
          <img src={src} alt="" />
        </button>
      ))}
    </div>
  );
}

export interface AvatarProps { src: string; label?: string; size?: number; }
/** A single Avatar badge, for the Home greeting and the recap chip. */
export function Avatar({ src, label = '', size = 56 }: AvatarProps) {
  useStyles();
  return <span className="mmr-avatar" data-static="true" style={{ width: size, height: size }}><img src={src} alt={label} /></span>;
}
