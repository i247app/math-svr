package exam

import (
	"strings"
	"testing"

	"math-ai.com/math-ai/internal/shared/enum"
)

// TestCacheTagIsStable is the property the whole cache rests on. If two
// identical requests can produce two different tags, every lookup misses
// and the server quietly pays for a generation it already had — with no
// error, no log, and no way to notice except the bill.
func TestCacheTagIsStable(t *testing.T) {
	a := BuildCacheTag(enum.ExamTypeAssessment, 1, nil, 10, "Học kỳ 1", "Cánh diều")
	b := BuildCacheTag(enum.ExamTypeAssessment, 1, nil, 10, "Học kỳ 1", "Cánh diều")
	if a != b {
		t.Fatalf("same request produced two tags: %q vs %q", a, b)
	}
}

func TestCacheTagNormalisesInput(t *testing.T) {
	tests := []struct {
		name     string
		semester string
		program  string
	}{
		{"leading and trailing space", " Học kỳ 1 ", " Cánh diều "},
		{"inner double space", "Học  kỳ  1", "Cánh  diều"},
		{"different case", "học kỳ 1", "cánh diều"},
	}
	want := BuildCacheTag(enum.ExamTypePractice, 2, nil, 10, "Học kỳ 1", "Cánh diều")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildCacheTag(enum.ExamTypePractice, 2, nil, 10, tc.semester, tc.program)
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}

// TestCacheTagKeepsItsArity guards the separator. A program named with a
// hyphen must not be able to invent an extra field in the tag.
func TestCacheTagKeepsItsArity(t *testing.T) {
	plain := BuildCacheTag(enum.ExamTypePractice, 1, nil, 10, "HK1", "Canh dieu")
	hyphenated := BuildCacheTag(enum.ExamTypePractice, 1, nil, 10, "HK1", "Chan troi - sang tao")

	if got := strings.Count(plain, "-"); got != 5 {
		t.Errorf("plain tag %q has %d separators, want 5", plain, got)
	}
	if got := strings.Count(hyphenated, "-"); got != 5 {
		t.Errorf("hyphenated tag %q has %d separators, want 5", hyphenated, got)
	}
}

func TestCacheTagDistinguishesRequests(t *testing.T) {
	base := BuildCacheTag(enum.ExamTypeAssessment, 1, nil, 10, "HK1", "CD")
	tests := []struct {
		name string
		tag  string
	}{
		{"different type", BuildCacheTag(enum.ExamTypePractice, 1, nil, 10, "HK1", "CD")},
		{"different grade", BuildCacheTag(enum.ExamTypeAssessment, 2, nil, 10, "HK1", "CD")},
		{"different length", BuildCacheTag(enum.ExamTypeAssessment, 1, nil, 15, "HK1", "CD")},
		{"different semester", BuildCacheTag(enum.ExamTypeAssessment, 1, nil, 10, "HK2", "CD")},
		{"different program", BuildCacheTag(enum.ExamTypeAssessment, 1, nil, 10, "HK1", "KNTT")},
		{"a level", BuildCacheTag(enum.ExamTypeAssessment, 1, lvl(3), 10, "HK1", "CD")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.tag == base {
				t.Errorf("tag collides with the base request: %q", tc.tag)
			}
		})
	}
}

// TestCacheTagMissingParts pins the placeholders: an absent field renders
// as a marker rather than vanishing, so the tag keeps a fixed arity and
// stays readable by a human.
func TestCacheTagMissingParts(t *testing.T) {
	got := BuildCacheTag(enum.ExamTypeGrade, 0, nil, 10, "", "")
	want := "GRADE-G0-LNA-Q10-SNA-PNA"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func lvl(v int) *int { return &v }

// TestCacheTagSeparatesLevels: a level-3 paper and a level-9 paper at the
// same grade are different papers and must never share a pool.
func TestCacheTagSeparatesLevels(t *testing.T) {
	l3 := BuildCacheTag(enum.ExamTypeGrade, 2, lvl(3), 10, "HK1", "CD")
	l9 := BuildCacheTag(enum.ExamTypeGrade, 2, lvl(9), 10, "HK1", "CD")
	if l3 == l9 {
		t.Fatalf("levels 3 and 9 share tag %q", l3)
	}
	if want := "GRADE-G2-L3-Q10-SHK1-PCD"; l3 != want {
		t.Errorf("got %q, want %q", l3, want)
	}
}
