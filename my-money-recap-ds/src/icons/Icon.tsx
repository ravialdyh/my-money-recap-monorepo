import type { SVGProps } from 'react';
import {
  PawIcon, DropletIcon, LemonIcon, ShareIcon, SparkleIcon,
  SnacksIcon, GamesIcon, ToysIcon, RidesIcon, ClothesIcon, BillsIcon, GiftsIcon, OtherIcon,
  SettingsIcon, ChevronIcon, PlusIcon, PiggyMarker,
} from './glyphs';

const GLYPHS = {
  paw: PawIcon, droplet: DropletIcon, lemon: LemonIcon, share: ShareIcon, sparkle: SparkleIcon,
  snacks: SnacksIcon, games: GamesIcon, toys: ToysIcon, rides: RidesIcon, clothes: ClothesIcon, bills: BillsIcon, gifts: GiftsIcon, other: OtherIcon,
  settings: SettingsIcon, chevron: ChevronIcon, plus: PlusIcon, piggy: PiggyMarker,
} as const;
export type IconName = keyof typeof GLYPHS;
export const ICON_NAMES = Object.keys(GLYPHS) as readonly IconName[];

/** 20px stroked glyphs that take their colour from `currentColor`. */
const STROKED: readonly IconName[] = ['share', 'settings', 'chevron', 'plus'] as const;

export interface IconProps extends Omit<SVGProps<SVGSVGElement>, 'name'> {
  /** Which glyph to draw. Add new glyphs to ./glyphs.tsx and the `GLYPHS` map. */
  name: IconName;
  /** Square size in px. Ignored by the marker (`piggy`), which carries its own width and height. Defaults: 16 for filled glyphs, 20 for stroked ones. */
  size?: number;
}

/** The one icon component. The eight Category glyphs and `paw`/`droplet`/`lemon` are 16px filled glyphs for IconWell; `share`/`settings`/`chevron`/`plus` are 20px stroked; `sparkle` is decorative; `piggy` is the saving-target bar marker. */
export function Icon({ name, size, ...rest }: IconProps) {
  const Glyph = GLYPHS[name];
  if (name === 'piggy') return <Glyph {...rest} />;
  const px = size ?? (STROKED.includes(name) ? 20 : 16);
  return <Glyph width={px} height={px} {...rest} />;
}
