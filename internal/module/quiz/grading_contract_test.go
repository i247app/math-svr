package quiz

import (
	"strings"
	"testing"

	domainBot "math-ai.com/math-ai/internal/domain/bot"
)

// The grading JSON key is a contract shared by two sides that never
// reference each other: the prompt templates tell the model which key to
// emit, and QuizGradingResult's struct tag tells the parser which key to
// read. If they drift, parsing still "succeeds" and the review silently
// lands as an empty string — no error anywhere. These tests pin both ends
// so a future rename cannot half-land.

func TestParseGradedQuizReadsReviewKey(t *testing.T) {
	const payload = `{"total_questions":10,"correct_number":8,"score_percentage":80,` +
		`"review":"Phép cộng tốt; cần luyện thêm phép trừ có nhớ.","assessment_grade":"Grade 1"}`

	got, err := ParseGradedQuiz(payload)
	if err != nil {
		t.Fatalf("ParseGradedQuiz: %v", err)
	}
	if got.Review == "" {
		t.Fatal("Review is empty — the struct tag no longer matches the key the prompt asks for")
	}
	if got.Review != "Phép cộng tốt; cần luyện thêm phép trừ có nhớ." {
		t.Errorf("Review = %q", got.Review)
	}
	if got.ScorePercentage != 80 || got.CorrectNumber != 8 || got.TotalQuestions != 10 {
		t.Errorf("score fields mis-parsed: %+v", got)
	}
}

// TestGradingPromptsAskForReviewKey walks every grading prompt in both
// languages and both purposes, asserting each one names "review" and none
// still names the old "ai_review".
func TestGradingPromptsAskForReviewKey(t *testing.T) {
	kinds := []domainBot.QuizPromptKind{
		domainBot.QuizPromptKindGrade,
		domainBot.QuizPromptKindGradeReinforce,
	}
	langs := []domainBot.QuizLanguage{"vi", "en"}
	purposes := []domainBot.QuizPurpose{
		domainBot.QuizPurposeAssessment,
		domainBot.QuizPurposePractice,
	}

	for _, kind := range kinds {
		for _, lang := range langs {
			for _, purpose := range purposes {
				sys, _, err := domainBot.BuildQuizPrompt(kind, domainBot.QuizPromptInput{
					Language: lang, Purpose: purpose,
					Questions: "[]", Answers: "{}",
				})
				if err != nil {
					t.Fatalf("kind=%d lang=%s purpose=%s: %v", kind, lang, purpose, err)
				}
				if !strings.Contains(sys, `"review"`) {
					t.Errorf("kind=%d lang=%s purpose=%s: prompt never names the \"review\" key", kind, lang, purpose)
				}
				if strings.Contains(sys, `"ai_review"`) {
					t.Errorf("kind=%d lang=%s purpose=%s: prompt still names the old \"ai_review\" key", kind, lang, purpose)
				}
			}
		}
	}
}
