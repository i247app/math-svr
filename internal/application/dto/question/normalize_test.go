package question

import "testing"

// TestNormalizeQuestionType: the render set is closed, but the spellings
// the prompt's own example teaches must land inside it rather than fall
// to ARITHMETIC with a warning.
func TestNormalizeQuestionType(t *testing.T) {
	tests := []struct {
		in       string
		want     string
		warnKind string
	}{
		{"COUNT", TypeCount, ""},
		{"counting", TypeCount, ""},
		{" Counting ", TypeCount, ""},
		{"ARITHMETIC", TypeArithmetic, ""},
		{"", TypeArithmetic, ""},
		{"SEQUENCE", TypeArithmetic, WarnUnknownType},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			qs, warnings := Normalize([]Question{{QuestionNumber: 1, QuestionType: tc.in}})
			if got := qs[0].QuestionType; got != tc.want {
				t.Errorf("Normalize(%q) type = %q, want %q", tc.in, got, tc.want)
			}
			if tc.warnKind == "" && len(warnings) != 0 {
				t.Errorf("Normalize(%q) warned %v, want none", tc.in, warnings)
			}
			if tc.warnKind != "" && (len(warnings) != 1 || warnings[0].Kind != tc.warnKind) {
				t.Errorf("Normalize(%q) warnings = %v, want one %s", tc.in, warnings, tc.warnKind)
			}
		})
	}
}
