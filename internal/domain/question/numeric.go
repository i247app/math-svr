package question

import (
	"fmt"
	"math/big"
	"regexp"
	"slices"
	"strings"
)

// Limits of a NUMERIC_INPUT key.
const (
	MaxNumericValues = 10
	MaxNumberLen     = 32 // characters of one value, sign and separator included
)

// numberPattern is an integer, a decimal ("." or ",") or a fraction a/b.
// No exponent, no "+", no spaces inside.
var numberPattern = regexp.MustCompile(`^(-?)(\d+)(?:[.,](\d+)|/(\d+))?$`)

// ParseNumber reads one numeric answer, typed by an admin or a child. It
// returns the value (exact — big.Rat, never a float) and the canonical way
// to store it: trimmed, "." as the separator, no leading zeros, no trailing
// zeros after the point, no sign on zero. A fraction keeps its notation
// unreduced ("2/4" stays "2/4"); it only has to equal the child's answer.
func ParseNumber(s string) (string, *big.Rat, error) {
	s = strings.TrimSpace(s)
	if len(s) > MaxNumberLen {
		return "", nil, fmt.Errorf("%q is longer than %d characters", s, MaxNumberLen)
	}
	m := numberPattern.FindStringSubmatch(s)
	if m == nil {
		return "", nil, fmt.Errorf("%q is not a number (integer, decimal or a/b)", s)
	}
	sign, whole, frac, den := m[1], stripZeros(m[2]), strings.TrimRight(m[3], "0"), m[4]

	var canon string
	switch {
	case den != "":
		den = stripZeros(den)
		if den == "0" {
			return "", nil, fmt.Errorf("%q divides by zero", s)
		}
		canon = whole + "/" + den
	case frac != "":
		canon = whole + "." + frac
	default:
		canon = whole
	}
	r, ok := new(big.Rat).SetString(canon)
	if !ok { // cannot happen: canon matches what SetString reads
		return "", nil, fmt.Errorf("%q is not a number", s)
	}
	if sign == "-" && r.Sign() != 0 {
		canon = "-" + canon
		r.Neg(r)
	}
	return canon, r, nil
}

func stripZeros(digits string) string {
	if d := strings.TrimLeft(digits, "0"); d != "" {
		return d
	}
	return "0"
}

// NormalizeNumbers is the stored form of a NUMERIC_INPUT key: every value
// canonical (ParseNumber), in the order given. Two values equal as numbers
// ("0.5" and "1/2") are refused — grading compares numbers, so the second
// would accept nothing new.
func NormalizeNumbers(values []string) ([]string, error) {
	if len(values) < 1 || len(values) > MaxNumericValues {
		return nil, fmt.Errorf("%w: %s needs 1..%d values, got %d", ErrInvalidAnswer, KindNumericInput, MaxNumericValues, len(values))
	}
	out := make([]string, 0, len(values))
	seen := make([]*big.Rat, 0, len(values))
	for i, v := range values {
		canon, r, err := ParseNumber(v)
		if err != nil {
			return nil, fmt.Errorf("%w: values[%d] %v", ErrInvalidAnswer, i, err)
		}
		if j := slices.IndexFunc(seen, func(x *big.Rat) bool { return x.Cmp(r) == 0 }); j >= 0 {
			return nil, fmt.Errorf("%w: values[%d] %q equals values[%d]", ErrInvalidAnswer, i, v, j)
		}
		seen = append(seen, r)
		out = append(out, canon)
	}
	return out, nil
}

// AcceptsNumber is the grading rule of NUMERIC_INPUT: input is right when it
// equals one of values as a number ("1/2", "0,5" and "0.50" all match
// "0.5"). Anything that is not a number is wrong.
func AcceptsNumber(values []string, input string) bool {
	_, in, err := ParseNumber(input)
	if err != nil {
		return false
	}
	return slices.ContainsFunc(values, func(v string) bool {
		_, r, err := ParseNumber(v)
		return err == nil && r.Cmp(in) == 0
	})
}

// numericHashValues is a key as a set of numbers for Hash: each value
// reduced ("2/4", "0.5" → "1/2"), sorted, so neither order nor notation
// counts. Values that do not parse are kept as written; such a question
// never reaches the bank.
func numericHashValues(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if _, r, err := ParseNumber(v); err == nil {
			v = r.RatString()
		}
		out = append(out, v)
	}
	slices.Sort(out)
	return slices.Compact(out)
}
