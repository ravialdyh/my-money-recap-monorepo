import type { Archetype } from '../components/types';
import { moneyFormat } from '../money';

/** One-line Archetype descriptions, verbatim from PRD §8.6. */
export const ARCHETYPE_BLURB: Record<Archetype, string> = {
  squirrel: "You stashed more than you meant to, and you didn't crack once.",
  hustler: 'You hit your number while still living a little.',
  treatYourself: 'One big treat tipped the month. Worth it?',
  slowBurner: 'Lots of little spends added up. Sneaky.',
};

/** Every Recap screen formats money the same way (PRD D45, D71): `currency` is the Member's ISO 4217
 *  code, `locale` the viewer's language tag (omit for the device language). See `../money.ts`. The
 *  sample defaults are Maya's September (PRD §8), in pounds. */
export interface MoneyProps {
  currency?: string;
  locale?: string;
}
export const useMoney = ({ currency = 'GBP', locale }: MoneyProps) => moneyFormat(currency, locale);
