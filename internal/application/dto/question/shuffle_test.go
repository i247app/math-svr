package question

import (
	"math/rand/v2"
	"testing"
)

func grade(g int) *int { return &g }

// fixture: ten questions, probes at 3 and 6 one grade up, four options
// each with the right answer always "A" canonically.
func fixture() []Question {
	out := make([]Question, 0, 10)
	for i := 1; i <= 10; i++ {
		g := 1
		if i == 3 || i == 6 {
			g = 2
		}
		out = append(out, Question{
			QuestionNumber:   i,
			QuestionName:     "q" + string(rune('0'+i)),
			Answers:          []AnswerChoice{{"A", "right"}, {"B", "w1"}, {"C", "w2"}, {"D", "w3"}},
			RightAnswerLabel: "A",
			QuestionGrade:    grade(g),
		})
	}
	return out
}

func seeded(seed uint64) *rand.Rand { return rand.New(rand.NewPCG(seed, seed)) }

// TestShufflePreservesDifficultyLayout is the invariant the whole design
// hangs on: the probe slots of an ASSESSMENT stay where they were put.
func TestShufflePreservesDifficultyLayout(t *testing.T) {
	qs := fixture()
	for seed := uint64(1); seed <= 50; seed++ {
		s := NewShuffle(qs, seeded(seed))
		served := s.Apply(qs)

		if len(served) != 10 {
			t.Fatalf("seed %d: served %d questions, want 10", seed, len(served))
		}
		for pos, q := range served {
			want := 1
			if pos == 2 || pos == 5 { // served positions 3 and 6
				want = 2
			}
			if q.QuestionGrade == nil || *q.QuestionGrade != want {
				t.Fatalf("seed %d: served position %d has grade %v, want %d", seed, pos+1, q.QuestionGrade, want)
			}
		}
	}
}

func TestShuffleIsAPermutation(t *testing.T) {
	qs := fixture()
	s := NewShuffle(qs, seeded(7))

	seen := make(map[int]bool)
	for _, qn := range s.Questions {
		if seen[qn] {
			t.Fatalf("question %d appears twice in served order", qn)
		}
		seen[qn] = true
	}
	if len(seen) != 10 {
		t.Fatalf("served order names %d distinct questions, want 10", len(seen))
	}

	for qn, labels := range s.Options {
		if len(labels) != 4 {
			t.Errorf("question %d has %d option labels, want 4", qn, len(labels))
		}
	}
}

// TestShuffleActuallyShuffles guards against a no-op: over many draws the
// served order must differ from stored order at least sometimes.
func TestShuffleActuallyShuffles(t *testing.T) {
	qs := fixture()
	moved := 0
	for seed := uint64(1); seed <= 20; seed++ {
		s := NewShuffle(qs, seeded(seed))
		for pos, qn := range s.Questions {
			if qn != pos+1 {
				moved++
				break
			}
		}
	}
	if moved == 0 {
		t.Fatal("served order never differed from stored order across 20 draws")
	}
}

// TestShuffleRoundTrip is the correctness contract: whatever label the
// child picks under the served view, translating it back must land on
// the right canonical option — and grading against the canonical key
// must agree with what the child saw.
func TestShuffleRoundTrip(t *testing.T) {
	qs := fixture()
	s := NewShuffle(qs, seeded(42))
	served := s.Apply(qs)

	// The child answers every served question with the label that the
	// served view marks as correct.
	answers := make([]StudentAnswer, 0, len(served))
	for _, q := range served {
		answers = append(answers, StudentAnswer{QuestionNumber: q.QuestionNumber, Label: q.RightAnswerLabel})
	}

	canonical, err := s.ToCanonical(answers)
	if err != nil {
		t.Fatalf("ToCanonical: %v", err)
	}
	if len(canonical) != 10 {
		t.Fatalf("got %d canonical answers, want 10", len(canonical))
	}
	for _, a := range canonical {
		// Canonically the right answer is always "A".
		if a.Label != "A" {
			t.Errorf("canonical question %d: translated label %q, want A", a.QuestionNumber, a.Label)
		}
	}
}

func TestShuffleServedContentFollowsOptions(t *testing.T) {
	qs := fixture()
	s := NewShuffle(qs, seeded(3))
	served := s.Apply(qs)

	for _, q := range served {
		// Whatever label is marked right, its content must be the right one.
		var content string
		for _, a := range q.Answers {
			if a.Label == q.RightAnswerLabel {
				content = a.Content
			}
		}
		if content != "right" {
			t.Errorf("served question %d marks %q as right but its content is %q", q.QuestionNumber, q.RightAnswerLabel, content)
		}
		// Labels are always A, B, C, D in served order.
		for i, a := range q.Answers {
			if a.Label != labelAt(i) {
				t.Errorf("served question %d option %d labelled %q, want %q", q.QuestionNumber, i, a.Label, labelAt(i))
			}
		}
	}
}

func TestShuffleRejectsUnknownServedAnswers(t *testing.T) {
	s := NewShuffle(fixture(), seeded(1))

	if _, err := s.ToCanonical([]StudentAnswer{{QuestionNumber: 11, Label: "A"}}); err == nil {
		t.Error("served question 11 of 10 should be rejected")
	}
	if _, err := s.ToCanonical([]StudentAnswer{{QuestionNumber: 1, Label: "E"}}); err == nil {
		t.Error("served label E on a four-option question should be rejected")
	}
	if _, err := s.ToCanonical([]StudentAnswer{{QuestionNumber: 1, Label: "x"}}); err == nil {
		t.Error("lower-case / non-alphabet label should be rejected")
	}
}

// TestNilShuffleIsIdentity covers sittings recorded before shuffling
// existed: they carry no ordering and must behave exactly as before.
func TestNilShuffleIsIdentity(t *testing.T) {
	qs := fixture()
	var s *Shuffle

	served := s.Apply(qs)
	for i, q := range served {
		if q.QuestionNumber != qs[i].QuestionNumber || q.RightAnswerLabel != "A" {
			t.Fatalf("nil shuffle changed question %d", i+1)
		}
	}

	in := []StudentAnswer{{QuestionNumber: 4, Label: "C"}}
	out, err := s.ToCanonical(in)
	if err != nil || len(out) != 1 || out[0] != in[0] {
		t.Fatalf("nil shuffle altered answers: %v %v", out, err)
	}
	if s.ServedNumber(4) != 4 || s.ServedLabel(4, "C") != "C" {
		t.Fatal("nil shuffle should map numbers and labels to themselves")
	}

	parsed, err := ParseShuffle(nil)
	if err != nil || parsed != nil {
		t.Fatalf("ParseShuffle(nil) = %v, %v; want nil, nil", parsed, err)
	}
}

func TestShuffleJSONRoundTrip(t *testing.T) {
	s := NewShuffle(fixture(), seeded(9))
	raw, err := s.JSON()
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	back, err := ParseShuffle(&raw)
	if err != nil {
		t.Fatalf("ParseShuffle: %v", err)
	}
	for i := range s.Questions {
		if back.Questions[i] != s.Questions[i] {
			t.Fatalf("served order changed through JSON at %d", i)
		}
	}
	for qn, labels := range s.Options {
		for i := range labels {
			if back.Options[qn][i] != labels[i] {
				t.Fatalf("option order for question %d changed through JSON", qn)
			}
		}
	}
}

func TestShuffleServedLookups(t *testing.T) {
	s := NewShuffle(fixture(), seeded(5))
	served := s.Apply(fixture())

	for pos, q := range served {
		canonicalQN := s.Questions[pos]
		if got := s.ServedNumber(canonicalQN); got != pos+1 {
			t.Errorf("ServedNumber(%d) = %d, want %d", canonicalQN, got, pos+1)
		}
		// The canonical right answer "A" must be reported under the label
		// the served view shows it as.
		if got := s.ServedLabel(canonicalQN, "A"); got != q.RightAnswerLabel {
			t.Errorf("ServedLabel(%d, A) = %q, want %q", canonicalQN, got, q.RightAnswerLabel)
		}
	}
}
