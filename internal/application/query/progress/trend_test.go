package progress

import (
	"math"
	"testing"

	"math-ai.com/math-ai/internal/shared/enum"
)

func TestPctTo10Pt(t *testing.T) {
	cases := []struct {
		pct  float64
		want float64
	}{
		{0, 0}, {60, 6}, {77, 7.7}, {87.5, 8.75}, {100, 10},
	}
	for _, c := range cases {
		if got := PctTo10Pt(c.pct); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("PctTo10Pt(%v)=%v want %v", c.pct, got, c.want)
		}
	}
}

func TestLinearSlope(t *testing.T) {
	if got := LinearSlope([]float64{1, 2, 3}); math.Abs(got-1) > 1e-9 {
		t.Errorf("rising slope = %v want 1", got)
	}
	if got := LinearSlope([]float64{5}); got != 0 {
		t.Errorf("single point slope = %v want 0", got)
	}
	if got := LinearSlope([]float64{3, 2, 1}); math.Abs(got+1) > 1e-9 {
		t.Errorf("falling slope = %v want -1", got)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		count      int
		avg, slope float64
		want       enum.ProgressComment
	}{
		{0, 0, 0, enum.ProgressCommentNoData},
		{1, 9, 1, enum.ProgressCommentInsufficient},
		{5, 4.9, 1, enum.ProgressCommentNeedToTry},
		{5, 8, -0.1, enum.ProgressCommentNeedToTry},
		{5, 8, 0.2, enum.ProgressCommentGoodProgress},
		{5, 6, 0.1, enum.ProgressCommentProgress},
	}
	for _, c := range cases {
		if got := Classify(c.count, c.avg, c.slope); got != c.want {
			t.Errorf("Classify(%d,%v,%v)=%v want %v", c.count, c.avg, c.slope, got, c.want)
		}
	}
}
