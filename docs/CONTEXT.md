# Glossary

The words this project uses, and the words it avoids. Use each term exactly as written here: in code, tests,
issues and documents. The rules behind the money terms are in
[MONEY-RULES.md](../apps/backend/cmd/ledger/MONEY-RULES.md).

## System

**BFF**:
The one backend the app talks to (backend for frontend). It checks the sign-in token, then joins the services'
answers into one answer per screen.
_Avoid_: API, server, gateway

## People

**Member**:
The adult (18+) who owns an account and tracks their own money. One Member per account.
_Avoid_: Kid, child, user, parent

**Display Name**:
The first name the Member chooses at setup, shown on Home and in the Recap ("Maya's recap").
_Avoid_: Username, nickname

**Avatar**:
One of six illustrated characters the Member picks at setup.
_Avoid_: Profile picture, photo

## Money

**Salary**:
The fixed take-home amount, after tax, the Member expects each month. Set at setup and editable later.
_Avoid_: Allowance, wage, pay, earnings, monthly income

**Spend**:
One amount the Member spent, logged by hand with a Category and a date.
_Avoid_: Transaction, expense, purchase, payment

**Category**:
One of eight fixed labels for a Spend: Snacks, Games, Toys, Rides, Clothes, Bills, Gifts, Other.
_Avoid_: Tag, type

**Spent**:
The total of a month's Spends.
_Avoid_: Expenses, spending total

**Target amount**:
The Saving Target as money: the Salary times its percent. Salary minus the Target amount is the **saving line**.
_Avoid_: Goal amount, budget

**Target Warning**:
The once-a-month alert raised the first time Spent passes the saving line. Its wording differs before and after
Payday.
_Avoid_: Budget alert, overspend alert

**Payday**:
The day of the month the Salary counts as received: 1 to 31, or `last`, shown as `Last day`.
_Avoid_: Pay date, salary date

**Money In**:
What arrived in a calendar month. Today that is the Salary, counted from its Payday.
_Avoid_: Income, revenue

**Currency**:
The one currency an account holds every amount in. Changing it converts every stored amount once, at that
moment's rate.
_Avoid_: Locale, region

**Saved**:
Money In minus Spent. It is worked out, never logged.
_Avoid_: Savings, balance, leftover

**Saving Target**:
The percent of the Salary the Member means to keep as Saved each month.
_Avoid_: Goal, savings goal, budget, target saving

**Splurge**:
The single largest Spend in a month.
_Avoid_: Biggest purchase, top expense

## Recap

**Recap**:
The read-only story of one calendar month, four pages long, worked out live from that month's data.
_Avoid_: Report, summary, Wrapped

**Recap Month**:
The calendar month a Recap covers, in the Member's local dates. Never a pay period.
_Avoid_: Period, cycle

**Streak**:
The longest run of days in the month with no Spend.
_Avoid_: No-spend run

**Partial Recap**:
The Recap of a partial month. It shows Spent, with no Saved, no Archetype and no sharing. The same four pages
show the current month's preview. "Partial month" is the month; "Partial Recap" is its Recap.
_Avoid_: Mini recap, preview recap

**Archetype**:
The money personality a Member gets for one month: Squirrel, Hustler, Treat-Yourself or Slow Burner.
_Avoid_: Personality type, badge, persona
