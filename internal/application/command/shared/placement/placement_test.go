package placement

import (
	"strings"
	"testing"
)

func TestBuildReview(t *testing.T) {
	tests := []struct {
		name string
		in   Input
		want []string
		not  []string
	}{
		{
			name: "nothing answered yet",
			in:   Input{},
			want: []string{"Chưa có dữ liệu"},
		},
		{
			name: "carries the counts",
			in:   Input{LifetimeTotal: 18, LifetimeCorrect: 12},
			want: []string{"18 câu", "đúng 12 câu", "66%"},
			not:  []string{"bỏ trống"},
		},
		{
			name: "names the skips only when there are any",
			in:   Input{LifetimeTotal: 5, LifetimeCorrect: 1, LifetimeSkipped: 3},
			want: []string{"bỏ trống 3 câu"},
		},
		{
			// The grade used to be spelled out here ("Đang ở mức Lớp 3").
			// It is client-stated now, not measured, so the review must
			// not present it as a finding.
			name: "never claims a placement",
			in:   Input{LifetimeTotal: 10, LifetimeCorrect: 10},
			not:  []string{"Đang ở mức", "Lớp", "Mẫu giáo"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Derive(tc.in).Review
			for _, w := range tc.want {
				if !strings.Contains(got, w) {
					t.Errorf("Review %q is missing %q", got, w)
				}
			}
			for _, n := range tc.not {
				if strings.Contains(got, n) {
					t.Errorf("Review %q must not contain %q", got, n)
				}
			}
		})
	}
}

func TestReviewIsBounded(t *testing.T) {
	got := Derive(Input{LifetimeTotal: 1 << 30, LifetimeCorrect: 1 << 29, LifetimeSkipped: 1 << 28}).Review
	if n := len([]rune(got)); n > reviewMaxLen {
		t.Fatalf("review is %d runes, cap is %d", n, reviewMaxLen)
	}
}

func TestPracticeReviewReadsTheSame(t *testing.T) {
	in := Input{LifetimeTotal: 4, LifetimeCorrect: 3}
	if Derive(in).Review != DerivePractice(in).Review {
		t.Fatal("practice and journey reviews diverged without a rule saying so")
	}
}
