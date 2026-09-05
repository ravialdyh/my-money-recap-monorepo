import { RecapCoverScreen } from '@kids-money-recap/ui';

export function Default() { return <RecapCoverScreen />; }
export function OctoberLeo() {
  return <RecapCoverScreen month="October 2026" childName="Leo" total={1800} saved={310} spent={1490} />;
}
export function BeforePayday() {
  return <RecapCoverScreen month="October 2026" total={0} saved={0} spent={0} />;
}
