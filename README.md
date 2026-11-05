# Personal Finance Tracker

A small web service for keeping track of what you earn and spend. Transactions
come in from a bank's CSV export or are added by hand, rules you write sort
them into categories, and a monthly report adds them up exactly and tells you
which categories went over budget. It is meant for one person watching their
own money.

- **Student:** Todor Ovidiu-Razvan, group 1
- **Project:** 2, Personal Finance Tracker
- **Language:** Go

## What it does

The service speaks JSON over HTTP on port 8080. Amounts are always JSON
strings with two decimals, negative for money going out (`"-12.50"`), and are
held internally as whole cents in an `int64`, never as a floating-point
number, so `"0.10"` plus `"0.20"` is `"0.30"` every time.

A CSV import does not trust the file: a malformed row is rejected and reported
with its line number and the reason, and the rest of the file is still
imported. A transaction without a category gets one from the first rule in
`/rules` whose `match` it satisfies, or `uncategorised` if none does.

It does not handle multiple users, multiple currencies, or bank connections.
Data lives in memory and is lost when the service stops.

**Current state (November 2026):** the four common routes work; the finance
routes are being built one at a time.

| Route | Method | State |
| --- | --- | --- |
| `/health`, `/version`, `/` | GET | done |
| `/reset` | POST | done |
| `/transactions` | POST | planned |
| `/transactions/import` | POST | planned |
| `/transactions` | GET | planned |
| `/rules` | PUT | planned |
| `/budgets` | PUT | planned |
| `/reports/monthly` | GET | planned |

## How to run it

```
docker run --rm -p 8080:8080 ghcr.io/razvantdf/financetracker:latest
```

To build and run without Docker (Go 1.27):

```
go run ./cmd/app
go run ./cmd/app --version
go test ./...
```

To build the image yourself:

```
docker build -t financetracker .
docker run --rm -p 8080:8080 financetracker
```

## What you should see

```
$ curl -i http://localhost:8080/health
HTTP/1.1 200 OK
Content-Type: application/json

{"status":"ok","uptime_seconds":3}
```

`http://localhost:8080/` in a browser shows the project and student details,
and `/version` reports the commit the image was built from.

The check this project is built around, once import and reports exist: import
two expenses of `-0.10` and `-0.20` plus one broken row, then ask for the
month.

```
curl -i -H "Content-Type: application/json" \
  -d '{"csv":"date,amount,description\n2026-10-01,-0.10,coffee\n2026-10-01,-0.20,tea\n2026-13-45,xx,broken"}' \
  http://localhost:8080/transactions/import
curl -i "http://localhost:8080/reports/monthly?month=2026-10"
```

The import reports two rows imported and line 4 rejected; the report's
`total_out` is exactly `"-0.30"`.
