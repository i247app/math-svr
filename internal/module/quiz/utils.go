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

func parseGeneratedQuestions(content string) ([]quizDto.QuizQuestion, error) {
	payload := extractJSONPayload(content)

	// Happy path: bare JSON array.
	var out []quizDto.QuizQuestion
	if err := json.Unmarshal([]byte(payload), &out); err == nil {
		if len(out) == 0 {
			return nil, errors.New("quiz: model returned zero questions")
		}
		return out, nil
	}

	// Some backends (notably OpenAI JSON mode) ignore the "return an
	// array" instruction and wrap it in an object — usually
	// {"questions": [...]} but occasionally "items" / "data". Try those
	// before assuming truncation.
	var wrap struct {
		Questions []quizDto.QuizQuestion `json:"questions"`
		Items     []quizDto.QuizQuestion `json:"items"`
		Data      []quizDto.QuizQuestion `json:"data"`
	}
	if err := json.Unmarshal([]byte(payload), &wrap); err == nil {
		switch {
		case len(wrap.Questions) > 0:
			return wrap.Questions, nil
		case len(wrap.Items) > 0:
			return wrap.Items, nil
		case len(wrap.Data) > 0:
			return wrap.Data, nil
		}
	}

	// LLM truncated mid-array (max_tokens hit, safety cut, network
	// reset, …). Salvage the longest prefix of complete objects rather
	// than fail the whole request.
	repaired, ok := salvageTruncatedJSONArray(payload)
	if !ok {
		return nil, fmt.Errorf("quiz: parse generated questions: payload not recoverable")
	}
	if err := json.Unmarshal([]byte(repaired), &out); err != nil {
		return nil, fmt.Errorf("quiz: parse generated questions (after salvage): %w", err)
	}
	if len(out) == 0 {
		return nil, errors.New("quiz: model returned zero questions")
	}
	return out, nil
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
