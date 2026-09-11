package question

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrNoQuestions reports a payload that parsed but carried no questions.
// It is a distinct error because "the model answered with an empty round"
// and "the model's answer was unreadable" call for different log lines.
var ErrNoQuestions = errors.New("question: model returned zero questions")

// Generation is the parsed result of a generation call. Title, ShortText
// and AssessmentGrade are best-effort: an empty string flows through to a
// NULL column, which the response layer then omits. Callers must tolerate
// them being blank without aborting the round.
type Generation struct {
	Title           string
	ShortText       string
	AssessmentGrade string
	Questions       []Question
}

// GenerationOf is Generation for a caller that spells its questions
// differently. The exam flow does: its wire vocabulary matches its own
// columns (right_answer_label, question_topic, ...) rather than the
// quiz-era names this package's Question carries.
type GenerationOf[T any] struct {
	Title           string
	ShortText       string
	AssessmentGrade string
	Questions       []T
}

// ParseGeneration extracts a generated round in this package's own
// question shape.
func ParseGeneration(content string) (*Generation, error) {
	g, err := ParseGenerationOf[Question](content)
	if err != nil {
		return nil, err
	}
	return &Generation{
		Title:           g.Title,
		ShortText:       g.ShortText,
		AssessmentGrade: g.AssessmentGrade,
		Questions:       g.Questions,
	}, nil
}

// ParseGenerationOf is the engine: same wrapper detection, same
// bare-array fallback, same truncation salvage, for any question type.
// It is generic so a second vocabulary costs a type, not a second copy of
// the salvage state machine — the part most likely to drift if duplicated.
//
// The prompt schema is {"title", "short_text", "assessment_grade",
// "questions": [...]} so the object wrapper is the happy path; the
// bare-array and truncation-salvage branches are defence in depth against
// backends that drift from the schema or hit a max-tokens cut.
func ParseGenerationOf[T any](content string) (*GenerationOf[T], error) {
	payload := extractJSONPayload(content)

	var wrap struct {
		Title           string `json:"title"`
		ShortText       string `json:"short_text"`
		AssessmentGrade string `json:"assessment_grade"`
		Questions       []T    `json:"questions"`
		Items           []T    `json:"items"`
		Data            []T    `json:"data"`
		Quiz            []T    `json:"quiz"`
	}
	if err := json.Unmarshal([]byte(payload), &wrap); err == nil {
		out := &GenerationOf[T]{
			Title:           strings.TrimSpace(wrap.Title),
			ShortText:       strings.TrimSpace(wrap.ShortText),
			AssessmentGrade: strings.TrimSpace(wrap.AssessmentGrade),
		}
		switch {
		case len(wrap.Questions) > 0:
			out.Questions = wrap.Questions
			return out, nil
		case len(wrap.Items) > 0:
			out.Questions = wrap.Items
			return out, nil
		case len(wrap.Data) > 0:
			out.Questions = wrap.Data
			return out, nil
		case len(wrap.Quiz) > 0:
			out.Questions = wrap.Quiz
			return out, nil
		}
	}

	// Some backends drop the wrapper and return a bare array; the header
	// fields are unavailable on this branch but the questions are usable.
	var bare []T
	if err := json.Unmarshal([]byte(payload), &bare); err == nil && len(bare) > 0 {
		return &GenerationOf[T]{Questions: bare}, nil
	}

	// Truncated mid-array (max_tokens hit, safety cut, network reset...).
	// The header fields appear before the questions array in the schema,
	// so they can usually be recovered by regex even when the array body
	// is cut off.
	out := &GenerationOf[T]{
		Title:           extractStringFieldFromTruncated(payload, titleFieldRe),
		ShortText:       extractStringFieldFromTruncated(payload, shortTextFieldRe),
		AssessmentGrade: extractStringFieldFromTruncated(payload, assessmentGradeFieldRe),
	}
	arr := extractQuestionsArrayPrefix(payload)
	if arr == "" {
		return nil, fmt.Errorf("question: parse generation: payload not recoverable")
	}
	repaired, ok := salvageTruncatedJSONArray(arr)
	if !ok {
		return nil, fmt.Errorf("question: parse generation: payload not recoverable")
	}
	if err := json.Unmarshal([]byte(repaired), &out.Questions); err != nil {
		return nil, fmt.Errorf("question: parse generation (after salvage): %w", err)
	}
	if len(out.Questions) == 0 {
		return nil, ErrNoQuestions
	}
	return out, nil
}

// ParseGrading decodes a grading response into its typed shape.
func ParseGrading(content string) (*GradingResult, error) {
	payload := extractJSONPayload(content)
	var out GradingResult
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		return nil, fmt.Errorf("question: parse grading: %w", err)
	}
	return &out, nil
}

// titleFieldRe / shortTextFieldRe / assessmentGradeFieldRe find the first
// top-level occurrence of their string field. Truncation-tolerant
// fallbacks for ParseGeneration.
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
// object payload so the array salvage can run on the substring.
var questionsArrayKeyRe = regexp.MustCompile(`"(questions|items|data|quiz)"\s*:\s*\[`)

func extractQuestionsArrayPrefix(payload string) string {
	loc := questionsArrayKeyRe.FindStringIndex(payload)
	if loc == nil {
		return ""
	}
	// loc[1] is one past the matched "["; back up so the slice starts at it.
	return payload[loc[1]-1:]
}

// salvageTruncatedJSONArray recovers the longest prefix of a JSON array of
// objects that ends at a complete top-level element. It returns
// (repaired, true) when at least one full object was found; the repaired
// string is guaranteed to be a syntactically-valid JSON array.
//
// The state machine tracks string/escape state so braces inside strings
// don't affect depth, and remembers the index of the last "}" that closed
// back to depth 0 (just inside the leading "["). On truncation that index
// is the cut point — everything after it is dropped and the array is
// re-closed with "]". Payloads that don't start with "[" (some backends
// drop the wrapper) are wrapped before salvage is attempted.
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

// codeFenceRe matches the ```lang ... ``` wrapper some backends emit even
// when JSON mode is requested. Stripping it lets json.Unmarshal see the
// payload directly.
var codeFenceRe = regexp.MustCompile("(?ms)^```[a-zA-Z0-9_-]*\\n?(.*?)\\n?```$")

func extractJSONPayload(s string) string {
	s = strings.TrimSpace(s)
	if m := codeFenceRe.FindStringSubmatch(s); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}
	return s
}
