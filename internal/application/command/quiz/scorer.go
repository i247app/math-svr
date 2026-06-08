package command

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	quizDto "math-ai.com/math-ai/internal/application/dto/quiz"
	"math-ai.com/math-ai/internal/shared/enum"
)

// ReviewSourceDeterministicV2 is the audit marker prefixed onto every
// ai_review produced by the deterministic submit path. It lets ops grep
// the column to attribute a review to v1 (bot) vs v2 (server) without a
// schema change. The mobile client renders it inline; the column is
// VARCHAR(255) so the prefix is part of the budget the builder respects.
const ReviewSourceDeterministicV2 = "[v2]"

// reviewMaxLen mirrors ma_quizzes.ai_review's VARCHAR(255) cap. We trim
// slightly under the column width so a future schema bump (e.g. adding a
// UTF-8 multibyte tail) cannot truncate mid-grapheme on write.
const reviewMaxLen = 250

// ScoreResult is the deterministic counterpart of QuizGradingResult. It
// carries the same shape the bot grader returns so the command/handler
// layer can persist v1 and v2 outputs through the same code path.
type ScoreResult struct {
	TotalQuestions  int
	CorrectNumber   int
	ScorePercentage int
	AIReview        string
	// AIDetectGrade is intentionally left nil by the deterministic
	// scorer today — see design note §3. Reserved for a follow-up that
	// derives a coarse grade signal from per-question difficulty.
	AIDetectGrade *string
}

// ScoreQuiz grades the student's answers against the quiz's questions
// payload entirely in process. No bot call, no I/O. lang controls the
// language the review string is built in; VN is the default to match
// errs.NewError's hardcode.
//
// questionsJSON is the row's `questions` column verbatim. answers is the
// parsed student payload from the v2 submit request.
//
// Returns a ScoreResult shaped to slot into quiz.GradingUpdate.
func ScoreQuiz(questionsJSON string, answers []quizDto.QuizStudentAnswer, lang enum.LanguageType) (*ScoreResult, error) {
	questions, err := parseQuestionsForScoring(questionsJSON)
	if err != nil {
		return nil, err
	}
	if len(questions) == 0 {
		return nil, fmt.Errorf("scorer: quiz has no questions")
	}

	answerByNumber, err := indexStudentAnswers(answers)
	if err != nil {
		return nil, err
	}

	total := len(questions)
	correct := 0

	// Per-topic accuracy table. When a question carries no topic tag
	// (legacy generation), it is bucketed under "" and the review builder
	// degrades to a generic wording.
	topics := make(map[string]*topicBucket, total)

	for _, q := range questions {
		topicKey := strings.TrimSpace(strings.ToLower(q.Topic))
		b, ok := topics[topicKey]
		if !ok {
			b = &topicBucket{}
			topics[topicKey] = b
		}
		b.total++

		studentLabel, present := answerByNumber[q.QuestionNumber]
		if !present {
			// Skipped — counts as incorrect, matching v1's
			// "round(correct/total*100)" semantics. No partial credit.
			continue
		}
		if questionIsCorrect(q, studentLabel) {
			correct++
			b.correct++
		}
	}

	percentage := 0
	if total > 0 {
		percentage = int((float64(correct)/float64(total))*100 + 0.5)
	}

	review := buildDeterministicReview(reviewInputs{
		Total:    total,
		Correct:  correct,
		Topics:   topics,
		Language: lang,
	})

	return &ScoreResult{
		TotalQuestions:  total,
		CorrectNumber:   correct,
		ScorePercentage: percentage,
		AIReview:        review,
		AIDetectGrade:   nil,
	}, nil
}

// parseQuestionsForScoring decodes the row's `questions` LONGTEXT into the
// DTO shape. Returns an error rather than nil so the caller can map it to
// QUIZ_GRADING_FAILED — empty payloads are the v1 path's existing
// "quiz has no questions to grade" condition.
func parseQuestionsForScoring(raw string) ([]quizDto.QuizQuestion, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("scorer: questions payload is empty")
	}
	var out []quizDto.QuizQuestion
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("scorer: parse questions: %w", err)
	}
	return out, nil
}

// indexStudentAnswers folds the answer slice into a map keyed by
// question_number. Duplicate numbers fail loudly — that's malformed
// client payload, distinct from a missing answer (which is allowed and
// scored as incorrect).
func indexStudentAnswers(answers []quizDto.QuizStudentAnswer) (map[int]string, error) {
	out := make(map[int]string, len(answers))
	for i, a := range answers {
		if _, dup := out[a.QuestionNumber]; dup {
			return nil, fmt.Errorf("scorer: duplicate answer for question_number %d at index %d", a.QuestionNumber, i)
		}
		out[a.QuestionNumber] = strings.TrimSpace(a.Label)
	}
	return out, nil
}

// questionIsCorrect implements the design note's matching rules:
// label match first (the v1-compatible path), value compare as fallback
// only when right_answer is absent.
func questionIsCorrect(q quizDto.QuizQuestion, studentLabel string) bool {
	rightLabel := strings.ToUpper(strings.TrimSpace(q.RightAnswer))
	if rightLabel != "" {
		return strings.ToUpper(studentLabel) == rightLabel
	}
	if strings.TrimSpace(q.CorrectAnswer) == "" {
		return false
	}
	// Value-compare path: locate the chosen choice's content, compare it
	// to correct_answer via the canonical numeric/string compare.
	for _, c := range q.Answers {
		if strings.EqualFold(strings.TrimSpace(c.Label), studentLabel) {
			return valuesEqual(c.Content, q.CorrectAnswer)
		}
	}
	return false
}

// valuesEqual implements the design note's canonical compare:
//   - trim + whitespace-collapse both sides
//   - convert Vietnamese decimal comma to dot when both look numeric
//   - float compare with absolute tolerance 1e-9
//   - ASCII fraction "a/b" cross-multiplication compare when both match
//   - fall back to case-insensitive string equality
func valuesEqual(a, b string) bool {
	a = collapseWS(a)
	b = collapseWS(b)
	if a == "" && b == "" {
		return true
	}
	if a == "" || b == "" {
		return false
	}

	// Strip leading + and a single trailing . from each side before the
	// numeric branches so "+8" and "8." compare equal to "8".
	aNum := normalizeNumericString(a)
	bNum := normalizeNumericString(b)

	// Fraction compare. Both must parse cleanly as `int/int` (no decimals
	// in either operand) for this branch — otherwise fall through to the
	// float branch which handles mixed cases like "0.5" vs "1/2".
	if aN, aD, ok := parseAsciiFraction(aNum); ok {
		if bN, bD, okB := parseAsciiFraction(bNum); okB {
			return aN*bD == bN*aD
		}
	}

	if af, ok := parseFloatVN(aNum); ok {
		if bf, okB := parseFloatVN(bNum); okB {
			diff := af - bf
			if diff < 0 {
				diff = -diff
			}
			return diff < 1e-9
		}
	}

	return strings.EqualFold(a, b)
}

func collapseWS(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		b.WriteRune(r)
		prevSpace = false
	}
	return b.String()
}

func normalizeNumericString(s string) string {
	s = strings.TrimPrefix(s, "+")
	if strings.HasSuffix(s, ".") && len(s) > 1 {
		s = s[:len(s)-1]
	}
	return s
}

// parseFloatVN parses a number that may use either '.' or ',' as the
// decimal separator. Returns ok=false for any value that doesn't look
// purely numeric.
func parseFloatVN(s string) (float64, bool) {
	if s == "" {
		return 0, false
	}
	// VN convention writes "1,5" for 1.5. Accept the comma form only when
	// there's exactly one comma and no dot — mixed punctuation looks like
	// a thousands separator and we leave it alone.
	if strings.Count(s, ",") == 1 && !strings.Contains(s, ".") {
		s = strings.Replace(s, ",", ".", 1)
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// parseAsciiFraction parses "a/b" with integer numerator and denominator.
// Denominator must be non-zero. Negative sign on either side is honoured.
func parseAsciiFraction(s string) (int64, int64, bool) {
	idx := strings.Index(s, "/")
	if idx <= 0 || idx == len(s)-1 {
		return 0, 0, false
	}
	num, err := strconv.ParseInt(strings.TrimSpace(s[:idx]), 10, 64)
	if err != nil {
		return 0, 0, false
	}
	den, err := strconv.ParseInt(strings.TrimSpace(s[idx+1:]), 10, 64)
	if err != nil || den == 0 {
		return 0, 0, false
	}
	return num, den, true
}

// topicBucket counts correct vs total questions per topic tag. Used by
// both ScoreQuiz (bucketing pass) and rankTopics (review-builder pass).
type topicBucket struct {
	correct int
	total   int
}

type reviewInputs struct {
	Total    int
	Correct  int
	Topics   map[string]*topicBucket
	Language enum.LanguageType
}

// buildDeterministicReview is split out so it can be exercised
// independently in tests. It always emits a string within reviewMaxLen
// runes and always carries the ReviewSourceDeterministicV2 prefix.
func buildDeterministicReview(in reviewInputs) string {
	en := in.Language == enum.LanguageTypeEnglish

	weakTopics, strongTopic := rankTopics(in.Topics)

	var body string
	switch {
	case len(weakTopics) == 0 && strongTopic == "":
		// No topic tags at all — degrade to a generic summary.
		body = genericSummary(en, in.Correct, in.Total)
	case len(weakTopics) == 0:
		body = strongOnlySummary(en, strongTopic, in.Correct, in.Total)
	case strongTopic == "":
		body = weakOnlySummary(en, weakTopics, in.Correct, in.Total)
	default:
		body = fullSummary(en, strongTopic, weakTopics, in.Correct, in.Total)
	}

	// out := ReviewSourceDeterministicV2 + " " + body
	out := body
	if runes := []rune(out); len(runes) > reviewMaxLen {
		out = string(runes[:reviewMaxLen-1]) + "…"
	}
	return out
}

// rankTopics returns up to two weakest topics (accuracy ascending, tie
// broken by topic name for determinism) and the single strongest topic.
// The empty-string bucket ("no topic tag") is skipped — it doesn't
// represent a real skill, it's the legacy-schema fallback.
func rankTopics(topics map[string]*topicBucket) (weakest []string, strongest string) {
	type ranked struct {
		topic    string
		accuracy float64
		total    int
	}
	all := make([]ranked, 0, len(topics))
	for name, b := range topics {
		if name == "" || b.total == 0 {
			continue
		}
		all = append(all, ranked{
			topic:    name,
			accuracy: float64(b.correct) / float64(b.total),
			total:    b.total,
		})
	}
	if len(all) == 0 {
		return nil, ""
	}

	sort.SliceStable(all, func(i, j int) bool {
		if all[i].accuracy != all[j].accuracy {
			return all[i].accuracy < all[j].accuracy
		}
		return all[i].topic < all[j].topic
	})

	// Strongest = the highest-accuracy topic that isn't a 0% bucket; the
	// weakest list is anything below 100% to surface real friction. Cap
	// at two so the review string stays short.
	for _, r := range all {
		if r.accuracy < 1.0 && len(weakest) < 2 {
			weakest = append(weakest, r.topic)
		}
	}
	strongest = all[len(all)-1].topic
	// If the "strongest" is also one of the weak entries (everything is
	// imperfect), don't list it as a strength.
	if slices.Contains(weakest, strongest) {
		strongest = ""
	}
	return weakest, strongest
}

func genericSummary(en bool, correct, total int) string {
	wrong := total - correct
	if en {
		return fmt.Sprintf("Got %d/%d correct. Keep practicing the missed items.", correct, total)
	}
	return fmt.Sprintf("Đúng %d/%d câu (sai %d). Cần luyện thêm phần các câu chưa đúng.", correct, total, wrong)
}

func strongOnlySummary(en bool, strong string, correct, total int) string {
	if en {
		return fmt.Sprintf("Strength: %s. %d/%d correct overall.", strong, correct, total)
	}
	return fmt.Sprintf("Điểm tốt: %s. Đúng %d/%d câu.", strong, correct, total)
}

func weakOnlySummary(en bool, weak []string, correct, total int) string {
	weakList := strings.Join(weak, ", ")
	wrong := total - correct
	if en {
		return fmt.Sprintf("Practice: %s (%d/%d wrong).", weakList, wrong, total)
	}
	return fmt.Sprintf("Cần luyện: %s (%d/%d sai).", weakList, wrong, total)
}

func fullSummary(en bool, strong string, weak []string, correct, total int) string {
	weakList := strings.Join(weak, ", ")
	wrong := total - correct
	if en {
		return fmt.Sprintf("Strength: %s. Practice: %s (%d/%d wrong).", strong, weakList, wrong, total)
	}
	return fmt.Sprintf("Điểm tốt: %s. Cần luyện: %s (%d/%d sai).", strong, weakList, wrong, total)
}
