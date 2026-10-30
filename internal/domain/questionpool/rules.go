package questionpool

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// Shape limits of a choice question.
const (
	MinOptions = 2
	MaxOptions = 10
)

// Ranges of question_grade (0 = Mẫu giáo, 1..5 = Lớp 1..5, 6 = Lớp 6 — the
// product stops at grade 5, but a grade-5 exam's probe questions sit one
// grade above) and question_level / question_sub_level (0 = easiest .. 9 =
// hardest).
const (
	GradeMin    = 0
	GradeMax    = 6
	LevelMin    = 0
	LevelMax    = 9
	SubLevelMin = 0
	SubLevelMax = 9
)

// The three ways a question's answer shape can be wrong. CheckShape wraps
// one of them, so a caller can tell them apart with errors.Is.
var (
	ErrInvalidKind    = errors.New("invalid question_kind")
	ErrInvalidOptions = errors.New("invalid options")
	ErrInvalidAnswer  = errors.New("invalid correct_answer")
)

var optionIdPattern = regexp.MustCompile(`^[a-z0-9_]{1,16}$`)

// CheckShape enforces the per-kind rules of migration 036: what options a
// kind carries and what its key looks like. It does not trim or reorder;
// options are checked as given. Lengths and image keys are the caller's.
func CheckShape(kind Kind, options []Option, answer CorrectAnswer) error {
	if !kind.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidKind, kind)
	}
	if !kind.HasOptions() {
		if options != nil {
			return fmt.Errorf("%w: %s takes no options", ErrInvalidOptions, kind)
		}
		if kind == KindNumericInput {
			if answer.Value != nil || answer.OptionIds != nil {
				return fmt.Errorf("%w: %s needs values (numbers as strings) and no value or option_ids", ErrInvalidAnswer, kind)
			}
			_, err := NormalizeNumbers(answer.Values)
			return err
		}
		// TRUE_FALSE
		if answer.Value == nil || answer.OptionIds != nil || answer.Values != nil {
			return fmt.Errorf("%w: %s needs value (a boolean) and no option_ids or values", ErrInvalidAnswer, kind)
		}
		return nil
	}

	if len(options) < MinOptions || len(options) > MaxOptions {
		return fmt.Errorf("%w: %s needs %d..%d options, got %d", ErrInvalidOptions, kind, MinOptions, MaxOptions, len(options))
	}
	ids := make([]string, 0, len(options))
	for i, o := range options {
		if !optionIdPattern.MatchString(o.Id) {
			return fmt.Errorf("%w: options[%d].id %q must be 1-16 of a-z 0-9 _", ErrInvalidOptions, i, o.Id)
		}
		if slices.Contains(ids, o.Id) {
			return fmt.Errorf("%w: options[%d].id %q is repeated", ErrInvalidOptions, i, o.Id)
		}
		if strings.TrimSpace(o.Text) == "" && o.ImageKey == nil {
			return fmt.Errorf("%w: options[%d] has neither text nor image_key", ErrInvalidOptions, i)
		}
		ids = append(ids, o.Id)
	}

	if answer.Value != nil || answer.Values != nil {
		return fmt.Errorf("%w: %s takes option_ids, not value or values", ErrInvalidAnswer, kind)
	}
	switch n := len(answer.OptionIds); {
	case kind == KindSingleChoice && n != 1:
		return fmt.Errorf("%w: %s needs exactly one option id, got %d", ErrInvalidAnswer, kind, n)
	case kind == KindMultipleChoice && n < 1:
		return fmt.Errorf("%w: %s needs at least one option id", ErrInvalidAnswer, kind)
	}
	var seen []string
	for i, id := range answer.OptionIds {
		if !slices.Contains(ids, id) {
			return fmt.Errorf("%w: option_ids[%d] %q is not an option", ErrInvalidAnswer, i, id)
		}
		if slices.Contains(seen, id) {
			return fmt.Errorf("%w: option_ids[%d] %q is repeated", ErrInvalidAnswer, i, id)
		}
		seen = append(seen, id)
	}
	return nil
}

// NormalizeText is the form of question / option text that counts for
// duplicates: Unicode NFC (a Vietnamese letter typed precomposed or with
// combining marks is the same letter), trimmed, inner whitespace collapsed.
func NormalizeText(s string) string {
	return strings.Join(strings.Fields(norm.NFC.String(s)), " ")
}

// hashInput is what Hash fingerprints; field order is fixed by the struct.
type hashInput struct {
	Kind     Kind         `json:"k"`
	Grade    int          `json:"g"`
	Text     string       `json:"t"`
	ImageKey *string      `json:"i"`
	Options  []hashOption `json:"o"`
	Value    *bool        `json:"v"`
	Numbers  []string     `json:"n,omitempty"` // omitempty: the hashes of the other kinds stay as they were
}

// hashOption is one option by CONTENT: its id is left out, so the same
// question written with ids a/b or x/y hashes the same.
type hashOption struct {
	Text     string `json:"t"`
	ImageKey string `json:"i"`
	Correct  bool   `json:"c"`
}

// Hash is the question_hash of q (migration 036): SHA-256 hex of its kind,
// grade, normalised text, stem image, its options as a set (normalised text
// + image key + whether correct), for TRUE_FALSE its value and, for
// NUMERIC_INPUT, its values as a set of numbers ("0.5" = "1/2"). Option ids
// and display order, topic, tags, level, sub level and explanation do not count — the
// same question shown in another order or filed under another topic is
// still the same question.
func Hash(q *QuestionPool) string {
	in := hashInput{
		Kind:     q.QuestionKind(),
		Grade:    q.QuestionGrade(),
		Text:     NormalizeText(q.QuestionText()),
		ImageKey: q.QuestionImageKey(),
		Value:    q.CorrectAnswer().Value,
	}
	if q.QuestionKind() == KindNumericInput {
		in.Numbers = numericHashValues(q.CorrectAnswer().Values)
	}
	for _, o := range q.Options() {
		h := hashOption{Text: NormalizeText(o.Text), Correct: slices.Contains(q.CorrectAnswer().OptionIds, o.Id)}
		if o.ImageKey != nil {
			h.ImageKey = *o.ImageKey
		}
		in.Options = append(in.Options, h)
	}
	slices.SortFunc(in.Options, func(a, b hashOption) int {
		if c := strings.Compare(a.Text, b.Text); c != 0 {
			return c
		}
		return strings.Compare(a.ImageKey, b.ImageKey)
	})
	raw, _ := json.Marshal(in) // only strings, ints, bools: cannot fail
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
