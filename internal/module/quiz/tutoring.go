package quiz

import (
	"context"
	"fmt"
	"strings"

	convcommand "math-ai.com/math-ai/internal/application/command/conversation"
	convquery "math-ai.com/math-ai/internal/application/query/conversation"
	conversationDomain "math-ai.com/math-ai/internal/domain/conversation"
	domain "math-ai.com/math-ai/internal/domain/quiz"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/shared/enum"
)

// maxTutoringLineChars caps each prior-context line forwarded to the LLM so
// a long ai_review can't blow up the prompt size.
const maxTutoringLineChars = 500

// loadTutoringContext reads the profile's recent QUIZ_TUTORING turns and
// renders them into a compact prompt block. Best-effort: any failure (or
// cold start) returns "" and the quiz is generated/graded without it.
func (s *Service) loadTutoringContext(ctx context.Context, profileID int64) string {
	if s.tutoringContextQuery == nil {
		return ""
	}
	msgs, err := s.tutoringContextQuery.Handle(ctx, convquery.GetTutoringContextQuery{
		ProfileID: profileID,
		Purpose:   conversationDomain.PurposeQuizTutoring,
		Limit:     s.tutoringWindow,
	})
	if err != nil {
		logger.From(ctx).Warnf("quiz.tutoring_context_failed profile_id=%d err=%v", profileID, err)
		return ""
	}
	if len(msgs) == 0 {
		return ""
	}
	lang := metadata.GetClientLanguage(ctx).ToEnumLanguage()
	return formatTutoringContext(msgs, lang)
}

// recordTutoringExchange appends the graded quiz's performance summary and
// AI review to the profile's tutoring thread. Best-effort: grading is
// already persisted, so a memory write failure is only logged.
func (s *Service) recordTutoringExchange(ctx context.Context, q *domain.Quiz, summary, review string) {
	if s.recordTutoringCmd == nil || q.UserId() == nil || q.ProfileId() == nil {
		return
	}
	if _, err := s.recordTutoringCmd.Handle(ctx, convcommand.RecordTutoringExchangeCommand{
		UserID:          *q.UserId(),
		ProfileID:       *q.ProfileId(),
		Purpose:         conversationDomain.PurposeQuizTutoring,
		UserSummary:     summary,
		AssistantReview: review,
	}); err != nil {
		logger.From(ctx).Warnf("quiz.tutoring_record_failed quiz_id=%d err=%v", q.QuizId(), err)
	}
}

// formatTutoringContext renders prior tutoring turns into a single system
// message body. Roles are flattened — only the content matters as context.
func formatTutoringContext(msgs []*conversationDomain.Message, lang enum.LanguageType) string {
	header := "Bối cảnh học tập gần đây của học sinh (dùng để cá nhân hoá đề bài và nhận xét; KHÔNG lặp lại nguyên văn):"
	if lang == enum.LanguageTypeEnglish {
		header = "Recent learning context for this student (use it to personalise the quiz and feedback; do NOT repeat verbatim):"
	}

	var b strings.Builder
	b.WriteString(header)
	for _, m := range msgs {
		line := strings.TrimSpace(m.Content())
		if line == "" {
			continue
		}
		if len([]rune(line)) > maxTutoringLineChars {
			line = string([]rune(line)[:maxTutoringLineChars]) + "…"
		}
		b.WriteString("\n- ")
		b.WriteString(line)
	}
	return b.String()
}

// buildPerformanceSummary turns a graded result into a one-line learning
// note (NOT the raw quiz JSON) stored as the user turn of the exchange.
// Source-agnostic so both the AI-graded (v1) and deterministic (v2) paths
// feed the tutoring thread the same way.
func buildPerformanceSummary(topic string, correct, total int, detectGrade string, lang enum.LanguageType) string {
	topic = strings.TrimSpace(topic)
	grade := strings.TrimSpace(detectGrade)

	if lang == enum.LanguageTypeEnglish {
		s := "Completed a quiz"
		if topic != "" {
			s += fmt.Sprintf(" on %q", topic)
		}
		s += fmt.Sprintf(": %d/%d correct.", correct, total)
		if grade != "" {
			s += fmt.Sprintf(" AI-estimated level: %s.", grade)
		}
		return s
	}

	s := "Đã làm quiz"
	if topic != "" {
		s += fmt.Sprintf(" chủ đề %q", topic)
	}
	s += fmt.Sprintf(": đúng %d/%d câu.", correct, total)
	if grade != "" {
		s += fmt.Sprintf(" AI nhận định trình độ: %s.", grade)
	}
	return s
}

// topicOf returns the quiz's short_text (preferred) or title.
func topicOf(q *domain.Quiz) string {
	if t := strings.TrimSpace(derefString(q.ShortText())); t != "" {
		return t
	}
	return strings.TrimSpace(derefString(q.Title()))
}

func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
