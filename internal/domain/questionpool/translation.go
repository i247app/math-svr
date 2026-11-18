package questionpool

import (
	"fmt"
	"slices"

	"math-ai.com/math-ai/internal/shared/enum"
)

// DefaultLanguage is the base language of a new bank question that states
// none. Rows written before the column existed are Vietnamese: the hand-run
// sql/prod/2026-11-17 script labels them 'vi'.
const DefaultLanguage = enum.LanguageTypeEnglish

// Translation is a question's display text in another language. Only text is
// translated: ids, images, the order of the options and the key stay on the
// base question. Options carry Id + Text only. "" = not translated.
type Translation struct {
	QuestionText  string
	TtsText       string
	Explanation   string
	QuestionTopic string
	Options       []Option
}

// Translations maps a language (a short code of enum.ContentLanguages, never
// the question's own language) to its translation. nil = none.
type Translations map[enum.LanguageType]Translation

// CheckTranslationKey returns why lang cannot key a translation of a question
// written in base, or "".
func CheckTranslationKey(base, lang enum.LanguageType) string {
	switch {
	case !slices.Contains(enum.ContentLanguages, lang):
		return fmt.Sprintf("%q is not one of %v", lang, enum.ContentLanguages)
	case lang == base:
		return fmt.Sprintf("%q is the question's own language", lang)
	}
	return ""
}

// CheckTranslationOptions returns why tr cannot translate base, or "": it
// must name exactly the base option ids, each once, with text and no image.
// A question without options takes none. Array order does not matter.
func CheckTranslationOptions(base, tr []Option) string {
	if len(base) == 0 {
		if len(tr) > 0 {
			return "the question has no options to translate"
		}
		return ""
	}
	if len(tr) != len(base) {
		return fmt.Sprintf("needs exactly the %d option ids of the question, got %d", len(base), len(tr))
	}
	seen := make([]string, 0, len(tr))
	for i, o := range tr {
		switch {
		case !slices.ContainsFunc(base, func(b Option) bool { return b.Id == o.Id }):
			return fmt.Sprintf("[%d].id %q is not an option of the question", i, o.Id)
		case slices.Contains(seen, o.Id):
			return fmt.Sprintf("[%d].id %q is repeated", i, o.Id)
		case o.ImageKey != nil:
			return fmt.Sprintf("[%d].image_key: images are not translated", i)
		}
		seen = append(seen, o.Id)
	}
	return ""
}
