import { RecapEarningScreen } from '@kids-money-recap/ui';

export function Default() { return <RecapEarningScreen />; }
export function MissedTarget() {
  return (
    <RecapEarningScreen
      month="October" moneyIn={2400} saved={310} targetAmount={480}
      streakDays={4} streakRange={[8, 11]} streakCaption="8 to 11 October. Beat it next month."
      daysInMonth={31}
    />
  );
}
export function NoTargetSet() {
  return <RecapEarningScreen month="October" moneyIn={2400} saved={310} targetAmount={null} streakDays={4} streakRange={[8, 11]} streakCaption="8 to 11 October. Beat it next month." daysInMonth={31} />;
}
