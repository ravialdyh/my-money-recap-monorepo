# Glossary

The words this project uses, and the words it avoids. Use each term exactly as written here: in code, tests,
issues and documents.

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
The fixed take-home amount, after tax, the Member expects each month. Set at setup and editable later. It is the
only income the app knows.
_Avoid_: Allowance, wage, pay, earnings, monthly income

**Spend**:
One purchase the Member logs by hand, with an amount, a Category and a date.
_Avoid_: Transaction, expense, purchase, payment

**Category**:
One of eight fixed labels for a Spend: Snacks, Games, Toys, Rides, Clothes, Bills, Gifts, Other.
_Avoid_: Tag, type

**Target Warning**:
The once-a-month alert raised the first time a month's Spends pass the **warning line**: Salary minus the Saving
Target amount. Its wording differs before and after Payday.
_Avoid_: Budget alert, overspend alert

**Payday**:
The day of the month the Salary counts as received: 1 to 31, or `last`, shown as `Last day`. A 29, 30 or 31
moves to the last day of a shorter month.
_Avoid_: Pay date, salary date

**Money In**:
What arrived in a calendar month. Today that is the Salary, counted from its Payday.
_Avoid_: Income, revenue

**Currency**:
The one currency an account holds every amount in. Changing it converts every stored amount once, at that
moment's rate.
_Avoid_: Locale, region

**Saved**:
Money In minus the month's Spends. It is worked out, never logged, and not shown at all while Money In is 0.
_Avoid_: Savings, balance, leftover

**Saving Target**:
The percent of the Salary the Member means to keep as Saved each month.
_Avoid_: Goal, savings goal, budget, target saving

**Splurge**:
The single largest Spend in a month.
_Avoid_: Biggest purchase, top expense

## Recap

**Recap**:
The read-only story of one calendar month, four screens long, worked out live from that month's data.
_Avoid_: Report, summary, Wrapped

**Recap Month**:
The calendar month a Recap covers, in the Member's local dates. Never a pay period.
_Avoid_: Period, cycle

**Streak**:
The longest run of days in the month with no Spend. In the sign-up month it starts on the sign-up day, not the 1st.
_Avoid_: No-spend run

**Partial Recap**:
The Recap of a partial month: the sign-up month when Payday came before sign-up, or a past month with Spends but
no Money In. It shows Spent, with no Saved, no Archetype and no sharing. The same four screens show the current
month's preview. "Partial month" is the month; "Partial Recap" is its Recap.
_Avoid_: Mini recap, preview recap

**Archetype**:
The money personality a Member gets for one month: Squirrel, Hustler, Treat-Yourself or Slow Burner. The rules are
in `apps/backend/cmd/ledger/MONEY-RULES.md`.
_Avoid_: Personality type, badge, persona
