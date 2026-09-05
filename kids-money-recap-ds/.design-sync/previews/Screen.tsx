import { Screen, Headline, SubCopy } from '@kids-money-recap/ui';

export function Recap() {
  return <Screen ground="recap"><Headline size="m">Violet ground</Headline><SubCopy style={{ marginTop: 10 }}>White text, the default for recap screens.</SubCopy></Screen>;
}
export function Earning() {
  return <Screen ground="earning"><Headline size="m">Yellow ground, dark ink</Headline><SubCopy style={{ marginTop: 10 }}>The only ground that switches text to ink.</SubCopy></Screen>;
}
export function Goal() {
  return <Screen ground="goal"><Headline size="m">Coral ground</Headline><SubCopy style={{ marginTop: 10 }}>Goal and spending screens.</SubCopy></Screen>;
}
export function Personality() {
  return <Screen ground="personality"><Headline size="m">Green ground</Headline><SubCopy style={{ marginTop: 10 }}>The money-personality reveal.</SubCopy></Screen>;
}
