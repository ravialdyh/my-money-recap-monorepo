# My Money Recap

A personal money tracker for adults, styled as a playful kids' app, that turns one month of income, spending and saving into a Wrapped-style monthly recap.

## Language

### System

**BFF**:
The one backend the app talks to (backend for frontend): it verifies the sign-in token, then joins answers from the services (`ARCHITECTURE.md` §6).
_Avoid_: API, server, gateway

### People

**Member**:
The adult (18+) who owns an account and tracks their own money. There is exactly one Member per account.
_Avoid_: Kid, child, user, parent

**Display Name**:
The first name the Member chooses at setup, shown on Home and in the recap ("Maya's recap").
_Avoid_: Username, nickname

**Avatar**:
One of a fixed set of illustrated characters the Member picks at setup to represent themselves.
_Avoid_: Profile picture, photo

### Money

**Salary**:
The fixed amount of money the Member expects to receive each month, set once at setup and editable later. It is the only recurring income in v1.
_Avoid_: Allowance, wage, pay, earnings, monthly income

**Spend**:
A single purchase the Member logs by hand, with an amount, a Category and a date.
_Avoid_: Transaction, expense, purchase, payment

**Category**:
One of eight fixed labels a Spend belongs to: Snacks, Games, Toys, Rides, Clothes, Bills, Gifts, Other.
_Avoid_: Tag, type

**Target Warning**:
The once-a-month alert raised the first time a month's Spends exceed the **warning line**, which is Salary minus the Saving Target amount. Its wording differs before and after Payday (PRD D53).
_Avoid_: Budget alert, overspend alert

**Payday**:
The day of the month on which the Salary is counted as received. Stored as 1 to 31, or the token `Last day`. A stored 29, 30 or 31 slides to the month's last day in a shorter month; the picker shows the three real dates that follow the selection rather than explaining the rule (PRD D51).
_Avoid_: Pay date, salary date

**Money In**:
Everything that arrived in a calendar month. In v1 this is the Salary, counted from its Payday.
_Avoid_: Income, revenue, "money that came your way" (display copy only)

**Currency**:
The single currency an account holds all amounts in. Changing it converts every stored amount once, at that moment's exchange rate.
_Avoid_: Locale, region

**Saved**:
The computed amount left over in a month: Money In minus the total of Spends. It is never logged directly, and it is **not shown at all while Money In is 0** — before Payday a Member is spending last month's salary, so a negative Saved would state the opposite of their real position (PRD D41).
_Avoid_: Savings, balance, leftover

**Saving Target**:
The share of Salary, stored as a percentage, that the Member intends to keep as Saved each month.
_Avoid_: Goal, savings goal, budget, target saving

**Splurge**:
The single largest Spend in a month.
_Avoid_: Biggest purchase, top expense

### Recap

**Recap**:
The read-only monthly story, four screens long, computed live from one calendar month of data.
_Avoid_: Report, summary, Wrapped

**Recap Month**:
The calendar month a Recap covers, in the Member's local dates. Never a pay period.
_Avoid_: Period, cycle

**Streak**:
The longest run of consecutive days with no Spend logged, inside the month. In the Member's sign-up month the run starts on the day they signed up, not the 1st, so no badge is ever awarded for days before the app existed (PRD D43).
_Avoid_: No-spend run

**Partial Recap**:
The four-screen Recap shown for a partial month (PRD §8.11): the sign-up month when Payday fell before sign-up, or a past month with Spends but no Money In. Built around Spent, with no Saved figure, no Archetype and no sharing. The same four screens serve the current month's live preview. "Partial month" is the month's state; "Partial Recap" is its Recap.
_Avoid_: Mini recap, preview recap

**Archetype**:
The single money personality (Squirrel, Hustler, Treat-Yourself, Slow Burner) assigned to a Member for one month by rules over that month's numbers.
_Avoid_: Personality type, badge, persona
