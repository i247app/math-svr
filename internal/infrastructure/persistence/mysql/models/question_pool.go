package models

type QuestionPoolModel struct {
	Audit

	QuestionPoolId   int64
	QuestionKind     string
	QuestionText     string
	QuestionImageKey *string
	Options          *string // JSON [{"id","text","image_key"?}]; NULL for kinds with no options
	CorrectAnswer    string  // JSON {"option_ids":[...]} or {"value":bool}
	Explanation      *string
	TtsText          *string
	QuestionLanguage string
	Translations     *string // JSON {"<lang>": {...}}; NULL = none
	QuestionGrade    int
	QuestionTopic    *string
	QuestionLevel    *int
	QuestionSubLevel *int
	QuestionTags     string // JSON array of strings
	QuestionSource   string
	VerifiedCount    int
	QuestionHash     string
	QuestionStatus   *string
}
