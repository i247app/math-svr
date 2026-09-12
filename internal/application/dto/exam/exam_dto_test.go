package exam

import (
	"encoding/json"
	"math/rand/v2"
	"testing"

	"math-ai.com/math-ai/internal/application/dto/question"
	domain "math-ai.com/math-ai/internal/domain/exam"
)

// storedSet is a two-question set in the exam vocabulary, as it sits in
// ma_ai_exams.ai_questions_json. The right answer is canonically "A" on
// both, with distinct contents so a mis-ordered option is detectable.
const storedSet = `[
{"question_number":1,"question_name":"1+1","answers":[{"label":"A","content":"2"},{"label":"B","content":"3"},{"label":"C","content":"4"},{"label":"D","content":"5"}],"right_answer_label":"A","right_answer_content":"2","question_grade":1},
{"question_number":2,"question_name":"2+2","answers":[{"label":"A","content":"4"},{"label":"B","content":"5"},{"label":"C","content":"6"},{"label":"D","content":"7"}],"right_answer_label":"A","right_answer_content":"4","question_grade":1}]`

func aiExam(id int64, json string) *domain.AiExam {
	e := domain.NewAiExam()
	e.SetAiExamId(id)
	e.SetAiQuestionsJson(json)
	return e
}

func detail(attemptID, aiExamID int64, qn int, selected string, correct bool) *domain.UserExamDetail {
	d := domain.NewUserExamDetail()
	d.SetUserAiExamId(attemptID)
	d.SetAiExamId(aiExamID)
	d.SetQuestionNumber(qn)
	right := "A"
	d.SetRightAnswerLabel(&right)
	d.SetSelectedLabel(selected)
	d.SetIsCorrect(correct)
	return d
}

func strp(s string) *string { return &s }

// TestDetailsCarryEveryOption is the fix for the review screen that could
// only show two lines: every row must carry the full option list.
func TestDetailsCarryEveryOption(t *testing.T) {
	sets := map[int64]*domain.AiExam{7: aiExam(7, storedSet)}
	rows := []*domain.UserExamDetail{
		detail(100, 7, 1, "A", true),
		detail(100, 7, 2, "C", false),
	}

	got := DetailsToResponse(rows, nil, sets)

	if len(got) != 2 {
		t.Fatalf("got %d rows, want 2", len(got))
	}
	for i, r := range got {
		if len(r.Answers) != 4 {
			t.Errorf("row %d carries %d options, want 4", i, len(r.Answers))
		}
	}
	// Unshuffled: stored order and labels, key and pick untouched.
	if got[0].Answers[0].Label != "A" || got[0].Answers[0].Content != "2" {
		t.Errorf("row 0 option 0 = %+v, want A/2", got[0].Answers[0])
	}
	if *got[1].RightAnswerLabel != "A" || got[1].SelectedLabel != "C" || got[1].IsCorrect {
		t.Errorf("row 1 key/pick = %s/%s correct=%v, want A/C false", *got[1].RightAnswerLabel, got[1].SelectedLabel, got[1].IsCorrect)
	}
}

// TestDetailsFollowTheSittingsShuffle: the option list, the key and the
// pick must all be expressed in the SAME served vocabulary, or the screen
// marks the wrong line. This drives a real shuffle through the mapping
// and checks the three agree with each other and with the content.
func TestDetailsFollowTheSittingsShuffle(t *testing.T) {
	var stored []ExamQuestion
	mustUnmarshal(t, storedSet, &stored)
	canonical := ToSharedQuestions(stored)

	for seed := uint64(1); seed <= 25; seed++ {
		sh := question.NewShuffle(canonical, rand.New(rand.NewPCG(seed, seed)))
		shuffles := map[int64]*question.Shuffle{100: sh}
		sets := map[int64]*domain.AiExam{7: aiExam(7, storedSet)}

		// Canonical log: question 2, the child picked canonical "B" (content "5").
		row := detail(100, 7, 2, "B", false)
		got := DetailsToResponse([]*domain.UserExamDetail{row}, shuffles, sets)[0]

		if got.QuestionNumber != sh.ServedNumber(2) {
			t.Fatalf("seed %d: served number %d, want %d", seed, got.QuestionNumber, sh.ServedNumber(2))
		}
		if len(got.Answers) != 4 {
			t.Fatalf("seed %d: %d options", seed, len(got.Answers))
		}
		// Labels are A..D in served order, and the content under the key
		// / the pick is what the canonical labels pointed at.
		for i, a := range got.Answers {
			if a.Label != string(rune('A'+i)) {
				t.Fatalf("seed %d: option %d labelled %q", seed, i, a.Label)
			}
		}
		if c := contentOf(got.Answers, *got.RightAnswerLabel); c != "4" {
			t.Fatalf("seed %d: key %q resolves to content %q, want 4", seed, *got.RightAnswerLabel, c)
		}
		if c := contentOf(got.Answers, got.SelectedLabel); c != "5" {
			t.Fatalf("seed %d: pick %q resolves to content %q, want 5", seed, got.SelectedLabel, c)
		}
	}
}

// TestDetailsSurviveAMissingSet: a retired question set must not break
// the review — the row renders without its option list, nothing more.
func TestDetailsSurviveAMissingSet(t *testing.T) {
	rows := []*domain.UserExamDetail{detail(100, 404, 1, "A", true)}

	got := DetailsToResponse(rows, nil, map[int64]*domain.AiExam{})
	if len(got) != 1 {
		t.Fatalf("got %d rows, want 1", len(got))
	}
	if got[0].Answers != nil {
		t.Errorf("expected no options for a missing set, got %v", got[0].Answers)
	}

	broken := map[int64]*domain.AiExam{404: aiExam(404, "not json")}
	got = DetailsToResponse(rows, nil, broken)
	if got[0].Answers != nil {
		t.Errorf("expected no options for an unreadable set, got %v", got[0].Answers)
	}
}

func contentOf(answers []question.AnswerChoice, label string) string {
	for _, a := range answers {
		if a.Label == label {
			return a.Content
		}
	}
	return ""
}

func mustUnmarshal(t *testing.T, raw string, into *[]ExamQuestion) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), into); err != nil {
		t.Fatalf("fixture: %v", err)
	}
}
