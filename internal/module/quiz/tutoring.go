package quiz

import (
	"context"
	"fmt"
	"strings"

	convcommand "math-ai.com/math-ai/internal/application/command/conversation"
	appconv "math-ai.com/math-ai/internal/application/conversation"
	conversationDomain "math-ai.com/math-ai/internal/domain/conversation"
	domain "math-ai.com/math-ai/internal/domain/quiz"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/libs/langchain"
	"math-ai.com/math-ai/internal/shared/enum"
)

// loadTutoringContext reads the profile's recent QUIZ_TUTORING turns via the
// framework memory and returns them as a prompt-ready string. Best-effort:
// any failure or cold start returns "".
func (s *Service) loadTutoringContext(ctx context.Context, profileID int64) string {
	conv, err := s.convRepo.FindLatestActiveByProfileAndPurpose(ctx, profileID, conversationDomain.PurposeQuizTutoring)
	if err != nil {
		logger.From(ctx).Warnf("quiz.tutoring_context_failed profile_id=%d err=%v", profileID, err)
		return ""
	}
	if conv == nil {
		return ""
	}

	history := appconv.NewChatMessageHistory(
		s.appendUserCmd, s.appendAssistantCmd, s.convMsgRepo,
		conv.ConversationId(), conv.UserId(), s.tutoringWindow*2)
	mem := langchain.NewWindowMemory(history, int(s.tutoringWindow))

	str, err := langchain.LoadHistoryString(ctx, mem)
	if err != nil {
		logger.From(ctx).Warnf("quiz.tutoring_context_failed profile_id=%d err=%v", profileID, err)
		return ""
	}
	return strings.TrimSpace(str)
}

// recordTutoringExchange appends the graded quiz's performance summary and
// review to the profile's tutoring thread (find-or-create), through the
// framework memory's SaveContext. Best-effort: grading is already persisted,
// so a memory failure is only logged.
func (s *Service) recordTutoringExchange(ctx context.Context, q *domain.Quiz, summary, review string) {
	if q.UserId() == nil || q.ProfileId() == nil {
		return
	}
	userID := *q.UserId()
	profileID := *q.ProfileId()

	conversationID, err := s.ensureTutoringConversation(ctx, userID, profileID)
	if err != nil {
		logger.From(ctx).Warnf("quiz.tutoring_record_failed quiz_id=%d err=%v", q.QuizId(), err)
		return
	}

	history := appconv.NewChatMessageHistory(
		s.appendUserCmd, s.appendAssistantCmd, s.convMsgRepo,
		conversationID, userID, s.tutoringWindow*2)
	mem := langchain.NewWindowMemory(history, int(s.tutoringWindow))

	if err := langchain.SaveTurn(ctx, mem, summary, review); err != nil {
		logger.From(ctx).Warnf("quiz.tutoring_record_failed quiz_id=%d err=%v", q.QuizId(), err)
	}
}

// ensureTutoringConversation resolves (find-or-create) the profile's
// QUIZ_TUTORING thread and returns its external id.
func (s *Service) ensureTutoringConversation(ctx context.Context, userID, profileID int64) (int64, error) {
	conv, err := s.convRepo.FindLatestActiveByProfileAndPurpose(ctx, profileID, conversationDomain.PurposeQuizTutoring)
	if err != nil {
		return 0, err
	}
	if conv != nil {
		return conv.ConversationId(), nil
	}
	created, err := s.createConvCmd.Handle(ctx, convcommand.CreateConversationCommand{
		UserID:    userID,
		ProfileID: &profileID,
		Purpose:   conversationDomain.PurposeQuizTutoring,
	})
	if err != nil {
		return 0, err
	}
	return created.ConversationId(), nil
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
