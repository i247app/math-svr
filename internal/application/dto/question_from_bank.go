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

func trueFalseOptions(lang enum.LanguageType) []QuestionOption {
	t, ok := trueFalseTexts[lang]
	if !ok {
		t = trueFalseTexts[enum.LanguageTypeVietnamese]
	}
	return []QuestionOption{
		{ID: TrueOptionID, Text: t[0]},
		{ID: FalseOptionID, Text: t[1]},
	}
}

// bankTranslations copies a bank row's translations into the exam shape.
// A TRUE_FALSE question has no options in the bank; the exam gives it the
// two served ones in each language.
func bankTranslations(q *questionpool.QuestionPool) map[enum.LanguageType]QuestionTranslation {
	if len(q.Translations()) == 0 {
		return nil
	}
	out := make(map[enum.LanguageType]QuestionTranslation, len(q.Translations()))
	for lang, t := range q.Translations() {
		e := QuestionTranslation{
			QuestionText:  t.QuestionText,
			TtsText:       t.TtsText,
			Explanation:   t.Explanation,
			QuestionTopic: t.QuestionTopic,
		}
		for _, o := range t.Options {
			e.Options = append(e.Options, QuestionOption{ID: o.Id, Text: o.Text})
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
// kind (CorrectAnswer documents the four shapes).
// Image keys travel with the question and are presigned when a response is
// built, by the module that serves it.
func QuestionFromBank(q *questionpool.QuestionPool, number int) (Question, error) {
	out := Question{
		QuestionNumber:   number,
		QuestionText:     q.QuestionText(),
		QuestionTopic:    utils.DerefString(q.QuestionTopic()),
		QuestionTags:     q.QuestionTags(),
		Explanation:      utils.DerefString(q.Explanation()),
		TtsText:          utils.DerefString(q.TtsText()),
		QuestionKind:     q.QuestionKind().String(),
		QuestionLevel:    q.QuestionLevel(),
		QuestionSubLevel: q.QuestionSubLevel(),
		Language:         LanguageOf(q.QuestionLanguage()),
		Translations:     bankTranslations(q),
	}
	grade := q.QuestionGrade()
	out.QuestionGrade = &grade
	out.QuestionImageKey = utils.DerefString(q.QuestionImageKey())

	out.Options = make([]QuestionOption, 0, len(q.Options()))
	has := make(map[string]bool, len(q.Options()))
	for _, o := range q.Options() {
		has[o.Id] = true
		out.Options = append(out.Options, QuestionOption{ID: o.Id, Text: o.Text, ImageKey: utils.DerefString(o.ImageKey)})
	}
	correct := q.CorrectAnswer()

	switch q.QuestionKind() {
	case questionpool.KindSingleChoice:
		if len(correct.OptionIds) != 1 {
			return out, fmt.Errorf("question %d: single-choice key does not name exactly one option", q.QuestionPoolId())
		}
		if !has[correct.OptionIds[0]] {
			return out, fmt.Errorf("question %d: single-choice key names unknown option %q", q.QuestionPoolId(), correct.OptionIds[0])
		}
		out.CorrectAnswer = &CorrectAnswer{OptionIDs: slices.Clone(correct.OptionIds)}
	case questionpool.KindMultipleChoice:
		for _, id := range correct.OptionIds {
			if !has[id] {
				return out, fmt.Errorf("question %d: multiple-choice key names unknown option %q", q.QuestionPoolId(), id)
			}
		}
		if len(correct.OptionIds) == 0 {
			return out, fmt.Errorf("question %d: multiple-choice key is empty", q.QuestionPoolId())
		}
		ids := slices.Clone(correct.OptionIds)
		slices.Sort(ids)
		out.CorrectAnswer = &CorrectAnswer{OptionIDs: ids}
	case questionpool.KindTrueFalse:
		if correct.Value == nil {
			return out, fmt.Errorf("question %d: true/false key has no value", q.QuestionPoolId())
		}
		value := *correct.Value
		out.Options = trueFalseOptions(out.Language)
		out.CorrectAnswer = &CorrectAnswer{Value: &value}
	case questionpool.KindNumericInput:
		if len(correct.Values) == 0 {
			return out, fmt.Errorf("question %d: numeric key has no values", q.QuestionPoolId())
		}
		out.CorrectAnswer = &CorrectAnswer{Values: slices.Clone(correct.Values)}
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
