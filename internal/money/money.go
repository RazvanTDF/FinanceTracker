// Package money holds amounts as whole cents, so that sums are exact.
package money

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Amount is a sum of money in cents. Negative means money going out.
type Amount int64

var ErrInvalid = errors.New("amount must be a decimal like \"-12.50\"")

// Parse reads a decimal such as "-12.50". More than two decimals are rounded
// half away from zero, so "0.125" becomes 13 cents.
func Parse(s string) (Amount, error) {
	neg := false
	if s != "" && (s[0] == '-' || s[0] == '+') {
		neg = s[0] == '-'
		s = s[1:]
	}
	whole, frac, hasDot := strings.Cut(s, ".")
	if !isDigits(whole) || (hasDot && !isDigits(frac)) || len(whole) > 15 {
		return 0, ErrInvalid
	}

	var cents int64
	for _, c := range whole {
		cents = cents*10 + int64(c-'0')
	}
	padded := frac + "00"
	cents = cents*100 + int64(padded[0]-'0')*10 + int64(padded[1]-'0')
	if len(frac) > 2 && frac[2] >= '5' {
		cents++
	}

	if neg {
		cents = -cents
	}
	return Amount(cents), nil
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// String formats the amount with exactly two decimals, e.g. "-12.50".
func (a Amount) String() string {
	sign := ""
	c := int64(a)
	if c < 0 {
		sign = "-"
		c = -c
	}
	return fmt.Sprintf("%s%d.%02d", sign, c/100, c%100)
}

func (a Amount) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

// UnmarshalJSON accepts only a JSON string; a JSON number is an error.
func (a *Amount) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return ErrInvalid
	}
	v, err := Parse(s)
	if err != nil {
		return err
	}
	*a = v
	return nil
}
