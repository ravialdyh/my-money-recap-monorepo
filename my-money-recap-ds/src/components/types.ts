/** Forced visual state for documentation and previews. Interactive components also respond to real :active/:disabled. */
export type InteractiveState = 'default' | 'pressed' | 'disabled' | 'selected';
export type Ground = 'recap' | 'earning' | 'goal' | 'personality';
export type DataTone = 'earned' | 'saved' | 'spent';
/** The eight palette hues. Every bar track, fill and icon well is one of these. */
export type Hue = 'coral' | 'violet' | 'yellow' | 'green' | 'sky' | 'gold' | 'rust' | 'ink';
export type BarTone = Hue;
export type WellTone = Hue;
export type Archetype = 'squirrel' | 'hustler' | 'treatYourself' | 'slowBurner';
export const ARCHETYPE_LABEL: Record<Archetype, string> = {
  squirrel: 'Squirrel', hustler: 'Hustler', treatYourself: 'Treat-Yourself', slowBurner: 'Slow Burner',
};

/** The eight fixed Spend Categories (PRD D20). No custom Categories exist. */
export const CATEGORY_LABEL = {
  snacks: 'Snacks', games: 'Games', toys: 'Toys', rides: 'Rides',
  clothes: 'Clothes', bills: 'Bills', gifts: 'Gifts', other: 'Other',
} as const;
export type Category = keyof typeof CATEGORY_LABEL;
export const CATEGORIES = Object.keys(CATEGORY_LABEL) as readonly Category[];
/** Category to hue (PRD D20). No two categories share one. */
export const CATEGORY_TONE: Record<Category, Hue> = {
  snacks: 'coral', games: 'violet', toys: 'yellow', rides: 'sky',
  clothes: 'green', bills: 'ink', gifts: 'gold', other: 'rust',
};
