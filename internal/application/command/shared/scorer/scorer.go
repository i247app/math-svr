// Package scorer is the deterministic, bot-free grader for MCQ-style
// quizzes and classroom exercises. Both aggregates ship the same
// `{question_number, question_name, answers:[{label, content}],
// right_answer, correct_answer?, topic?, difficulty?}` schema, so a
// single scoring engine serves both submit/v2 endpoints:
//
//   - /quizzes/submit/v2          → quiz module
//   - /classroom-exercise/submissions/submit/v2 → exercise module
//
// The package lives under application/command/shared so the two command
// packages can both depend on it without depending on each other (which
// would be a layer-smell cross-aggregate dep at the application layer).
//
// What it does NOT do:
//   - persistence (no UoW, no repos);
//   - request validation (the per-module validator catches duplicate
//     question_numbers and missing labels before Score is called);
//   - free-form / multi-select / ordering question types (none are
//     emitted by the live prompts; the design note next to the v2
//     endpoints documents the constraint).
package scorer

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

// ReviewSourceMarker is the deterministic-path identifier kept here for
// callers that want to audit/tag a v2-graded row separately from a bot
// row (e.g. log lines, future analytics columns). The Score function
// does NOT prefix the generated review with it — by design, the
// persisted review is plain prose so the mobile client renders identical
// shapes whether v1 (bot) or v2 (server) produced it.
const ReviewSourceMarker = "deterministic_v2"

// reviewMaxLen mirrors the tightest column ai_review may land in —
// today ma_quizzes.ai_review's VARCHAR(255). ma_exercise_submissions
// uses LONGTEXT so it has plenty of headroom; clamping both paths to the
// same budget keeps the review compact and consistent across aggregates.
const reviewMaxLen = 250

// Result is the deterministic counterpart of quizDto.QuizGradingResult.
// Both quiz and exercise commands map this into their respective
// row-update / row-insert types (quiz.GradingUpdate for quiz,
// SubmitExerciseAnswersV2Command's per-column fields for exercise).
type Result struct {
	TotalQuestions  int
	CorrectNumber   int
	ScorePercentage int
	AIReview        string
	// AssessmentGrade stays nil today — only quiz writes this column, and
	// deriving a coarse grade signal from difficulty is reserved for a
	// follow-up once topic+difficulty tags appear in real traffic.
	AssessmentGrade *string
}

// Score grades the student's answers against the questions payload
// entirely in process. No bot call, no I/O. lang controls the language
// the review string is built in.
//
// questionsJSON is the row's `questions` column verbatim. answers is
// the parsed student payload from the v2 submit request.
//
// Returns a Result shaped to slot into either aggregate's grading-write
// path.
func Score(questionsJSON string, answers []quizDto.QuizStudentAnswer, lang enum.LanguageType) (*Result, error) {
	questions, err := parseQuestionsForScoring(questionsJSON)
	if err != nil {
		return nil, err
	}
	if len(questions) == 0 {
		return nil, fmt.Errorf("scorer: payload has no questions")
	}

	answerByNumber, err := indexStudentAnswers(answers)
	if err != nil {
		return nil, err
	}

	total := len(questions)
	correct := 0

	// Per-topic accuracy table. Questions with no topic tag (legacy
	// generation) bucket under "" and the review builder degrades to a
	// generic wording.
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
			// Skipped — counts as incorrect. Matches both the quiz and
			// exercise bot-grade prompts: missing answer ⇒ wrong.
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

	return &Result{
		TotalQuestions:  total,
		CorrectNumber:   correct,
		ScorePercentage: percentage,
		AIReview:        review,
		AssessmentGrade: nil,
	}, nil
}

// parseQuestionsForScoring decodes the row's `questions` LONGTEXT into
// the DTO shape. Returns an error on empty / malformed input so the
// caller can surface the right status code (QUIZ_GRADING_FAILED /
// CLASSROOM_EXERCISE_SUBMISSION_GRADING_FAILED).
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
// question_number. Duplicate numbers fail loudly — the per-module
// validator catches it earlier, but the scorer guards too so out-of-band
// callers can't smuggle them past.
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

// questionIsCorrect: label match first (the v1-compatible path), value
// compare as fallback only when right_answer is absent.
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

// valuesEqual: canonical compare. Order of checks:
//   - trim + whitespace-collapse both sides
//   - convert Vietnamese decimal comma to dot when both look numeric
//   - ASCII fraction "a/b" cross-multiplication compare when both match
//   - float compare with absolute tolerance 1e-9
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

	aNum := normalizeNumericString(a)
	bNum := normalizeNumericString(b)

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
// decimal separator. Returns ok=false when the value doesn't look purely
// numeric. The VN form is accepted only when there's exactly one comma
// and no dot — anything else looks like a thousands separator and we
// leave it alone.
func parseFloatVN(s string) (float64, bool) {
	if s == "" {
		return 0, false
	}
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
// Denominator must be non-zero. Negative sign on either side is honoured
// by strconv.ParseInt.
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
// both Score (bucketing pass) and rankTopics (review-builder pass).
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

// buildDeterministicReview emits a plain-prose review under reviewMaxLen
// runes. No source marker is prefixed; auditing is done via the caller's
// structured log line (see ReviewSourceMarker docstring).
func buildDeterministicReview(in reviewInputs) string {
	en := in.Language == enum.LanguageTypeEnglish

	weakTopics, strongTopic := rankTopics(in.Topics)

	var body string
	switch {
	case len(weakTopics) == 0 && strongTopic == "":
		body = genericSummary(en, in.Correct, in.Total)
	case len(weakTopics) == 0:
		body = strongOnlySummary(en, strongTopic, in.Correct, in.Total)
	case strongTopic == "":
		body = weakOnlySummary(en, weakTopics, in.Correct, in.Total)
	default:
		body = fullSummary(en, strongTopic, weakTopics, in.Correct, in.Total)
	}

	if runes := []rune(body); len(runes) > reviewMaxLen {
		body = string(runes[:reviewMaxLen-1]) + "…"
	}
	return body
}

// rankTopics: up to two weakest topics (accuracy ascending, name as tie
// breaker for determinism), single strongest topic. Empty-string bucket
// ("no topic tag") is skipped — it isn't a real skill.
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

	for _, r := range all {
		if r.accuracy < 1.0 && len(weakest) < 2 {
			weakest = append(weakest, r.topic)
		}
	}
	strongest = all[len(all)-1].topic
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
