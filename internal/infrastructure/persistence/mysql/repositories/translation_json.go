package repositories

import (
	"encoding/json"
	"fmt"

	"math-ai.com/math-ai/internal/domain/questionpool"
	"math-ai.com/math-ai/internal/shared/enum"
)

// translationJSON is the stored shape of one entry of the translations
// column of ma_question_pools and ma_questions (migrations 036 / 037).
type translationJSON struct {
	QuestionText  string                  `json:"question_text"`
	TtsText       string                  `json:"tts_text,omitempty"`
	Explanation   string                  `json:"explanation,omitempty"`
	QuestionTopic string                  `json:"question_topic,omitempty"`
	Options       []translationOptionJSON `json:"options,omitempty"`
}

type translationOptionJSON struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

// marshalTranslations is the column value: NULL for none.
func marshalTranslations(in questionpool.Translations) (*string, error) {
	if len(in) == 0 {
		return nil, nil
	}
	out := make(map[enum.LanguageType]translationJSON, len(in))
	for lang, t := range in {
		j := translationJSON{QuestionText: t.QuestionText, TtsText: t.TtsText, Explanation: t.Explanation, QuestionTopic: t.QuestionTopic}
		for _, o := range t.Options {
			j.Options = append(j.Options, translationOptionJSON{Id: o.Id, Text: o.Text})
		}
		out[lang] = j
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("marshal translations: %w", err)
	}
	s := string(raw)
	return &s, nil
}

func unmarshalTranslations(raw *string) (questionpool.Translations, error) {
	if raw == nil || *raw == "" || *raw == "null" {
		return nil, nil
	}
	var stored map[enum.LanguageType]translationJSON
	if err := json.Unmarshal([]byte(*raw), &stored); err != nil {
		return nil, err
	}
	if len(stored) == 0 {
		return nil, nil
	}
	out := make(questionpool.Translations, len(stored))
	for lang, j := range stored {
		t := questionpool.Translation{QuestionText: j.QuestionText, TtsText: j.TtsText, Explanation: j.Explanation, QuestionTopic: j.QuestionTopic}
		for _, o := range j.Options {
			t.Options = append(t.Options, questionpool.Option{Id: o.Id, Text: o.Text})
		}
		out[lang] = t
	}
	return out, nil
}

// languageOrDefault is what a row stores for a question that states no
// language: the column's own default.
func languageOrDefault(l enum.LanguageType) string {
	if l == "" {
		return string(questionpool.DefaultLanguage)
	}
	return string(l)
}
