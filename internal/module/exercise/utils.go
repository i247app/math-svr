package exercise

import (
	"encoding/json"
	"strings"

	quizDto "math-ai.com/math-ai/internal/application/dto/quiz"
	"math-ai.com/math-ai/internal/shared/enum"
)

func normalizeLanguage(lang enum.LanguageType) string {
	s := strings.ToLower(strings.TrimSpace(string(lang)))
	if s == "" {
		return string(enum.LanguageTypeVietnamese)
	}
	return s
}

// buildAnswerKey extracts {question_number: right_answer_label} from the
// generated questions and serialises it. Stored separately so a future
// submission grader can compare against this column without having to
// re-parse the full questions blob.
func buildAnswerKey(questions []quizDto.QuizQuestion) []byte {
	key := make(map[string]string, len(questions))
	for _, q := range questions {
		if q.RightAnswer == "" {
			continue
		}
		key[itoa(q.QuestionNumber)] = q.RightAnswer
	}
	if len(key) == 0 {
		return []byte("{}")
	}
	out, err := json.Marshal(key)
	if err != nil {
		return []byte("{}")
	}
	return out
}

// itoa avoids the strconv import dance for a single use site. The
// quiz dto already validated that QuestionNumber is a non-negative int.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
