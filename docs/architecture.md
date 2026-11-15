# Architecture

## The shape

The service is one Go program that answers JSON requests over HTTP on port
8080. It runs in a small Docker image as a normal user, not as root.

The code is split into a few parts:

- cmd/app/main.go starts the server, registers the routes and stops the
  server cleanly when the container is stopped.
- cmd/app/transactions.go handles the finance routes.
- internal/money reads and writes amounts of money.
- internal/store keeps the transactions in memory.
- internal/version gives the build details shown by /version.

When a request comes to POST /transactions, the server sends it to the
transaction handler. The handler reads the JSON body and checks it: the
amount must be a string like "-12.50", the date must be a real date in the
form YYYY-MM-DD, and the description cannot be empty. If something is wrong,
it answers 400 with an error message. If everything is fine, the store gives
the transaction the next id, saves it, and the handler sends it back with 201.

The routes are registered in one function, newMux, so the tests can send
requests to the same handlers without starting a real server.

## The data

For each transaction I keep an id, a date, an amount, a description and a
category. The transactions are kept in a list in memory, in the order they
were added. A mutex makes sure two requests at the same time cannot break the
list or get the same id. Ids start at 1.

Nothing is saved to disk, so all data is lost when the service stops.
POST /reset empties the list and starts the ids from 1 again. When rules and
budgets are added, they will be kept in the same place and /reset will clear
them as well.

## Three decisions

1. Amounts are stored as whole cents in an int64, so "-12.50" becomes -1250.
Adding whole numbers is always exact, so -0.10 plus -0.20 is always -0.30.
I did not use float64, because it cannot hold 0.10 exactly and the totals
would slowly drift. I also did not use a decimal library, because reading and
writing two decimals by hand is short and I understand all of it. If an
amount has more than two decimals, it is rounded half away from zero, as the
course asks.

2. The data is kept in memory and not in a database. The project does not ask
for data to survive a restart, and keeping it in memory makes /reset very
simple. I thought about SQLite, but it would mean an extra library, a table
schema and a file inside the container, and nothing in the project needs it.

3. Dates are kept as text in the form YYYY-MM-DD. They are checked when they
come in, so 2026-13-45 is refused, but they are stored as plain strings.
Dates written this way sort in the right order even as text, which is enough
for the date filters and the monthly report. The other option was Go's
time.Time, but that brings time zones into a problem that does not have any.

## What is not built yet

- Importing a CSV file, with decimal commas, quoted fields, a missing header
  line, and bad rows reported by their line number.
- What happens when the same file is imported twice. I have not decided yet
  if duplicates should be detected.
- Rules for automatic categories. For now a transaction without a category
  is "uncategorised".
- Listing transactions with the from, to, category and q filters.
- Budgets and the list of categories that went over budget.
- The monthly report, including the -0.30 check from start to end.
