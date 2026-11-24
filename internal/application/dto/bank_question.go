package dto

import (
	"fmt"
	"slices"

	"math-ai.com/math-ai/internal/domain/questionpool"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// Texts of the two options a TRUE_FALSE bank question is served with, per
// language; a language missing here falls back to Vietnamese.
var trueFalseTexts = map[enum.LanguageType][2]string{
	enum.LanguageTypeVietnamese: {"Đúng", "Sai"},
	enum.LanguageTypeEnglish:    {"True", "False"},
}

func trueFalseOptions(lang enum.LanguageType) []ExamOption {
	t, ok := trueFalseTexts[lang]
	if !ok {
		t = trueFalseTexts[enum.LanguageTypeVietnamese]
	}
	return []ExamOption{
		{ID: TrueOptionID, Text: t[0]},
		{ID: FalseOptionID, Text: t[1]},
	}
}

// bankTranslations copies a bank row's translations into the exam shape.
// A TRUE_FALSE question has no options in the bank; the exam gives it the
// two served ones in each language.
func bankTranslations(q *questionpool.QuestionPool) map[enum.LanguageType]ExamTranslation {
	if len(q.Translations()) == 0 {
		return nil
	}
	out := make(map[enum.LanguageType]ExamTranslation, len(q.Translations()))
	for lang, t := range q.Translations() {
		e := ExamTranslation{
			QuestionText:  t.QuestionText,
			TtsText:       t.TtsText,
			Explanation:   t.Explanation,
			QuestionTopic: t.QuestionTopic,
		}
		for _, o := range t.Options {
			e.Options = append(e.Options, ExamOption{ID: o.Id, Text: o.Text})
		}
		if q.QuestionKind() == questionpool.KindTrueFalse {
			e.Options = trueFalseOptions(lang)
		}
		out[lang] = e
	}
	return out
}

// QuestionFromBank converts a bank question (ma_question_pools) into the
// question at position number of a paper (exam) or round (game).
// Options keep their bank ids, in display order; the key follows from the
// kind (Question documents the four shapes).
// Image keys travel with the question and are presigned when a response is
// built, by the module that serves it.
func QuestionFromBank(q *questionpool.QuestionPool, number int) (Question, error) {
	out := Question{
		QuestionNumber:   number,
		QuestionName:     q.QuestionText(),
		QuestionTopic:    utils.DerefString(q.QuestionTopic()),
		QuestionTags:     q.QuestionTags(),
		Explanation:      utils.DerefString(q.Explanation()),
		TtsText:          utils.DerefString(q.TtsText()),
		Kind:             q.QuestionKind().String(),
		QuestionLevel:    q.QuestionLevel(),
		QuestionSubLevel: q.QuestionSubLevel(),
		Language:         LanguageOf(q.QuestionLanguage()),
		Translations:     bankTranslations(q),
	}
	grade := q.QuestionGrade()
	out.QuestionGrade = &grade
	out.ImageKey = utils.DerefString(q.QuestionImageKey())

	textOf := make(map[string]string, len(q.Options()))
	for _, o := range q.Options() {
		textOf[o.Id] = o.Text
		out.Answers = append(out.Answers, AnswerChoice{Label: o.Id, Content: o.Text, ImageKey: utils.DerefString(o.ImageKey)})
	}
	correct := q.CorrectAnswer()

	switch q.QuestionKind() {
	case questionpool.KindSingleChoice:
		if len(correct.OptionIds) != 1 {
			return out, fmt.Errorf("question %d: single-choice key does not name exactly one option", q.QuestionPoolId())
		}
		text, ok := textOf[correct.OptionIds[0]]
		if !ok {
			return out, fmt.Errorf("question %d: single-choice key names unknown option %q", q.QuestionPoolId(), correct.OptionIds[0])
		}
		out.RightAnswerLabel = correct.OptionIds[0]
		out.RightAnswerContent = text
	case questionpool.KindMultipleChoice:
		for _, id := range correct.OptionIds {
			if _, ok := textOf[id]; !ok {
				return out, fmt.Errorf("question %d: multiple-choice key names unknown option %q", q.QuestionPoolId(), id)
			}
			out.RightOptionIDs = append(out.RightOptionIDs, id)
		}
		slices.Sort(out.RightOptionIDs)
		if len(out.RightOptionIDs) == 0 {
			return out, fmt.Errorf("question %d: multiple-choice key is empty", q.QuestionPoolId())
		}
	case questionpool.KindTrueFalse:
		if correct.Value == nil {
			return out, fmt.Errorf("question %d: true/false key has no value", q.QuestionPoolId())
		}
		out.Answers = nil
		for _, o := range trueFalseOptions(out.Language) {
			out.Answers = append(out.Answers, AnswerChoice{Label: o.ID, Content: o.Text})
		}
		out.RightAnswerLabel = FalseOptionID
		if *correct.Value {
			out.RightAnswerLabel = TrueOptionID
		}
	case questionpool.KindNumericInput:
		if len(correct.Values) == 0 {
			return out, fmt.Errorf("question %d: numeric key has no values", q.QuestionPoolId())
		}
		out.RightValues = slices.Clone(correct.Values)
	default:
		return out, fmt.Errorf("question %d: unknown kind %q", q.QuestionPoolId(), q.QuestionKind())
	}
	return out, nil
}

// BankDifficulty orders bank questions: level, then sub-level. A
// question that is not rated counts as the middle of the scale.
func BankDifficulty(q *questionpool.QuestionPool) int {
	const unrated = 4
	level, sub := unrated, unrated
	if l := q.QuestionLevel(); l != nil {
		level = *l
	}
	if l := q.QuestionSubLevel(); l != nil {
		sub = *l
	}
	return level*10 + sub
}
