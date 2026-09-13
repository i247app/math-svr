package question

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Shuffle is one sitting's private ordering of a shared question set.
//
// A stored question set is shared: several children — or the same child
// twice — are served the same rows from cache. The set itself must stay
// canonical, because every analytic and every answer key is written
// against its numbering. What can vary is the ORDER a child meets the
// questions in, and the order of the options inside each. This type is
// that variation, recorded per sitting so it can be undone at grading.
//
// Two lists describe it, both in CANONICAL terms:
//
//   - Questions: the canonical question_number at each served position.
//     Served question 1 is Questions[0].
//   - Options: for each canonical question, the canonical labels in the
//     order they were shown. Served label "B" on that question is
//     Options[qn][1].
//
// Shuffling is done within difficulty groups only. An ASSESSMENT puts its
// probe questions at fixed positions on purpose (a harder question must
// not be the first thing a child sees); moving them would silently undo
// that design. So positions are partitioned by question_grade and each
// partition is permuted among its own slots.
//
// A nil *Shuffle is the identity: served order is stored order. Sittings
// recorded before shuffling existed carry NULL and behave that way.
type Shuffle struct {
	Version   int              `json:"v"`
	Questions []int            `json:"questions"`
	Options   map[int][]string `json:"options"`
}

// shuffleVersion is bumped if the meaning of the stored shape changes, so
// a sitting written under an old shape is never misread under a new one.
const shuffleVersion = 1

// Shuffler is the randomness source. *rand.Rand from math/rand/v2
// satisfies it, and a test can pass something deterministic.
type Shuffler interface {
	Shuffle(n int, swap func(i, j int))
}

// NewShuffle draws a fresh ordering for questions.
func NewShuffle(questions []Question, rng Shuffler) *Shuffle {
	s := &Shuffle{
		Version:   shuffleVersion,
		Questions: make([]int, len(questions)),
		Options:   make(map[int][]string, len(questions)),
	}

	// Positions grouped by grade; each group is permuted among its own
	// positions so the difficulty-by-position layout is preserved.
	groups := make(map[int][]int)
	var gradeKeys []int
	for i, q := range questions {
		g := 0
		if q.QuestionGrade != nil {
			g = *q.QuestionGrade
		}
		if _, seen := groups[g]; !seen {
			gradeKeys = append(gradeKeys, g)
		}
		groups[g] = append(groups[g], i)
		s.Questions[i] = q.QuestionNumber
	}
	sort.Ints(gradeKeys) // deterministic iteration so a seeded rng is reproducible
	for _, g := range gradeKeys {
		positions := groups[g]
		rng.Shuffle(len(positions), func(a, b int) {
			pa, pb := positions[a], positions[b]
			s.Questions[pa], s.Questions[pb] = s.Questions[pb], s.Questions[pa]
		})
	}

	for _, q := range questions {
		labels := make([]string, len(q.Answers))
		for i, a := range q.Answers {
			labels[i] = a.Label
		}
		rng.Shuffle(len(labels), func(a, b int) { labels[a], labels[b] = labels[b], labels[a] })
		s.Options[q.QuestionNumber] = labels
	}
	return s
}

// ParseShuffle decodes a stored ordering. A nil or blank input is the
// identity and returns (nil, nil).
func ParseShuffle(raw *string) (*Shuffle, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	var s Shuffle
	if err := json.Unmarshal([]byte(*raw), &s); err != nil {
		return nil, fmt.Errorf("question: parse shuffle: %w", err)
	}
	if s.Version != shuffleVersion {
		return nil, fmt.Errorf("question: shuffle version %d is not %d", s.Version, shuffleVersion)
	}
	return &s, nil
}

// JSON is the stored form.
func (s *Shuffle) JSON() (string, error) {
	raw, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("question: marshal shuffle: %w", err)
	}
	return string(raw), nil
}

// Apply returns the SERVED view of a canonical set: questions in served
// order and renumbered 1..N, options in served order and relabelled
// A..; right_answer_label is translated so it names the correct option
// under the served labels. The input is not modified.
//
// A nil receiver returns a copy in stored order.
func (s *Shuffle) Apply(canonical []Question) []Question {
	byNumber := make(map[int]Question, len(canonical))
	for _, q := range canonical {
		byNumber[q.QuestionNumber] = q
	}

	var order []int
	if s == nil {
		order = make([]int, 0, len(canonical))
		for _, q := range canonical {
			order = append(order, q.QuestionNumber)
		}
	} else {
		order = s.Questions
	}

	served := make([]Question, 0, len(order))
	for pos, qn := range order {
		q, ok := byNumber[qn]
		if !ok {
			continue // the stored set no longer holds this question; skip rather than invent
		}
		q.QuestionNumber = pos + 1
		q.Answers, q.RightAnswerLabel = s.servedOptions(qn, q)
		served = append(served, q)
	}
	return served
}

// ServedAnswers is one question's options in the order — and under the
// labels — this sitting showed them. The review screen uses it to lay out
// every option beside the one the child picked; a nil receiver returns
// the stored order.
func (s *Shuffle) ServedAnswers(canonicalQN int, q Question) []AnswerChoice {
	answers, _ := s.servedOptions(canonicalQN, q)
	return answers
}

// servedOptions reorders one question's options and translates the
// answer key into served labels.
func (s *Shuffle) servedOptions(canonicalQN int, q Question) ([]AnswerChoice, string) {
	byLabel := make(map[string]AnswerChoice, len(q.Answers))
	for _, a := range q.Answers {
		byLabel[a.Label] = a
	}

	var canonicalOrder []string
	if s != nil {
		canonicalOrder = s.Options[canonicalQN]
	}
	if len(canonicalOrder) != len(q.Answers) {
		// No usable ordering recorded — serve as stored.
		return q.Answers, q.RightAnswerLabel
	}

	out := make([]AnswerChoice, 0, len(canonicalOrder))
	rightServed := q.RightAnswerLabel
	for i, canonicalLabel := range canonicalOrder {
		a, ok := byLabel[canonicalLabel]
		if !ok {
			return q.Answers, q.RightAnswerLabel
		}
		servedLabel := labelAt(i)
		if canonicalLabel == q.RightAnswerLabel {
			rightServed = servedLabel
		}
		out = append(out, AnswerChoice{Label: servedLabel, Content: a.Content})
	}
	return out, rightServed
}

// ToCanonical translates the answers a child gave under the served
// numbering back into canonical terms, so grading runs against the
// stored set. A nil receiver passes answers through unchanged.
//
// An answer naming a served position or label that does not exist is an
// error, not a silent drop: the client is speaking a numbering the server
// handed it moments ago, so a mismatch means a bug, not a skipped question.
func (s *Shuffle) ToCanonical(served []StudentAnswer) ([]StudentAnswer, error) {
	if s == nil {
		return served, nil
	}
	out := make([]StudentAnswer, 0, len(served))
	for _, a := range served {
		if a.QuestionNumber < 1 || a.QuestionNumber > len(s.Questions) {
			return nil, fmt.Errorf("question: served question %d is outside 1..%d", a.QuestionNumber, len(s.Questions))
		}
		canonicalQN := s.Questions[a.QuestionNumber-1]

		idx, ok := labelIndex(a.Label)
		labels := s.Options[canonicalQN]
		if !ok || idx >= len(labels) {
			return nil, fmt.Errorf("question: served label %q is not an option of question %d", a.Label, a.QuestionNumber)
		}
		out = append(out, StudentAnswer{QuestionNumber: canonicalQN, Label: labels[idx]})
	}
	return out, nil
}

// ServedNumber is the position a canonical question was shown at, or 0
// when it was not shown. Identity on a nil receiver.
func (s *Shuffle) ServedNumber(canonicalQN int) int {
	if s == nil {
		return canonicalQN
	}
	for pos, qn := range s.Questions {
		if qn == canonicalQN {
			return pos + 1
		}
	}
	return 0
}

// ServedLabel is the label a canonical option was shown under, or the
// canonical label itself when no ordering applies.
func (s *Shuffle) ServedLabel(canonicalQN int, canonicalLabel string) string {
	if s == nil {
		return canonicalLabel
	}
	for i, l := range s.Options[canonicalQN] {
		if l == canonicalLabel {
			return labelAt(i)
		}
	}
	return canonicalLabel
}

// labelAt / labelIndex map option positions to the A.. alphabet the
// product uses. Sets have four options today; the alphabet is not capped
// so a longer set still labels cleanly.
func labelAt(i int) string { return string(rune('A' + i)) }

func labelIndex(label string) (int, bool) {
	if len(label) != 1 || label[0] < 'A' || label[0] > 'Z' {
		return 0, false
	}
	return int(label[0] - 'A'), true
}
