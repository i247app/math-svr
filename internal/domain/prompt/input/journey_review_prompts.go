package input

import (
	"errors"
	"fmt"
	"strings"
)

// Journey review: the model reads what a child answered across one
// journey and writes feedback in two lengths. The instructions are in
// English for the same reason the exam prompt is (shorter tokens); the
// review itself is always Vietnamese — it is read by a Vietnamese parent
// and child.
//
// Only answers reach the prompt: question, topic, the right answer, the
// child's answer and whether it was right. Never the child's name or any
// other profile field — the review needs none of it, and a prompt is
// sent to a third party.

// ReviewAnswer is one answered question as the reviewer sees it.
type ReviewAnswer struct {
	Question    string
	Topic       string
	Grade       *int // the grade the question targets (a probe is one above)
	RightAnswer string
	ChildAnswer string
	Correct     bool
}

// JourneyReviewInput is everything the review prompt is built from.
type JourneyReviewInput struct {
	ExamType string // ASSESSMENT or GRADE
	Grade    *int   // the journey's current grade
	Level    *int   // GRADE only
	// ScorePercentage is the journey's cumulative score, when it has one.
	ScorePercentage *int
	// Answers in the order they were given, oldest first.
	Answers []ReviewAnswer
}

// ReviewShortMaxChars bounds the short review: the size of the column it
// is stored in (ai_review_short), enforced on the reply whatever the stored
// system prompt asks for.
const ReviewShortMaxChars = 200

var ErrReviewNoAnswers = errors.New("bot: a journey review needs at least one answered question")

// BuildJourneyReviewUserPrompt renders the USER message: the journey and
// its answers. The SYSTEM prompt is the EXAM_REVIEW row of ma_prompts,
// written by admins and sent verbatim.
func BuildJourneyReviewUserPrompt(in JourneyReviewInput) (string, error) {
	if len(in.Answers) == 0 {
		return "", ErrReviewNoAnswers
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Exam type: %s\n", in.ExamType)
	if in.Grade != nil {
		fmt.Fprintf(&b, "Current grade: %d (%s)\n", *in.Grade, GradeLabel(*in.Grade))
	}
	if in.Level != nil {
		fmt.Fprintf(&b, "Level: %d (scale 0–9)\n", *in.Level)
	}
	correct := 0
	for _, a := range in.Answers {
		if a.Correct {
			correct++
		}
	}
	fmt.Fprintf(&b, "Answered: %d, correct: %d", len(in.Answers), correct)
	if in.ScorePercentage != nil {
		fmt.Fprintf(&b, " (journey score %d%%)", *in.ScorePercentage)
	}
	b.WriteString("\n\nANSWERS (oldest first):\n")
	for i, a := range in.Answers {
		mark := "WRONG"
		if a.Correct {
			mark = "RIGHT"
		}
		fmt.Fprintf(&b, "%d. [%s", i+1, strings.TrimSpace(a.Topic))
		if a.Grade != nil {
			fmt.Fprintf(&b, ", grade %d", *a.Grade)
		}
		fmt.Fprintf(&b, "] %s — correct: %s; child: %s — %s\n",
			strings.TrimSpace(a.Question), a.RightAnswer, a.ChildAnswer, mark)
	}
	return strings.TrimRight(b.String(), "\n"), nil
}
