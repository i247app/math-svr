package bot

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

// Bounds the prompt asks for. ReviewShortMaxChars is also the column
// size the caller guards against (ai_review_short).
const (
	ReviewShortMaxChars = 200
	reviewLongMaxChars  = 1200
)

var ErrReviewNoAnswers = errors.New("bot: a journey review needs at least one answered question")

const systemJourneyReviewEN = `You review a Vietnamese child's math practice (kindergarten to grade 5) for the child's parent.

Write TWO reviews, both in natural, warm VIETNAMESE, addressed to the parent ("bé" for the child):
- "short": ONE or TWO sentences, at most %d characters. The overall result and the single most useful next step.
- "long": at most %d characters, plain text, 2–4 short paragraphs:
  1. What the child does well (name the topics).
  2. Where the child struggles: name the topics and the kind of mistake, using the wrong answers given (e.g. forgetting to carry, mixing up more/less).
  3. Two or three concrete practice suggestions a parent can do at home.

RULES
- Base every statement ONLY on the answers listed. Do not invent topics or mistakes.
- Be encouraging and specific; never shame the child. No emoji, no Markdown, no lists with symbols.
- If every answer is correct, praise and suggest moving to harder material instead of inventing weaknesses.
- Do not repeat the questions verbatim.

OUTPUT: only valid JSON, nothing else:
{"short": "...", "long": "..."}`

// BuildJourneyReviewPrompt renders the system and user messages.
func BuildJourneyReviewPrompt(in JourneyReviewInput) (system, user string, err error) {
	if len(in.Answers) == 0 {
		return "", "", ErrReviewNoAnswers
	}
	system = fmt.Sprintf(systemJourneyReviewEN, ReviewShortMaxChars, reviewLongMaxChars)

	var b strings.Builder
	fmt.Fprintf(&b, "Exam type: %s\n", in.ExamType)
	if in.Grade != nil {
		fmt.Fprintf(&b, "Current grade: %d (%s)\n", *in.Grade, ExamTitle(*in.Grade))
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
	return system, strings.TrimRight(b.String(), "\n"), nil
}
