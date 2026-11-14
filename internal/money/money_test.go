package money

import (
	"encoding/json"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want Amount
	}{
		{"12.50", 1250},
		{"-12.50", -1250},
		{"+3.00", 300},
		{"0.10", 10},
		{"12", 1200},
		{"12.5", 1250},
		{"007.00", 700},
		{"-0.00", 0},
		{"0.125", 13},
		{"-0.125", -13},
		{"0.124", 12},
		{"1.995", 200},
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if err != nil || got != c.want {
			t.Errorf("Parse(%q) = %d, %v; want %d", c.in, got, err, c.want)
		}
	}
}

func TestParseRejects(t *testing.T) {
	for _, in := range []string{"", "-", "abc", "1,50", "1.", ".5", "1.2.3", " 1.00", "1e3", "--1", "1234567890123456"} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) succeeded; want an error", in)
		}
	}
}

func TestString(t *testing.T) {
	cases := map[Amount]string{0: "0.00", 5: "0.05", -5: "-0.05", 100: "1.00", -1250: "-12.50"}
	for a, want := range cases {
		if got := a.String(); got != want {
			t.Errorf("Amount(%d).String() = %q; want %q", int64(a), got, want)
		}
	}
}

func TestSumIsExact(t *testing.T) {
	a, _ := Parse("-0.10")
	b, _ := Parse("-0.20")
	if got := (a + b).String(); got != "-0.30" {
		t.Errorf("-0.10 + -0.20 = %s; want -0.30", got)
	}
}

func TestJSON(t *testing.T) {
	var v struct {
		Amount Amount `json:"amount"`
	}
	if err := json.Unmarshal([]byte(`{"amount":"-12.50"}`), &v); err != nil || v.Amount != -1250 {
		t.Fatalf("unmarshal string: %d, %v", v.Amount, err)
	}
	out, _ := json.Marshal(v)
	if string(out) != `{"amount":"-12.50"}` {
		t.Errorf("marshal = %s", out)
	}
	if err := json.Unmarshal([]byte(`{"amount":-12.5}`), &v); err == nil {
		t.Error("a JSON number was accepted")
	}
}
