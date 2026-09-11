package scorer

import (
	"testing"

	"math-ai.com/math-ai/internal/application/dto/question"
	"math-ai.com/math-ai/internal/shared/enum"
)

// tenQuestions builds a payload where the right answer is always "A", so
// a test can pick a wrong answer simply by choosing "B".
const tenQuestions = `[
{"question_number":1,"question_name":"1+1","answers":[{"label":"A","content":"2"},{"label":"B","content":"3"}],"right_answer_label":"A","question_topic":"add"},
{"question_number":2,"question_name":"2+2","answers":[{"label":"A","content":"4"},{"label":"B","content":"5"}],"right_answer_label":"A","question_topic":"add"},
{"question_number":3,"question_name":"3+3","answers":[{"label":"A","content":"6"},{"label":"B","content":"7"}],"right_answer_label":"A","question_topic":"add"},
{"question_number":4,"question_name":"4+4","answers":[{"label":"A","content":"8"},{"label":"B","content":"9"}],"right_answer_label":"A","question_topic":"add"},
{"question_number":5,"question_name":"5+5","answers":[{"label":"A","content":"10"},{"label":"B","content":"11"}],"right_answer_label":"A","question_topic":"add"},
{"question_number":6,"question_name":"6+6","answers":[{"label":"A","content":"12"},{"label":"B","content":"13"}],"right_answer_label":"A","question_topic":"sub"},
{"question_number":7,"question_name":"7+7","answers":[{"label":"A","content":"14"},{"label":"B","content":"15"}],"right_answer_label":"A","question_topic":"sub"},
{"question_number":8,"question_name":"8+8","answers":[{"label":"A","content":"16"},{"label":"B","content":"17"}],"right_answer_label":"A","question_topic":"sub"},
{"question_number":9,"question_name":"9+9","answers":[{"label":"A","content":"18"},{"label":"B","content":"19"}],"right_answer_label":"A","question_topic":"sub"},
{"question_number":10,"question_name":"10+10","answers":[{"label":"A","content":"20"},{"label":"B","content":"21"}],"right_answer_label":"A","question_topic":"sub"}]`

func answerAll(n int, label string) []question.StudentAnswer {
	out := make([]question.StudentAnswer, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, question.StudentAnswer{QuestionNumber: i, Label: label})
	}
	return out
}

// TestDenominatorSplit is the whole reason the parameter exists: six
// correct answers out of a ten-question sheet is 60% to a quiz and 100%
// to an exam, and both are intended.
func TestDenominatorSplit(t *testing.T) {
	answers := answerAll(6, "A")

	all, err := ScoreDetailed(tenQuestions, answers, enum.LanguageTypeVietnamese, DenominatorAllQuestions)
	if err != nil {
		t.Fatalf("all-questions: %v", err)
	}
	if all.TotalQuestions != 10 || all.CorrectNumber != 6 || all.ScorePercentage != 60 {
		t.Errorf("all-questions: total=%d correct=%d pct=%d, want 10/6/60",
			all.TotalQuestions, all.CorrectNumber, all.ScorePercentage)
	}

	answered, err := ScoreDetailed(tenQuestions, answers, enum.LanguageTypeVietnamese, DenominatorAnsweredOnly)
	if err != nil {
		t.Fatalf("answered-only: %v", err)
	}
	if answered.TotalQuestions != 6 || answered.CorrectNumber != 6 || answered.ScorePercentage != 100 {
		t.Errorf("answered-only: total=%d correct=%d pct=%d, want 6/6/100",
			answered.TotalQuestions, answered.CorrectNumber, answered.ScorePercentage)
	}

	// Whichever denominator was used, the blanks stay visible — this is
	// what lets a placement rule tell the two 100%s apart.
	if all.SkippedNumber != 4 || answered.SkippedNumber != 4 {
		t.Errorf("skipped = %d / %d, want 4 in both modes", all.SkippedNumber, answered.SkippedNumber)
	}
}

// TestScoreKeepsLegacyBehaviour guards quiz and exercise submissions: the
// exported Score must not have moved when the exam flow added its rule.
func TestScoreKeepsLegacyBehaviour(t *testing.T) {
	got, err := Score(tenQuestions, answerAll(6, "A"), enum.LanguageTypeVietnamese)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}
	if got.TotalQuestions != 10 || got.ScorePercentage != 60 {
		t.Errorf("total=%d pct=%d, want 10/60", got.TotalQuestions, got.ScorePercentage)
	}
}

func TestScoreDetailedOutcomes(t *testing.T) {
	answers := []question.StudentAnswer{
		{QuestionNumber: 1, Label: "A"},
		{QuestionNumber: 2, Label: "B"},
	}
	got, err := ScoreDetailed(tenQuestions, answers, enum.LanguageTypeVietnamese, DenominatorAnsweredOnly)
	if err != nil {
		t.Fatalf("ScoreDetailed: %v", err)
	}

	if len(got.Outcomes) != 2 {
		t.Fatalf("got %d outcomes, want 2 — skipped questions must produce none", len(got.Outcomes))
	}
	if !got.Outcomes[0].IsCorrect || got.Outcomes[1].IsCorrect {
		t.Errorf("outcomes correctness = %v/%v, want true/false",
			got.Outcomes[0].IsCorrect, got.Outcomes[1].IsCorrect)
	}
	// The chosen option's text is snapshot so a review row can render what
	// the child actually picked without re-reading the questions blob.
	if got.Outcomes[1].SelectedContent != "5" {
		t.Errorf("SelectedContent = %q, want %q", got.Outcomes[1].SelectedContent, "5")
	}
	if got.Outcomes[0].Question.QuestionName != "1+1" {
		t.Errorf("outcome lost its question snapshot: %+v", got.Outcomes[0].Question)
	}
}
