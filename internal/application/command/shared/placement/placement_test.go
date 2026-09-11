package placement

import (
	"strings"
	"testing"
)

func intPtr(n int) *int { return &n }

func TestDeriveGrade(t *testing.T) {
	tests := []struct {
		name string
		in   Input
		want int
	}{
		{
			name: "first submission above threshold moves up from the exam's grade",
			in:   Input{ExamGrade: 1, LastScorePercentage: 80},
			want: 2,
		},
		{
			name: "first submission below threshold moves down",
			in:   Input{ExamGrade: 3, LastScorePercentage: 20},
			want: 2,
		},
		{
			name: "exactly at the threshold holds",
			in:   Input{ExamGrade: 3, LastScorePercentage: PromotionThreshold},
			want: 3,
		},
		{
			name: "an existing placement wins over the exam's grade",
			in:   Input{CurrentGrade: intPtr(4), ExamGrade: 1, LastScorePercentage: 90},
			want: 5,
		},
		{
			name: "cannot be promoted past the top band",
			in:   Input{CurrentGrade: intPtr(5), ExamGrade: 5, LastScorePercentage: 100},
			want: 5,
		},
		{
			name: "cannot be demoted below kindergarten",
			in:   Input{CurrentGrade: intPtr(0), ExamGrade: 0, LastScorePercentage: 0},
			want: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Derive(tc.in)
			if got.Grade == nil {
				t.Fatal("Grade is nil")
			}
			if *got.Grade != tc.want {
				t.Errorf("Grade = %d, want %d", *got.Grade, tc.want)
			}
		})
	}
}

// TestDeriveGradeIgnoresLifetime pins the N = 1 window. A child with a
// long, weak history who has just done well must still move up — a
// lifetime average would freeze them in place, which is the failure this
// rule exists to avoid.
func TestDeriveGradeIgnoresLifetime(t *testing.T) {
	got := Derive(Input{
		CurrentGrade:        intPtr(2),
		ExamGrade:           2,
		LastScorePercentage: 100,
		LifetimeTotal:       500,
		LifetimeCorrect:     50, // 10% lifetime accuracy
	})
	if *got.Grade != 3 {
		t.Errorf("Grade = %d, want 3 — the last attempt decides, not the lifetime average", *got.Grade)
	}
}

func TestDeriveLevel(t *testing.T) {
	tests := []struct {
		name string
		in   Input
		want int
	}{
		{"no history starts at the bottom", Input{}, levelStart},
		{"weak lifetime accuracy", Input{LifetimeTotal: 20, LifetimeCorrect: 4}, levelStruggling},
		{"at the steady floor", Input{LifetimeTotal: 20, LifetimeCorrect: 10}, levelSteady},
		{"at the strong floor stays steady", Input{LifetimeTotal: 20, LifetimeCorrect: 16}, levelSteady},
		{"above the strong floor", Input{LifetimeTotal: 20, LifetimeCorrect: 19}, levelStrong},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Derive(tc.in)
			if got.Level == nil {
				t.Fatal("Level is nil")
			}
			if *got.Level != tc.want {
				t.Errorf("Level = %d, want %d", *got.Level, tc.want)
			}
		})
	}
}

func TestBuildReview(t *testing.T) {
	t.Run("no history says so plainly", func(t *testing.T) {
		got := Derive(Input{}).Review
		if !strings.Contains(got, "Chưa có dữ liệu") {
			t.Errorf("Review = %q", got)
		}
	})

	t.Run("carries counts and the current placement", func(t *testing.T) {
		got := Derive(Input{
			CurrentGrade: intPtr(2), ExamGrade: 2, LastScorePercentage: 80,
			LifetimeTotal: 18, LifetimeCorrect: 12,
		}).Review
		for _, want := range []string{"18 câu", "12 câu", "66%", "Lớp 3"} {
			if !strings.Contains(got, want) {
				t.Errorf("Review %q is missing %q", got, want)
			}
		}
	})

	t.Run("mentions blanks only when there are some", func(t *testing.T) {
		with := Derive(Input{ExamGrade: 1, LifetimeTotal: 10, LifetimeCorrect: 5, LifetimeSkipped: 4}).Review
		if !strings.Contains(with, "bỏ trống 4 câu") {
			t.Errorf("Review %q should mention the blanks", with)
		}
		without := Derive(Input{ExamGrade: 1, LifetimeTotal: 10, LifetimeCorrect: 5}).Review
		if strings.Contains(without, "bỏ trống") {
			t.Errorf("Review %q should not mention blanks when there are none", without)
		}
	})

	t.Run("kindergarten is named, not numbered", func(t *testing.T) {
		got := Derive(Input{CurrentGrade: intPtr(0), ExamGrade: 0, LastScorePercentage: 10,
			LifetimeTotal: 5, LifetimeCorrect: 1}).Review
		if !strings.Contains(got, "Mẫu giáo") {
			t.Errorf("Review = %q, want it to name kindergarten", got)
		}
	})
}
