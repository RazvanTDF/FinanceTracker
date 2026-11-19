// Package csvimport reads a bank export. A bad row is reported with its line
// number and the rest of the file is still read.
package csvimport

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
	"time"

	"atad-project/internal/money"
)

type Row struct {
	Date        string
	Amount      money.Amount
	Description string
	Category    string
}

type Rejection struct {
	Line   int    `json:"line"`
	Reason string `json:"reason"`
}

// columns says where each field is in a row. Without a header the order is
// date, amount, description and an optional category.
type columns struct {
	date, amount, description, category int
	count                               int
}

var noHeader = columns{date: 0, amount: 1, description: 2, category: 3}

var dateLayouts = []string{"2006-01-02", "02.01.2006", "02/01/2006"}

// Parse reads the whole file. Lines are numbered from 1, counting the header
// and empty lines, so the numbers match what the user sees in an editor.
func Parse(text string) ([]Row, []Rejection) {
	var rows []Row
	var rejected []Rejection
	cols := noHeader
	first := true

	for i, line := range strings.Split(text, "\n") {
		n := i + 1
		line = strings.TrimSuffix(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields, err := csv.NewReader(strings.NewReader(line)).Read()
		if err != nil {
			rejected = append(rejected, Rejection{n, "the line is not valid CSV"})
			first = false
			continue
		}
		if first {
			first = false
			if isHeader(fields) {
				h, err := readHeader(fields)
				if err != nil {
					rejected = append(rejected, Rejection{n, err.Error()})
				} else {
					cols = h
				}
				continue
			}
		}
		row, err := readRow(fields, cols)
		if err != nil {
			rejected = append(rejected, Rejection{n, err.Error()})
			continue
		}
		rows = append(rows, row)
	}
	return rows, rejected
}

func isHeader(fields []string) bool {
	for _, f := range fields {
		if strings.EqualFold(strings.TrimSpace(f), "date") {
			return true
		}
	}
	return false
}

func readHeader(fields []string) (columns, error) {
	cols := columns{date: -1, amount: -1, description: -1, category: -1, count: len(fields)}
	for i, f := range fields {
		switch strings.ToLower(strings.TrimSpace(f)) {
		case "date":
			cols.date = i
		case "amount":
			cols.amount = i
		case "description":
			cols.description = i
		case "category":
			cols.category = i
		}
	}
	switch {
	case cols.date < 0:
		return noHeader, errors.New("the header has no date column")
	case cols.amount < 0:
		return noHeader, errors.New("the header has no amount column")
	case cols.description < 0:
		return noHeader, errors.New("the header has no description column")
	}
	return cols, nil
}

func readRow(fields []string, cols columns) (Row, error) {
	if cols.count > 0 && len(fields) != cols.count {
		return Row{}, fmt.Errorf("expected %d fields, found %d", cols.count, len(fields))
	}
	if cols.count == 0 && len(fields) != 3 && len(fields) != 4 {
		return Row{}, fmt.Errorf("expected 3 or 4 fields, found %d", len(fields))
	}
	get := func(i int) string {
		if i < 0 || i >= len(fields) {
			return ""
		}
		return strings.TrimSpace(fields[i])
	}

	date, err := readDate(get(cols.date))
	if err != nil {
		return Row{}, err
	}
	amount, err := readAmount(get(cols.amount))
	if err != nil {
		return Row{}, err
	}
	desc := get(cols.description)
	if desc == "" {
		return Row{}, errors.New("the description is empty")
	}
	return Row{Date: date, Amount: amount, Description: desc, Category: get(cols.category)}, nil
}

// readDate accepts YYYY-MM-DD, DD.MM.YYYY and DD/MM/YYYY and always returns
// YYYY-MM-DD.
func readDate(s string) (string, error) {
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format(time.DateOnly), nil
		}
	}
	return "", fmt.Errorf("date %q is not a real date (use YYYY-MM-DD)", s)
}

// readAmount accepts a decimal point or a decimal comma. When both appear,
// the last one is the decimal separator and the other one groups thousands,
// so "1.234,56" and "1,234.56" are both 1234.56.
func readAmount(s string) (money.Amount, error) {
	orig := s
	s = strings.ReplaceAll(s, " ", "")
	dot, comma := strings.LastIndex(s, "."), strings.LastIndex(s, ",")
	switch {
	case comma > dot:
		s = strings.ReplaceAll(s, ".", "")
		if strings.Count(s, ",") > 1 {
			return 0, fmt.Errorf("amount %q is not a number like -12.50", orig)
		}
		s = strings.Replace(s, ",", ".", 1)
	case comma >= 0:
		s = strings.ReplaceAll(s, ",", "")
	}
	a, err := money.Parse(s)
	if err != nil {
		return 0, fmt.Errorf("amount %q is not a number like -12.50", orig)
	}
	return a, nil
}
