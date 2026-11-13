import { RecapCoverScreen } from '@my-money-recap/ui';

export function Default() { return <RecapCoverScreen />; }
export function OctoberLeo() {
  return <RecapCoverScreen month="October 2026" childName="Leo" total={1800} saved={310} spent={1490} />;
}
export function BeforePayday() {
  return <RecapCoverScreen month="October 2026" total={0} saved={0} spent={0} />;
}

/** D45, X4: Budi's cover in rupiah on an Indonesian phone. The hero steps down to fit and the chips wrap. */
export function Rupiah() {
  return <RecapCoverScreen month="September 2026" childName="Budi" total={150000000} saved={56000000} spent={94000000} currency="IDR" locale="id-ID" />;
}
