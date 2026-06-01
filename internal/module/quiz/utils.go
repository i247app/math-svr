package quiz

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	quizDto "math-ai.com/math-ai/internal/application/dto/quiz"
	"math-ai.com/math-ai/internal/shared/enum"
)

func normalizeLanguage(lang enum.LanguageType) string {
	s := strings.ToLower(strings.TrimSpace(string(lang)))
	if s == "" {
		return string(enum.LanguageTypeEnglish)
	}
	return s
}

// parseGeneration extracts the AI-generated quiz title + questions from
// the LLM response. The prompt schema is
// `{"title": "...", "questions": [...]}` so the happy path is the object
// wrapper; the bare-array and truncation-salvage branches are kept as
// defence-in-depth for backends that drift from the schema or hit a
// max-tokens cut. Title is best-effort: an empty string flows through
// to a NULL DB column, which the response layer omits.
// ParseGeneration is the exported entry point for sibling modules that
// drive the same generation JSON contract (e.g. classroomexercise). It
// reuses the same payload-extraction + salvage pipeline as the in-
// package parseGeneration so any future schema tweak only has to be
// made in one place.
func ParseGeneration(content string) (string, []quizDto.QuizQuestion, error) {
	return parseGeneration(content)
}

func parseGeneration(content string) (string, []quizDto.QuizQuestion, error) {
	payload := extractJSONPayload(content)

	var wrap struct {
		Title     string                 `json:"title"`
		Questions []quizDto.QuizQuestion `json:"questions"`
		Items     []quizDto.QuizQuestion `json:"items"`
		Data      []quizDto.QuizQuestion `json:"data"`
		Quiz      []quizDto.QuizQuestion `json:"quiz"`
	}
	if err := json.Unmarshal([]byte(payload), &wrap); err == nil {
		switch {
		case len(wrap.Questions) > 0:
			return strings.TrimSpace(wrap.Title), wrap.Questions, nil
		case len(wrap.Items) > 0:
			return strings.TrimSpace(wrap.Title), wrap.Items, nil
		case len(wrap.Data) > 0:
			return strings.TrimSpace(wrap.Title), wrap.Data, nil
		case len(wrap.Quiz) > 0:
			return strings.TrimSpace(wrap.Title), wrap.Quiz, nil
		}
	}

	// Some backends drop the wrapper and return a bare array; title is
	// unavailable on this branch but the questions are still usable.
	var out []quizDto.QuizQuestion
	if err := json.Unmarshal([]byte(payload), &out); err == nil && len(out) > 0 {
		return "", out, nil
	}

	// LLM truncated mid-array (max_tokens hit, safety cut, network
	// reset, …). The title appears before the questions array in our
	// schema, so we can usually recover it via regex even when the
	// array body is truncated.
	title := extractTitleFromTruncated(payload)
	arr := extractQuestionsArrayPrefix(payload)
	if arr == "" {
		return "", nil, fmt.Errorf("quiz: parse generated questions: payload not recoverable")
	}
	repaired, ok := salvageTruncatedJSONArray(arr)
	if !ok {
		return "", nil, fmt.Errorf("quiz: parse generated questions: payload not recoverable")
	}
	if err := json.Unmarshal([]byte(repaired), &out); err != nil {
		return "", nil, fmt.Errorf("quiz: parse generated questions (after salvage): %w", err)
	}
	if len(out) == 0 {
		return "", nil, errors.New("quiz: model returned zero questions")
	}
	return title, out, nil
}

// titleFieldRe finds the first top-level `"title": "..."` pair. Used as a
// truncation-tolerant fallback for parseGeneration.
var titleFieldRe = regexp.MustCompile(`"title"\s*:\s*"((?:[^"\\]|\\.)*)"`)

func extractTitleFromTruncated(payload string) string {
	m := titleFieldRe.FindStringSubmatch(payload)
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

// maxQuizTitleLen mirrors the VARCHAR(255) limit on ma_quizzes.title.
// The prompt asks the model for <= 80 characters; the clamp here is a
// defensive backstop against drift, not a primary enforcement point.
const maxQuizTitleLen = 255

// sanitizeQuizTitle trims whitespace and clamps the title to the DB
// column's rune budget. Returns nil when nothing usable is left so the
// row stores a real NULL (which DomainToResponse then omits).
func sanitizeQuizTitle(title string) *string {
	t := strings.TrimSpace(title)
	if t == "" {
		return nil
	}
	if runes := []rune(t); len(runes) > maxQuizTitleLen {
		t = string(runes[:maxQuizTitleLen])
	}
	return &t
}
