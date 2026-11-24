package dto

import (
	"slices"
	"strings"
	"unicode/utf8"

	"math-ai.com/math-ai/internal/domain/questionpool"
	"math-ai.com/math-ai/internal/shared/enum"
)

// NormalizeIDs trims and lower-cases the option ids and the key of a
// model-written question, so "A" and "a " name the option "a" both in its
// list and in its key. Bank ids are already canonical and are not passed
// through here.
func (q *Question) NormalizeIDs() {
	clean := func(id string) string { return strings.ToLower(strings.TrimSpace(id)) }
	for i := range q.Options {
		q.Options[i].ID = clean(q.Options[i].ID)
	}
	if q.CorrectAnswer != nil {
		for i := range q.CorrectAnswer.OptionIDs {
			q.CorrectAnswer.OptionIDs[i] = clean(q.CorrectAnswer.OptionIDs[i])
		}
	}
}

// NormalizeExtras applies the rules of ma_question_pools to the
// explanation, the tts_text and the tags, leniently: a model's slip must not
// fail the round, so the explanation and the tts_text are trimmed and cut at
// MaxExplanationRunes / MaxTtsTextRunes, and
// tags are trimmed, blanks and repeats dropped, over-long ones dropped and
// the list cut at MaxTags — where the bank would refuse the question.
func (q *Question) NormalizeExtras() {
	q.Explanation = strings.TrimSpace(q.Explanation)
	if r := []rune(q.Explanation); len(r) > questionpool.MaxExplanationRunes {
		q.Explanation = strings.TrimSpace(string(r[:questionpool.MaxExplanationRunes]))
	}
	q.TtsText = strings.TrimSpace(q.TtsText)
	if r := []rune(q.TtsText); len(r) > questionpool.MaxTtsTextRunes {
		q.TtsText = strings.TrimSpace(string(r[:questionpool.MaxTtsTextRunes]))
	}
	tags := make([]string, 0, len(q.QuestionTags))
	for _, t := range q.QuestionTags {
		t = strings.TrimSpace(t)
		if t == "" || utf8.RuneCountInString(t) > questionpool.MaxTagRunes || slices.Contains(tags, t) {
			continue
		}
		tags = append(tags, t)
	}
	if len(tags) > questionpool.MaxTags {
		tags = tags[:questionpool.MaxTags]
	}
	q.QuestionTags = tags
}

// NormalizeTranslations applies the bank's translation rules to a
// model-written question, leniently (a slip must not fail the round): the
// language is read as a short code (unknown = Vietnamese); an entry keyed by
// an unknown language or the question's own, or with no question_text, is
// dropped; text is trimmed and cut at the bank's limits; translated options
// that are not exactly the question's ids (after NormalizeIDs' lower-casing)
// are dropped, so the base option text shows instead. It returns the
// languages it dropped, for the caller to log.
func (q *Question) NormalizeTranslations() []string {
	q.Language = LanguageOf(q.Language)
	var dropped []string
	if len(q.Translations) == 0 {
		q.Translations = nil
		return nil
	}
	out := make(map[enum.LanguageType]QuestionTranslation, len(q.Translations))
	for lang, t := range q.Translations {
		key := lang.Base()
		if key == "" || key == q.Language || key != lang {
			dropped = append(dropped, string(lang))
			continue
		}
		t.QuestionText = cut(t.QuestionText, questionpool.MaxTtsTextRunes)
		if t.QuestionText == "" {
			dropped = append(dropped, string(lang))
			continue
		}
		t.TtsText = cut(t.TtsText, questionpool.MaxTtsTextRunes)
		t.Explanation = cut(t.Explanation, questionpool.MaxExplanationRunes)
		t.QuestionTopic = cut(t.QuestionTopic, 255)
		for i := range t.Options {
			t.Options[i] = QuestionOption{ID: strings.ToLower(strings.TrimSpace(t.Options[i].ID)), Text: cut(t.Options[i].Text, 500)}
		}
		if !sameOptionIDs(q.Options, t.Options) {
			t.Options = nil
		}
		out[key] = t
	}
	q.Translations = nil
	if len(out) > 0 {
		q.Translations = out
	}
	slices.Sort(dropped)
	return dropped
}

func cut(s string, max int) string {
	s = strings.TrimSpace(s)
	if r := []rune(s); len(r) > max {
		s = strings.TrimSpace(string(r[:max]))
	}
	return s
}

// sameOptionIDs reports whether tr names exactly the ids of base, each once.
func sameOptionIDs(base, tr []QuestionOption) bool {
	if len(base) != len(tr) {
		return false
	}
	seen := make(map[string]bool, len(tr))
	for _, o := range tr {
		if seen[o.ID] || !slices.ContainsFunc(base, func(b QuestionOption) bool { return b.ID == o.ID }) {
			return false
		}
		seen[o.ID] = true
	}
	return true
}
