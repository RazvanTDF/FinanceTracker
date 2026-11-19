package csvimport

import (
	"strings"
	"testing"
)

func TestCourseExample(t *testing.T) {
	rows, rejected := Parse("date,amount,description\n2026-10-01,-0.10,coffee\n2026-10-01,-0.20,tea\n2026-13-45,xx,broken")
	if len(rows) != 2 || rows[0].Amount != -10 || rows[1].Amount != -20 {
		t.Fatalf("rows = %+v", rows)
	}
	if len(rejected) != 1 || rejected[0].Line != 4 {
		t.Fatalf("rejected = %+v; want line 4", rejected)
	}
}

func TestNoHeader(t *testing.T) {
	rows, rejected := Parse("2026-10-01,-5.00,bread\n2026-10-02,100.00,salary,income")
	if len(rejected) != 0 || len(rows) != 2 {
		t.Fatalf("rows = %+v, rejected = %+v", rows, rejected)
	}
	if rows[1].Category != "income" || rows[1].Amount != 10000 {
		t.Errorf("second row = %+v", rows[1])
	}
}

func TestHeaderInAnotherOrder(t *testing.T) {
	rows, rejected := Parse("Description,Category,Amount,Date\nrent,home,-500.00,2026-10-03")
	if len(rejected) != 0 || len(rows) != 1 {
		t.Fatalf("rows = %+v, rejected = %+v", rows, rejected)
	}
	want := Row{Date: "2026-10-03", Amount: -50000, Description: "rent", Category: "home"}
	if rows[0] != want {
		t.Errorf("row = %+v; want %+v", rows[0], want)
	}
}

func TestQuotedFieldWithComma(t *testing.T) {
	rows, _ := Parse(`2026-10-01,-5.00,"SHOP, BUCHAREST"`)
	if len(rows) != 1 || rows[0].Description != "SHOP, BUCHAREST" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestAmountFormats(t *testing.T) {
	cases := map[string]int64{
		`"-12,50"`:   -1250,
		`"1.234,56"`: 123456,
		`"1,234.56"`: 123456,
		`-7`:         -700,
		`" 3.10 "`:   310,
	}
	for in, want := range cases {
		rows, rejected := Parse("2026-10-01," + in + ",x")
		if len(rows) != 1 || int64(rows[0].Amount) != want {
			t.Errorf("amount %s: rows %+v, rejected %+v; want %d", in, rows, rejected, want)
		}
	}
}

func TestDateFormats(t *testing.T) {
	for _, in := range []string{"2026-10-01", "01.10.2026", "01/10/2026"} {
		rows, _ := Parse(in + ",-1.00,x")
		if len(rows) != 1 || rows[0].Date != "2026-10-01" {
			t.Errorf("date %s: rows %+v", in, rows)
		}
	}
}

func TestBadRowsAreRejectedAndTheRestIsKept(t *testing.T) {
	file := strings.Join([]string{
		"date,amount,description",
		"2026-10-01,-1.00,ok one",
		"2026-10-01,-12,50,unquoted decimal comma",
		"2026-02-30,-1.00,no such day",
		"2026-10-01,abc,not a number",
		"2026-10-01,-1.00,",
		`2026-10-01,-1.00,"unclosed`,
		"",
		"2026-10-02,-2.00,ok two\r",
	}, "\n")
	rows, rejected := Parse(file)
	if len(rows) != 2 || rows[1].Description != "ok two" {
		t.Errorf("rows = %+v", rows)
	}
	var lines []int
	for _, r := range rejected {
		if r.Reason == "" {
			t.Errorf("line %d has no reason", r.Line)
		}
		lines = append(lines, r.Line)
	}
	want := []int{3, 4, 5, 6, 7}
	if len(lines) != len(want) {
		t.Fatalf("rejected lines = %v; want %v", lines, want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("rejected lines = %v; want %v", lines, want)
			break
		}
	}
}

func TestHeaderWithoutAmount(t *testing.T) {
	_, rejected := Parse("date,description\n2026-10-01,-1.00,x")
	if len(rejected) == 0 || rejected[0].Line != 1 {
		t.Errorf("rejected = %+v; want the header on line 1", rejected)
	}
}
