package quiz

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	quizDto "math-ai.com/math-ai/internal/application/dto/quiz"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
)

// validQuestionTypes is the closed render set the client understands.
// Anything else (including empty) normalizes to ARITHMETIC.
var validQuestionTypes = map[string]struct{}{
	quizDto.QuestionTypeArithmetic:    {},
	quizDto.QuestionTypeCount:         {},
	quizDto.QuestionTypePickByIcon:    {},
	quizDto.QuestionTypeIdentifyShape: {},
}

// geometryIconWhitelist is the closed set of shape tokens the client ships
// SVG assets for. The prompt tells the model to stay inside this set;
// normalization only LOGS drift — it never drops a question, because a
// stray token still renders as a client-side fallback and the MCQ
// structure (labels + right_answer) stays valid and gradable.
var geometryIconWhitelist = map[string]struct{}{
	"triangle": {}, "square": {}, "rectangle": {}, "circle": {},
	"star": {}, "diamond": {}, "oval": {}, "pentagon": {},
	"hexagon": {}, "heart": {},
}

// iconTokenRe matches the "[icon:NAME]" shape tokens embedded in a stem or
// answer content. Emoji are literal UTF-8 and are intentionally not matched.
var iconTokenRe = regexp.MustCompile(`\[icon:([a-zA-Z_]+)\]`)

// normalizeGeneratedQuestions clamps each question's QuestionType to the
// known set (unknown/empty -> ARITHMETIC) and logs any [icon:NAME] token
// outside the geometry whitelist. It mutates in place and returns the
// slice. Visual drift never fails generation — grading is label-based and
// unaffected by icon content.
func normalizeGeneratedQuestions(ctx context.Context, questions []quizDto.QuizQuestion) []quizDto.QuizQuestion {
	log := logger.From(ctx)

	warnTokens := func(s string, qnum int) {
		for _, m := range iconTokenRe.FindAllStringSubmatch(s, -1) {
			if _, ok := geometryIconWhitelist[strings.ToLower(m[1])]; !ok {
				log.Warnf("quiz.normalize.unknown_icon_token token=%q q=%d", m[1], qnum)
			}
		}
	}

	for i := range questions {
		q := &questions[i]
		qt := strings.ToUpper(strings.TrimSpace(q.QuestionType))
		if _, ok := validQuestionTypes[qt]; !ok {
			if qt != "" {
				log.Warnf("quiz.normalize.unknown_question_type type=%q q=%d -> ARITHMETIC", q.QuestionType, q.QuestionNumber)
			}
			qt = quizDto.QuestionTypeArithmetic
		}
		q.QuestionType = qt

		warnTokens(q.QuestionName, q.QuestionNumber)
		for _, a := range q.Answers {
			warnTokens(a.Content, q.QuestionNumber)
		}
	}
	return questions
}

// NormalizeGeneratedQuestions is the exported entry point so sibling
// modules that share the generation JSON contract (e.g. classroomexercise)
// apply the same question_type clamp + icon-token validation as quizzes.
func NormalizeGeneratedQuestions(ctx context.Context, questions []quizDto.QuizQuestion) []quizDto.QuizQuestion {
	return normalizeGeneratedQuestions(ctx, questions)
}

func normalizeLanguage(lang enum.LanguageType) string {
	s := strings.ToLower(strings.TrimSpace(string(lang)))
	if s == "" {
		return string(enum.LanguageTypeEnglish)
	}
	return s
}

// parseGeneration extracts the AI-generated quiz title + short_text +
// questions from the LLM response. The prompt schema is
// `{"title": "...", "short_text": "...", "questions": [...]}` so the
// happy path is the object wrapper; the bare-array and truncation-salvage
// branches are kept as defence-in-depth for backends that drift from the
// schema or hit a max-tokens cut. title (grade/level label) and
// short_text (topic description) are both best-effort: an empty string
// flows through to a NULL DB column, which the response layer omits.
// ParseGeneration is the exported entry point for sibling modules that
// drive the same generation JSON contract (e.g. classroomexercise). It
// reuses the same payload-extraction + salvage pipeline as the in-
// package parseGeneration so any future schema tweak only has to be
// made in one place.
func ParseGeneration(content string) (title string, shortText string, questions []quizDto.QuizQuestion, err error) {
	t, st, _, q, err := parseGeneration(content)
	return t, st, q, err
}

func parseGeneration(content string) (title string, shortText string, assessmentGrade string, questions []quizDto.QuizQuestion, err error) {
	payload := extractJSONPayload(content)

	var wrap struct {
		Title           string                 `json:"title"`
		ShortText       string                 `json:"short_text"`
		AssessmentGrade string                 `json:"assessment_grade"`
		Questions       []quizDto.QuizQuestion `json:"questions"`
		Items           []quizDto.QuizQuestion `json:"items"`
		Data            []quizDto.QuizQuestion `json:"data"`
		Quiz            []quizDto.QuizQuestion `json:"quiz"`
	}
	if err := json.Unmarshal([]byte(payload), &wrap); err == nil {
		t := strings.TrimSpace(wrap.Title)
		st := strings.TrimSpace(wrap.ShortText)
		ag := strings.TrimSpace(wrap.AssessmentGrade)
		switch {
		case len(wrap.Questions) > 0:
			return t, st, ag, wrap.Questions, nil
		case len(wrap.Items) > 0:
			return t, st, ag, wrap.Items, nil
		case len(wrap.Data) > 0:
			return t, st, ag, wrap.Data, nil
		case len(wrap.Quiz) > 0:
			return t, st, ag, wrap.Quiz, nil
		}
	}

	// Some backends drop the wrapper and return a bare array; title,
	// short_text and assessment_grade are unavailable on this branch but
	// the questions are still usable.
	var out []quizDto.QuizQuestion
	if err := json.Unmarshal([]byte(payload), &out); err == nil && len(out) > 0 {
		return "", "", "", out, nil
	}

	// LLM truncated mid-array (max_tokens hit, safety cut, network
	// reset, …). title, short_text and assessment_grade appear before the
	// questions array in our schema, so we can usually recover them via
	// regex even when the array body is truncated.
	t := extractStringFieldFromTruncated(payload, titleFieldRe)
	st := extractStringFieldFromTruncated(payload, shortTextFieldRe)
	ag := extractStringFieldFromTruncated(payload, assessmentGradeFieldRe)
	arr := extractQuestionsArrayPrefix(payload)
	if arr == "" {
		return "", "", "", nil, fmt.Errorf("quiz: parse generated questions: payload not recoverable")
	}
	repaired, ok := salvageTruncatedJSONArray(arr)
	if !ok {
		return "", "", "", nil, fmt.Errorf("quiz: parse generated questions: payload not recoverable")
	}
	if err := json.Unmarshal([]byte(repaired), &out); err != nil {
		return "", "", "", nil, fmt.Errorf("quiz: parse generated questions (after salvage): %w", err)
	}
	if len(out) == 0 {
		return "", "", "", nil, ErrQuizModelReturnedZeroQuestions
	}
	return t, st, ag, out, nil
}

// titleFieldRe / shortTextFieldRe find the first top-level `"title": "..."`
// and `"short_text": "..."` pairs. Used as truncation-tolerant fallbacks
// for parseGeneration.
var (
	titleFieldRe           = regexp.MustCompile(`"title"\s*:\s*"((?:[^"\\]|\\.)*)"`)
	shortTextFieldRe       = regexp.MustCompile(`"short_text"\s*:\s*"((?:[^"\\]|\\.)*)"`)
	assessmentGradeFieldRe = regexp.MustCompile(`"assessment_grade"\s*:\s*"((?:[^"\\]|\\.)*)"`)
)

func extractStringFieldFromTruncated(payload string, re *regexp.Regexp) string {
	m := re.FindStringSubmatch(payload)
	if len(m) < 2 {
		return ""
	}
	// Re-quote so json.Unmarshal handles escape sequences for us.
	var unq string
	if err := json.Unmarshal([]byte(`"`+m[1]+`"`), &unq); err == nil {
		return strings.TrimSpace(unq)
	}
	return strings.TrimSpace(m[1])
}

// questionsArrayKeyRe locates the start of the questions array inside an
// object payload so the existing array salvage can run on the substring.
var questionsArrayKeyRe = regexp.MustCompile(`"(questions|items|data|quiz)"\s*:\s*\[`)

func extractQuestionsArrayPrefix(payload string) string {
	loc := questionsArrayKeyRe.FindStringIndex(payload)
	if loc == nil {
		return ""
	}
	// loc[1] is one past the matched `[`; back up so the slice starts at it.
	return payload[loc[1]-1:]
}

// salvageTruncatedJSONArray recovers the longest prefix of a JSON array
// of objects that ends at a complete top-level element. It returns
// (repaired, true) when at least one full object was found; the repaired
// string is guaranteed to be a syntactically-valid JSON array.
//
// The state machine tracks string/escape state so braces inside strings
// don't affect depth, and remembers the index of the last "}" that
// closed back to depth 0 (just inside the leading "["). On truncation
// that index is the cut point — everything after gets dropped and the
// array is re-closed with "]". Payloads that don't start with "["
// (some backends drop the wrapper) are wrapped before salvage is
// attempted.
func salvageTruncatedJSONArray(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	if s[0] != '[' {
		if s[0] != '{' {
			return "", false
		}
		s = "[" + s
	}

	inStr := false
	esc := false
	depth := 0
	lastGood := -1

	for i := 1; i < len(s); i++ {
		c := s[i]
		if inStr {
			if esc {
				esc = false
				continue
			}
			switch c {
			case '\\':
				esc = true
			case '"':
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				lastGood = i
			}
		case ']':
			if depth == 0 {
				return s[:i+1], true
			}
		}
	}

	if lastGood < 0 {
		return "", false
	}
	return s[:lastGood+1] + "]", true
}

func parseGradedQuiz(content string) (*quizDto.QuizGradingResult, error) {
	payload := extractJSONPayload(content)
	var out quizDto.QuizGradingResult
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		return nil, fmt.Errorf("quiz: parse graded quiz: %w", err)
	}
	return &out, nil
}

// ParseGradedQuiz is the exported wrapper around parseGradedQuiz. The
// exercise submission module reuses it so the grading-response shape
// stays unified across quiz and exercise flows.
func ParseGradedQuiz(content string) (*quizDto.QuizGradingResult, error) {
	return parseGradedQuiz(content)
}

// codeFenceRe matches ```lang\n...\n``` wrappers some backends emit even
// when JSON mode is requested. Strip them so json.Unmarshal sees the
// payload directly.
var codeFenceRe = regexp.MustCompile("(?ms)^```[a-zA-Z0-9_-]*\\n?(.*?)\\n?```$")

func extractJSONPayload(s string) string {
	s = strings.TrimSpace(s)
	if m := codeFenceRe.FindStringSubmatch(s); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}
	return s
}

// maxQuizTextLen mirrors the VARCHAR(255) limit on ma_quizzes.title and
// ma_quizzes.short_text. The prompt asks the model for <= 80 characters;
// the clamp here is a defensive backstop against drift, not a primary
// enforcement point.
const maxQuizTextLen = 255

// sanitizeQuizText trims whitespace and clamps an AI-generated text field
// (title or short_text) to the DB column's rune budget. Returns nil when
// nothing usable is left so the row stores a real NULL (which
// DomainToResponse then omits).
func sanitizeQuizText(text string) *string {
	t := strings.TrimSpace(text)
	if t == "" {
		return nil
	}
	if runes := []rune(t); len(runes) > maxQuizTextLen {
		t = string(runes[:maxQuizTextLen])
	}
	return &t
}
