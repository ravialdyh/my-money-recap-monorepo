import { RecapPersonalityScreen } from '@my-money-recap/ui';

export function Default() { return <RecapPersonalityScreen />; }
export function Hustler() { return <RecapPersonalityScreen childName="Leo" archetype="hustler" saved={520} spent={1880} />; }
export function TreatYourself() { return <RecapPersonalityScreen childName="Ana" archetype="treatYourself" saved={210} spent={2190} />; }
export function SlowBurner() { return <RecapPersonalityScreen childName="Sam" archetype="slowBurner" saved={180} spent={2220} />; }
