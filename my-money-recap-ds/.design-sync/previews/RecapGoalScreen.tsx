import { RecapGoalScreen, formatMoney } from '@my-money-recap/ui';

export function Default() { return <RecapGoalScreen />; }
export function MissedTarget() {
  return (
    <RecapGoalScreen
      saved={310} target={480} helper="£170 short of your target." subCopy="Close. Bills took a bigger bite than usual."
      spentTotal={2090}
      categories={[
        { category: 'bills', percent: 41 },
        { category: 'clothes', percent: 22 },
        { category: 'snacks', percent: 15 },
        { category: 'rides', percent: 12 },
        { category: 'games', percent: 6 },
        { category: 'other', percent: 4 },
      ]}
      splurge="£560 on a winter coat. Your biggest single buy."
    />
  );
}
export function NoSpends() {
  return (
    <RecapGoalScreen
      saved={2400} target={480} helper="Target met. £1,920 to spare." subCopy="Not a single spend logged this month."
      spentTotal={0} categories={[]} splurge={null}
    />
  );
}

/** SCREEN-STATES C3 · no target set: the headline names the total, the sub-copy goes, the target card is a PromptCard. */
export function NoTargetSet() { return <RecapGoalScreen target={null} spentTotal={1760} />; }
/** D45: the same screen in rupiah on an Indonesian phone. */
export function Rupiah() {
  return <RecapGoalScreen currency="IDR" locale="id-ID" saved={56000000} target={30000000} spentTotal={94000000} splurge={`${formatMoney(25000000, 'IDR', 'id-ID')} on a flight home. Your biggest single buy.`} />;
}
