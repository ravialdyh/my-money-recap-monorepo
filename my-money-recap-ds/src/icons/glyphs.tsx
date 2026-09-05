import type { SVGProps } from 'react';

type IconProps = SVGProps<SVGSVGElement>;

/** 16px filled glyphs used inside IconWell. Colour comes from the fill prop or currentColor. */
export function PawIcon({ fill = 'var(--mmr-palette-violet-500)', ...rest }: IconProps) {
  return (
    <svg viewBox="0 0 16 16" fill={fill} aria-hidden="true" {...rest}>
      <circle cx="4" cy="6" r="2" /><circle cx="12" cy="6" r="2" /><circle cx="6.5" cy="3" r="1.8" /><circle cx="9.5" cy="3" r="1.8" /><ellipse cx="8" cy="10.5" rx="4" ry="3.2" />
    </svg>
  );
}
export function DropletIcon({ fill = 'var(--mmr-palette-sky-500)', ...rest }: IconProps) {
  return (
    <svg viewBox="0 0 16 16" fill={fill} aria-hidden="true" {...rest}>
      <path d="M8 1.5 C8 1.5 3 7.5 3 10.5 C3 13.3 5.2 15 8 15 C10.8 15 13 13.3 13 10.5 C13 7.5 8 1.5 8 1.5 Z" />
    </svg>
  );
}
export function LemonIcon(props: IconProps) {
  return (
    <svg viewBox="0 0 16 16" fill="none" aria-hidden="true" {...props}>
      <circle cx="8" cy="9" r="6" fill="var(--mmr-palette-gold-500)" />
      <path d="M8 3 C8 3 9 0.5 12 1.5 C11 4 8 3 8 3 Z" fill="var(--mmr-palette-green-500)" />
    </svg>
  );
}
/** 20px stroked glyph used inside Button. */
export function ShareIcon(props: IconProps) {
  return (
    <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" {...props}>
      <path d="M4 11v5a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2v-5" /><path d="M10 13V3" /><path d="M6 7l4-4 4 4" />
    </svg>
  );
}
/** Decorative four-point sparkle, white by default. */
export function SparkleIcon({ fill = 'var(--mmr-palette-white)', ...rest }: IconProps) {
  return (
    <svg viewBox="0 0 32 32" fill={fill} aria-hidden="true" {...rest}>
      <path d="M16 2 L20 12 L30 16 L20 20 L16 30 L12 20 L2 16 L12 12 Z" />
    </svg>
  );
}

/* ---- Spend Category glyphs (PRD D20). 16px filled, one per Category, each defaulting to its
   own category colour so <IconWell tone={CATEGORY_TONE[c]}><Icon name={c} /></IconWell> agrees. ---- */

export function SnacksIcon({ fill = 'var(--mmr-color-data-category-snacks)', ...rest }: IconProps) {
  return (
    <svg viewBox="0 0 16 16" fill={fill} aria-hidden="true" {...rest}>
      <circle cx="8" cy="8" r="4.2" /><path d="M1 4 L4.4 8 L1 12 Z" /><path d="M15 4 L11.6 8 L15 12 Z" />
    </svg>
  );
}
export function GamesIcon({ fill = 'var(--mmr-color-data-category-games)', ...rest }: IconProps) {
  return (
    <svg viewBox="0 0 16 16" fill={fill} aria-hidden="true" {...rest}>
      <circle cx="8" cy="3.4" r="2.6" /><rect x="7" y="5.4" width="2" height="5.2" /><path d="M2 13.5 A6 3.2 0 0 1 14 13.5 Z" />
    </svg>
  );
}
export function ToysIcon({ fill = 'var(--mmr-color-data-category-toys)', ...rest }: IconProps) {
  return (
    <svg viewBox="0 0 16 16" fill={fill} aria-hidden="true" {...rest}>
      <circle cx="3.9" cy="4.6" r="2.4" /><circle cx="12.1" cy="4.6" r="2.4" /><circle cx="8" cy="9.6" r="5" />
    </svg>
  );
}
export function RidesIcon({ fill = 'var(--mmr-color-data-category-rides)', ...rest }: IconProps) {
  return (
    <svg viewBox="0 0 16 16" fill={fill} aria-hidden="true" {...rest}>
      <path d="M2.6 6.6 L4.2 3.9 A1.6 1.6 0 0 1 5.6 3.1 H10.4 A1.6 1.6 0 0 1 11.8 3.9 L13.4 6.6 Z" />
      <rect x="1" y="6.3" width="14" height="4.6" rx="1.8" />
      <circle cx="4.4" cy="12.5" r="1.9" /><circle cx="11.6" cy="12.5" r="1.9" />
    </svg>
  );
}
export function ClothesIcon({ fill = 'var(--mmr-color-data-category-clothes)', ...rest }: IconProps) {
  return (
    <svg viewBox="0 0 16 16" fill={fill} aria-hidden="true" {...rest}>
      <path d="M6 2 h4 l4 2.3 -1.7 3.1 -1.3 -0.75 V14 H5 V6.65 l-1.3 0.75 L2 4.3 Z" />
    </svg>
  );
}
export function BillsIcon({ fill = 'var(--mmr-color-data-category-bills)', ...rest }: IconProps) {
  return (
    <svg viewBox="0 0 16 16" fill={fill} aria-hidden="true" {...rest}>
      <path d="M3 1.5 h10 v11 l-1.67 -1.1 L9.67 12.5 8 11.4 6.33 12.5 4.67 11.4 3 12.5 Z" />
    </svg>
  );
}
export function GiftsIcon({ fill = 'var(--mmr-color-data-category-gifts)', ...rest }: IconProps) {
  return (
    <svg viewBox="0 0 16 16" fill={fill} aria-hidden="true" {...rest}>
      <circle cx="5.6" cy="3" r="2.1" /><circle cx="10.4" cy="3" r="2.1" />
      <rect x="1.4" y="4.4" width="13.2" height="3.1" rx="1" /><rect x="2.4" y="7.5" width="11.2" height="7" rx="1" />
      <rect x="6.9" y="4.4" width="2.2" height="10.1" fill="var(--mmr-palette-white)" />
    </svg>
  );
}
export function OtherIcon({ fill = 'var(--mmr-color-data-category-other)', ...rest }: IconProps) {
  return (
    <svg viewBox="0 0 16 16" fill={fill} aria-hidden="true" {...rest}>
      <circle cx="3" cy="8" r="1.8" /><circle cx="8" cy="8" r="1.8" /><circle cx="13" cy="8" r="1.8" />
    </svg>
  );
}

/* ---- 20px stroked UI glyphs, same construction as ShareIcon. ---- */

export function SettingsIcon(props: IconProps) {
  return (
    <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" aria-hidden="true" {...props}>
      <path d="M3 6h14" /><path d="M3 14h14" /><circle cx="8" cy="6" r="2.2" /><circle cx="13" cy="14" r="2.2" />
    </svg>
  );
}
export function ChevronIcon(props: IconProps) {
  return (
    <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" {...props}>
      <path d="M7.5 4l6 6-6 6" />
    </svg>
  );
}
export function PlusIcon(props: IconProps) {
  return (
    <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" aria-hidden="true" {...props}>
      <path d="M10 4v12" /><path d="M4 10h12" />
    </svg>
  );
}

/** Piggy-bank marker that rides the saving-target progress bar. 64x30. Rides the saving-target bar (PRD D24). */
export function PiggyMarker(props: IconProps) {
  return (
    <svg width={64} height={30} viewBox="0 0 64 30" fill="none" aria-hidden="true" {...props}>
      <circle cx="31.5" cy="4" r="3.6" fill="var(--mmr-palette-gold-500)" />
      <path d="M16 15c-3 0-3 4-6 3" stroke="var(--mmr-palette-coral-500)" strokeWidth="2.4" strokeLinecap="round" />
      <path d="M20 8.5 l5.4 1.2 -1.2 5 Z" fill="var(--mmr-palette-coral-500)" />
      <rect x="22" y="23" width="4.2" height="6" rx="1.7" fill="var(--mmr-palette-coral-500)" />
      <rect x="35.8" y="23" width="4.2" height="6" rx="1.7" fill="var(--mmr-palette-coral-500)" />
      <ellipse cx="31" cy="17" rx="15" ry="9.5" fill="var(--mmr-palette-coral-500)" />
      <ellipse cx="45" cy="18.4" rx="5" ry="4.2" fill="var(--mmr-palette-coral-500)" />
      <rect x="26" y="7.6" width="11" height="2.6" rx="1.3" fill="var(--mmr-palette-ink-900)" />
      <circle cx="37.6" cy="14.4" r="1.5" fill="var(--mmr-palette-ink-900)" />
      <circle cx="43.6" cy="18.4" r="0.9" fill="var(--mmr-palette-ink-900)" />
      <circle cx="46.4" cy="18.4" r="0.9" fill="var(--mmr-palette-ink-900)" />
      <circle cx="35" cy="13.2" r="0.7" fill="var(--mmr-palette-white)" />
    </svg>
  );
}
