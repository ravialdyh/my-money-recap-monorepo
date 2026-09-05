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
export type Category = 'snacks' | 'games' | 'toys' | 'rides' | 'clothes' | 'bills' | 'gifts' | 'other';
export const CATEGORIES: readonly Category[] = ['snacks', 'games', 'toys', 'rides', 'clothes', 'bills', 'gifts', 'other'] as const;
export const CATEGORY_LABEL: Record<Category, string> = {
  snacks: 'Snacks', games: 'Games', toys: 'Toys', rides: 'Rides',
  clothes: 'Clothes', bills: 'Bills', gifts: 'Gifts', other: 'Other',
};
/** Category to hue. Snacks, Games and Toys keep the hues drawn on the Goal artboard; no two categories share one. */
export const CATEGORY_TONE: Record<Category, Hue> = {
  snacks: 'coral', games: 'violet', toys: 'yellow', rides: 'sky',
  clothes: 'green', bills: 'ink', gifts: 'gold', other: 'rust',
};
