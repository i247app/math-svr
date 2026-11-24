package dto

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// QuestionOption is one option of a question on the wire: its id (the key
// a child's answer names it by, stable wherever the option is shown — the
// client draws the letters A, B, C by position) and what is shown for it.
// ImageKey is the S3 key of an image shown with it (bank sets only);
// ImageURL is its presigned URL, set when a response is built and never
// stored. It is the shape of an option in the question bank's import file.
type QuestionOption struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	ImageKey string `json:"image_key,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

// Question is one question of an exam paper or a game round: what the model
// is asked to write (the OUTPUT FORMAT of the exam prompts), what
// ma_exam_pools.questions_json and ma_game_sessions.questions_json store,
// what the scorer grades and what the client is served.
//
// Its content comes from the model (then also copied to ma_questions) or
// from the admin bank (ma_question_pools, QuestionFromBank); both write the
// same fields. What only one of those tables has — ids, source, review
// status, verified_count, audit — is deliberately not here: this struct is
// served to children. The bank's admin shapes are in question_pool_dto.go.
//
// The JSON tags are a wire contract with the mobile client, the LLM and the
// stored rows: renaming a Go field is free, renaming a tag is not. There is
// no reader for older names (question_name, answers[{label, content}],
// right_answer_label, right_option_ids, ...): CheckSet refuses a set written
// under them rather than let it parse as empty questions. Classroom
// exercises keep those names on ExerciseQuestion.
//
//   - QuestionKind is SINGLE_CHOICE (the only kind a model writes),
//     MULTIPLE_CHOICE, TRUE_FALSE or NUMERIC_INPUT; absent = single-choice.
//   - CorrectAnswer is the key, in the shape of ma_question_pools.correct_answer
//     (CorrectAnswer). It is blanked on a live exam.
//   - QuestionImageKey / QuestionImageURL: the stem's image, as for options.
//   - Explanation is the worked solution (why the key is right): it gives
//     the answer away, so it is blanked with the key on a live exam.
//     QuestionTags are short English labels ("addition"). Both follow the
//     rules of ma_question_pools (NormalizeExtras).
//   - TtsText is the stem as the client reads it aloud (text-to-speech):
//     numbers and symbols written out as words. It gives nothing away, so it
//     is served on a live exam; trimmed and cut at MaxTtsTextRunes
//     (NormalizeExtras). Empty = the client has nothing to read.
//   - QuestionLevel / QuestionSubLevel: the level pair, 0..9 each, on the
//     scale of ma_question_pools. The prompt asks for them; the server
//     stamps them (exam.StampQuestionLevels) like question_grade.
type Question struct {
	Language         enum.LanguageType `json:"lang"`
	QuestionNumber   int               `json:"question_number"`
	QuestionKind     string            `json:"question_kind,omitempty"`
	QuestionText     string            `json:"question_text"`
	QuestionImageKey string            `json:"question_image_key,omitempty"`
	QuestionImageURL string            `json:"question_image_url,omitempty"`
	Options          []QuestionOption  `json:"options"`
	CorrectAnswer    *CorrectAnswer    `json:"correct_answer,omitempty"`
	Explanation      string            `json:"explanation,omitempty"`
	TtsText          string            `json:"tts_text,omitempty"`
	QuestionTopic    string            `json:"question_topic,omitempty"`
	QuestionGrade    *int              `json:"question_grade,omitempty"`
	QuestionLevel    *int              `json:"question_level,omitempty"`
	QuestionSubLevel *int              `json:"question_sub_level,omitempty"`
	QuestionTags     []string          `json:"question_tags,omitempty"`
	// Translations holds the question's text in other languages (keyed
	// "vi" / "en", never Language itself). Stored and shown to admins; a
	// child's paper is served in one language (Localized) without it.
	Translations map[enum.LanguageType]QuestionTranslation `json:"translations,omitempty"`
}

// QuestionTranslation is a question's text in another language: the shape of
// one entry of ma_question_pools.translations. Options carry id + text only
// and name the question's own option ids; ids, images, order and the key
// stay on the question.
type QuestionTranslation struct {
	QuestionText  string           `json:"question_text"`
	TtsText       string           `json:"tts_text,omitempty"`
	Explanation   string           `json:"explanation,omitempty"`
	QuestionTopic string           `json:"question_topic,omitempty"`
	Options       []QuestionOption `json:"options,omitempty"`
}

// CorrectAnswer is the key of one question, the shape of
// ma_question_pools.correct_answer: exactly one part is used, by kind —
// OptionIDs for SINGLE_CHOICE (one id) and MULTIPLE_CHOICE (one or more),
// Value for TRUE_FALSE, Values (accepted numbers) for NUMERIC_INPUT. A new
// kind is a new part here, not a new top-level key in the prompt's output.
type CorrectAnswer struct {
	OptionIDs []string `json:"option_ids,omitempty"`
	Value     *bool    `json:"value,omitempty"`
	Values    []string `json:"values,omitempty"`
}

// UnmarshalJSON reads question_grade as either a number or a band label,
// and question_level / question_sub_level as integers — anything else is
// read as absent, since the server stamps all three afterwards.
// The prompt asks the model for the label ("Mẫu giáo", "Lớp 3") because
// that is the vocabulary the rules are written in; the column is an int.
// A label that is neither is read as absent rather than failing the whole
// round — the server re-stamps question_grade from the question's
// position afterwards and only logs what the model claimed.
func (q *Question) UnmarshalJSON(data []byte) error {
	type plain Question
	aux := struct {
		*plain
		QuestionGrade    json.RawMessage `json:"question_grade,omitempty"`
		QuestionLevel    json.RawMessage `json:"question_level,omitempty"`
		QuestionSubLevel json.RawMessage `json:"question_sub_level,omitempty"`
	}{plain: (*plain)(q)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	q.QuestionGrade = parseGrade(aux.QuestionGrade)
	q.QuestionLevel = parseInt(aux.QuestionLevel)
	q.QuestionSubLevel = parseInt(aux.QuestionSubLevel)
	return nil
}

// parseInt reads an integer, or nil for anything else.
func parseInt(raw json.RawMessage) *int {
	var n int
	if len(raw) == 0 || json.Unmarshal(raw, &n) != nil {
		return nil
	}
	return &n
}

var gradeLabelRe = regexp.MustCompile(`(?i)^\s*(?:lớp|grade)\s*(\d+)\s*$`)

// parseGrade maps the model's question_grade to the stored int, or nil.
func parseGrade(raw json.RawMessage) *int {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return &n
	}
	var label string
	if err := json.Unmarshal(raw, &label); err != nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "mẫu giáo", "mầm non", "kindergarten":
		zero := 0
		return &zero
	}
	if m := gradeLabelRe.FindStringSubmatch(label); m != nil {
		if n, err := utils.StringToIntErr(m[1]); err == nil {
			return &n
		}
	}
	return nil
}

// Answer kinds, the same strings as questionpool.Kind.
//
//   - SINGLE_CHOICE   — Options, exactly one correct option id.
//   - MULTIPLE_CHOICE — Options, one or more correct option ids.
//   - TRUE_FALSE      — two Options (ids "true" and "false") in exams and
//     games; the bank stores none (QuestionFromBank adds them).
//   - NUMERIC_INPUT   — no Options; the child types a number, right when it
//     equals one of CorrectAnswer.Values as a number.
const (
	KindSingleChoice   = "SINGLE_CHOICE"
	KindMultipleChoice = "MULTIPLE_CHOICE"
	KindTrueFalse      = "TRUE_FALSE"
	KindNumericInput   = "NUMERIC_INPUT"
)

// EffectiveKind is the question's kind, SINGLE_CHOICE when none is stored.
func (q Question) EffectiveKind() string {
	if q.QuestionKind == "" {
		return KindSingleChoice
	}
	return q.QuestionKind
}

// TrueOptionID and FalseOptionID are the option ids of a TRUE_FALSE question.
const (
	TrueOptionID  = "true"
	FalseOptionID = "false"
)

// KeyIDs are the option ids the key names: none for a numeric question,
// "true" or "false" for a TRUE_FALSE one.
func (q Question) KeyIDs() []string {
	ca := q.CorrectAnswer
	if ca == nil {
		return nil
	}
	switch q.EffectiveKind() {
	case KindNumericInput:
		return nil
	case KindTrueFalse:
		if ca.Value == nil {
			return nil
		}
		return []string{strconv.FormatBool(*ca.Value)}
	}
	return ca.OptionIDs
}

// KeyOptionIDs is the answer key as one string, the form the answer log
// stores: the correct option ids joined, "" for a numeric question.
func (q Question) KeyOptionIDs() string { return JoinIDs(q.KeyIDs()) }

// KeyText is the displayed text of the answer key: the correct options'
// texts (in option order) or the accepted numbers, joined.
func (q Question) KeyText() string {
	if q.EffectiveKind() == KindNumericInput {
		return strings.Join(q.KeyValues(), ContentSeparator)
	}
	ids := q.KeyIDs()
	var texts []string
	for _, o := range q.Options {
		if slices.Contains(ids, o.ID) {
			texts = append(texts, o.Text)
		}
	}
	return strings.Join(texts, ContentSeparator)
}

// KeyValues are the accepted numbers of a NUMERIC_INPUT question.
func (q Question) KeyValues() []string {
	if q.CorrectAnswer == nil {
		return nil
	}
	return q.CorrectAnswer.Values
}

// Canonical is the question as it is stored and served: the language always
// stated, options never null (a numeric question renders [], not null), only
// the part of the key its kind uses, and no image URL (URLs are presigned per
// response, never stored). Options and key are copies, so a caller may
// reorder, sign or blank them without touching the original.
func (q Question) Canonical() Question {
	q.Language = LanguageOf(q.Language)
	q.QuestionImageURL = ""
	q.Options = OptionsFrom(q.Options)
	q.CorrectAnswer = q.canonicalKey()
	return q
}

func (q Question) canonicalKey() *CorrectAnswer {
	ca := q.CorrectAnswer
	if ca == nil {
		return nil
	}
	switch q.EffectiveKind() {
	case KindNumericInput:
		if len(ca.Values) > 0 {
			return &CorrectAnswer{Values: slices.Clone(ca.Values)}
		}
	case KindTrueFalse:
		if ca.Value != nil {
			v := *ca.Value
			return &CorrectAnswer{Value: &v}
		}
	default:
		if len(ca.OptionIDs) > 0 {
			return &CorrectAnswer{OptionIDs: slices.Clone(ca.OptionIDs)}
		}
	}
	return nil
}

// CanonicalSet is Canonical over a whole set.
func CanonicalSet(qs []Question) []Question {
	out := make([]Question, 0, len(qs))
	for _, q := range qs {
		out = append(out, q.Canonical())
	}
	return out
}

// OptionsFrom is a served option list on the wire: a copy without image
// URLs. It is never nil, so a question without options renders as [] and
// not null.
func OptionsFrom(in []QuestionOption) []QuestionOption {
	out := make([]QuestionOption, 0, len(in))
	for _, o := range in {
		o.ImageURL = ""
		out = append(out, o)
	}
	return out
}

// ContentSeparator joins several option texts, or several accepted numbers,
// into the one string a log column holds.
const ContentSeparator = " | "

// JoinIDs is the stored form of a set of option ids: sorted, comma-joined
// ("a,c"; an id never holds a comma). The log keeps one string per answer
// whatever the kind.
func JoinIDs(ids []string) string {
	sorted := slices.Clone(ids)
	slices.Sort(sorted)
	return strings.Join(sorted, ",")
}

// SplitIDs undoes JoinIDs; "" is no ids.
func SplitIDs(joined string) []string {
	if joined == "" {
		return nil
	}
	return strings.Split(joined, ",")
}

// QuestionAnswer is what a child sent for one exam question: OptionIDs for a
// choice question (one id, or several for MULTIPLE_CHOICE), Value for a
// NUMERIC_INPUT one. Exactly one of the two is set; which one a question
// takes follows its kind. The ids are the ones the question was served
// with: they do not change when the options are shuffled.
type QuestionAnswer struct {
	QuestionNumber int      `json:"question_number"`
	OptionIDs      []string `json:"option_ids,omitempty"`
	Value          string   `json:"value,omitempty"`
}

// Usable reports whether the question can be shown and graded: it has a
// stem (text or image), a correct_answer whose part fits its kind, and,
// unless it is a numeric one, at least two options with distinct non-empty
// ids — a choice key naming options it has (exactly one for single-choice),
// a true/false key holding its value; a numeric one needs its accepted
// values. A question written under an older field vocabulary has no
// correct_answer (or no question_text / option ids), so it is not usable.
func (q Question) Usable() bool {
	if strings.TrimSpace(q.QuestionText) == "" && q.QuestionImageKey == "" {
		return false
	}
	ca := q.CorrectAnswer
	if ca == nil {
		return false
	}
	if q.QuestionKind == KindNumericInput {
		return len(ca.Values) > 0
	}
	if len(q.Options) < 2 {
		return false
	}
	ids := make(map[string]bool, len(q.Options))
	for _, o := range q.Options {
		if o.ID == "" || ids[o.ID] {
			return false
		}
		ids[o.ID] = true
	}
	if q.QuestionKind == KindTrueFalse {
		return ca.Value != nil
	}
	if len(ca.OptionIDs) == 0 {
		return false
	}
	if q.QuestionKind != KindMultipleChoice && len(ca.OptionIDs) != 1 {
		return false
	}
	for _, id := range ca.OptionIDs {
		if !ids[id] {
			return false
		}
	}
	return true
}

// LanguageOf is a stored question's language: a set written before the
// field existed states none and is Vietnamese.
func LanguageOf(l enum.LanguageType) enum.LanguageType {
	if b := l.Base(); b != "" {
		return b
	}
	return enum.LanguageTypeVietnamese
}

// Localized is the question as served in lang: a translation's text over
// the base fields (option text matched by id, so the served order is kept),
// Language set to lang, and Translations dropped either way — a child's
// paper carries one language. A field the translation leaves empty keeps the
// base text, except tts_text: base text read by a voice of another language
// is wrong, so it is left empty. No translation for lang (or lang "" or the
// question's own) = the base question.
func (q Question) Localized(lang enum.LanguageType) Question {
	t, ok := q.Translations[lang.Base()]
	q.Translations = nil
	q.Language = LanguageOf(q.Language)
	if !ok {
		return q
	}
	q.Language = lang.Base()
	q.QuestionText = orBase(t.QuestionText, q.QuestionText)
	q.TtsText = orBase(t.TtsText, q.TtsText)
	q.Explanation = orBase(t.Explanation, q.Explanation)
	q.QuestionTopic = orBase(t.QuestionTopic, q.QuestionTopic)
	if len(t.Options) > 0 {
		textOf := make(map[string]string, len(t.Options))
		for _, o := range t.Options {
			textOf[o.ID] = o.Text
		}
		opts := slices.Clone(q.Options)
		for i := range opts {
			opts[i].Text = orBase(textOf[opts[i].ID], opts[i].Text)
		}
		q.Options = opts
	}
	return q
}

func orBase(tr, base string) string {
	if tr == "" {
		return base
	}
	return tr
}

// CheckSet refuses a question set with an unusable question, naming it.
// Every reader of a stored or generated set runs it, so a set written under
// the older vocabulary — or with a key that names no option — fails loudly
// instead of being served, or graded, as empty or unanswerable questions.
func CheckSet(qs []Question) error {
	if len(qs) == 0 {
		return fmt.Errorf("question: the set is empty")
	}
	for _, q := range qs {
		if !q.Usable() {
			return fmt.Errorf("question: question %d is not usable: it needs a question_text, at least two options with distinct ids, and a correct_answer that fits its kind (option_ids naming its options for a choice question) — a set written with an older field vocabulary? the prompt must emit question_text / options[{id, text}] / correct_answer{option_ids}", q.QuestionNumber)
		}
	}
	return nil
}

// servedQuestions decodes the stored (canonical) set, arranges it the way
// this sitting (or game round: nil shuffle) was served, and — when the key must stay hidden — blanks
// it field by field rather than dropping the question: the child still
// needs the stem and the options to answer.
//
// Each question is served in lang (Localized), so no translation leaves
// here: a child's paper never ships the translated explanations.
func servedQuestions(raw string, sh *Shuffle, includeAnswerKey bool, lang enum.LanguageType) []Question {
	stored := storedQuestions(raw)
	if stored == nil {
		return nil
	}
	out := CanonicalSet(sh.Apply(stored))
	for i := range out {
		out[i] = out[i].Localized(lang)
	}
	if !includeAnswerKey {
		for i := range out {
			out[i].CorrectAnswer = nil
			out[i].Explanation = ""
		}
	}
	return out
}

// storedQuestions decodes a stored set as it is, translations included (the
// admin view), nil when it is empty or unusable.
func storedQuestions(raw string) []Question {
	if raw == "" {
		return nil
	}
	var stored []Question
	if err := json.Unmarshal([]byte(raw), &stored); err != nil || CheckSet(stored) != nil {
		return nil
	}
	for i := range stored {
		stored[i].Language = LanguageOf(stored[i].Language)
	}
	return stored
}
