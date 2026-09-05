import type { SVGProps } from 'react';
import {
  PawIcon, DropletIcon, LemonIcon, ShareIcon, SparkleIcon, SkateboardMarker,
  SnacksIcon, GamesIcon, ToysIcon, RidesIcon, ClothesIcon, BillsIcon, GiftsIcon, OtherIcon,
  SettingsIcon, ChevronIcon, PlusIcon, PiggyMarker,
} from './glyphs';

export type IconName =
  | 'paw' | 'droplet' | 'lemon' | 'share' | 'sparkle' | 'skateboard'
  | 'snacks' | 'games' | 'toys' | 'rides' | 'clothes' | 'bills' | 'gifts' | 'other'
  | 'settings' | 'chevron' | 'plus' | 'piggy';
export const ICON_NAMES: readonly IconName[] = [
  'paw', 'droplet', 'lemon', 'share', 'sparkle', 'skateboard',
  'snacks', 'games', 'toys', 'rides', 'clothes', 'bills', 'gifts', 'other',
  'settings', 'chevron', 'plus', 'piggy',
] as const;

/** Fixed-size bar markers: they carry their own width and height and ignore `size`. */
const MARKERS: readonly IconName[] = ['skateboard', 'piggy'] as const;
/** 20px stroked glyphs that take their colour from `currentColor`. */
const STROKED: readonly IconName[] = ['share', 'settings', 'chevron', 'plus'] as const;

export interface IconProps extends Omit<SVGProps<SVGSVGElement>, 'name'> {
  /** Which glyph to draw. Add new glyphs to ./glyphs.tsx and this union. */
  name: IconName;
  /** Square size in px. Ignored by the markers (`skateboard`, `piggy`). Defaults: 16 for filled glyphs, 20 for stroked ones. */
  size?: number;
}
const GLYPHS = {
  paw: PawIcon, droplet: DropletIcon, lemon: LemonIcon, share: ShareIcon, sparkle: SparkleIcon, skateboard: SkateboardMarker,
  snacks: SnacksIcon, games: GamesIcon, toys: ToysIcon, rides: RidesIcon, clothes: ClothesIcon, bills: BillsIcon, gifts: GiftsIcon, other: OtherIcon,
  settings: SettingsIcon, chevron: ChevronIcon, plus: PlusIcon, piggy: PiggyMarker,
} as const;

/** The one icon component. The eight Category glyphs and `paw`/`droplet`/`lemon` are 16px filled glyphs for IconWell; `share`/`settings`/`chevron`/`plus` are 20px stroked; `sparkle` is decorative; `skateboard` and `piggy` are the goal-bar markers. */
export function Icon({ name, size, ...rest }: IconProps) {
  const Glyph = GLYPHS[name];
  if (MARKERS.includes(name)) return <Glyph {...rest} />;
  const px = size ?? (STROKED.includes(name) ? 20 : 16);
  return <Glyph width={px} height={px} {...rest} />;
}
