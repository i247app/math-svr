package question

import (
	"regexp"
	"strings"
)

// validTypes is the closed render set the client understands. Anything
// else, including empty, normalizes to TypeArithmetic.
var validTypes = map[string]struct{}{
	TypeArithmetic:    {},
	TypeCount:         {},
	TypePickByIcon:    {},
	TypeIdentifyShape: {},
}

// typeAliases folds spellings the exam prompt's own schema example uses
// onto the render set. COUNTING is the one the few-shot example teaches;
// without this it would fall to ARITHMETIC and lose the count rendering
// for exactly the questions that need it most.
var typeAliases = map[string]string{
	"COUNTING": TypeCount,
}

// geometryIconWhitelist is the closed set of shape tokens the client ships
// SVG assets for. The prompt tells the model to stay inside this set;
// Normalize only REPORTS drift — it never drops a question, because a
// stray token still renders as a client-side fallback while the MCQ
// structure (labels + right_answer) stays valid and gradable.
var geometryIconWhitelist = map[string]struct{}{
	"triangle": {}, "square": {}, "rectangle": {}, "circle": {},
	"star": {}, "diamond": {}, "oval": {}, "pentagon": {},
	"hexagon": {}, "heart": {},
}

// iconTokenRe matches the "[icon:NAME]" shape tokens embedded in a stem or
// an answer's content. Emoji are literal UTF-8 and intentionally unmatched.
var iconTokenRe = regexp.MustCompile(`\[icon:([a-zA-Z_]+)\]`)

// Warning kinds reported by Normalize.
const (
	WarnUnknownType      = "unknown_question_type"
	WarnUnknownIconToken = "unknown_icon_token"
)

// Warning is one piece of drift found in a generated payload. Normalize
// returns these rather than logging them itself: this package sits in the
// application layer, which may not import the infrastructure logger. The
// caller — always a module-layer service — decides how loudly to report.
type Warning struct {
	QuestionNumber int
	Kind           string
	Value          string
}

// Normalize clamps each question's type to the known set (unknown or
// empty becomes ARITHMETIC) and reports any [icon:NAME] token outside the
// geometry whitelist. It mutates in place and returns the same slice.
//
// Visual drift never fails a generation: grading is label-based and
// unaffected by icon content.
func Normalize(questions []Question) ([]Question, []Warning) {
	var warnings []Warning

	collect := func(s string, qnum int) {
		for _, m := range iconTokenRe.FindAllStringSubmatch(s, -1) {
			if _, ok := geometryIconWhitelist[strings.ToLower(m[1])]; !ok {
				warnings = append(warnings, Warning{
					QuestionNumber: qnum,
					Kind:           WarnUnknownIconToken,
					Value:          m[1],
				})
			}
		}
	}

	for i := range questions {
		q := &questions[i]
		qt := strings.ToUpper(strings.TrimSpace(q.QuestionType))
		if canonical, ok := typeAliases[qt]; ok {
			qt = canonical
		}
		if _, ok := validTypes[qt]; !ok {
			if qt != "" {
				warnings = append(warnings, Warning{
					QuestionNumber: q.QuestionNumber,
					Kind:           WarnUnknownType,
					Value:          q.QuestionType,
				})
			}
			qt = TypeArithmetic
		}
		q.QuestionType = qt

		collect(q.QuestionName, q.QuestionNumber)
		for _, a := range q.Answers {
			collect(a.Content, q.QuestionNumber)
		}
	}
	return questions, warnings
}
