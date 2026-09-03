package bot

import (
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/shared/enum"
)

// QuizPromptKind discriminates the four supported quiz LLM calls. Each
// kind maps to a distinct (system, user) template pair so the model is
// steered explicitly toward generation, grading, or remediation.
//
// Naming note: this concept used to be called QuizPurpose. It was
// renamed to QuizPromptKind once the quiz domain itself adopted "purpose"
// as the name of the ASSESSMENT / PRACTICE / EXAM dimension persisted on
// ma_quizzes.purpose.
type QuizPromptKind int

const (
	// QuizPromptKindGenerate asks the model to produce a fresh quiz from
	// curriculum context (grade, semester, program).
	QuizPromptKindGenerate QuizPromptKind = iota + 1
	// QuizPromptKindReinforce asks the model to produce a remedial quiz
	// targeting the student's weak spots from a prior quiz.
	QuizPromptKindReinforce
	// QuizPromptKindGrade asks the model to score the student's answers
	// to a fresh quiz and (for ASSESSMENT) predict assessment_grade.
	QuizPromptKindGrade
	// QuizPromptKindGradeReinforce grades a reinforce quiz; the prompt
	// carries the student's currently configured grade as context so the
	// prediction reflects progression, not just raw accuracy.
	QuizPromptKindGradeReinforce
)

// QuizLanguage is the response language requested from the model. The
// model is instructed to emit ai_review in this language; JSON keys are
// always English regardless.
type QuizLanguage string

const (
	QuizLanguageVietnamese QuizLanguage = "vn"
	QuizLanguageEnglish    QuizLanguage = "en"
)

// QuizPurpose narrows the prompt's tone and grading shape — the same
// vocabulary persisted on ma_quizzes.purpose. EXAM is reserved for a
// future template set; today only ASSESSMENT and PRACTICE are honoured
// by the builder (EXAM falls back to the ASSESSMENT templates).
type QuizPurpose string

const (
	QuizPurposeAssessment QuizPurpose = "ASSESSMENT"
	QuizPurposePractice   QuizPurpose = "PRACTICE"
	QuizPurposeExam       QuizPurpose = "EXAM"
)

// QuizTypeOfQuiz captures the *learning intent* persisted on
// ma_quizzes.type_of_quiz. GENERAL = the quiz introduces new knowledge
// from the curriculum; REINFORCEMENT = the quiz consolidates / reviews
// material the student previously got wrong. Both values are passed to
// the prompt so the model can adapt its phrasing (e.g. emphasise
// targeted practice for REINFORCEMENT) without us having to fork the
// system prompt by row state.
type QuizTypeOfQuiz string

const (
	QuizTypeOfQuizGeneral       QuizTypeOfQuiz = "GENERAL"
	QuizTypeOfQuizReinforcement QuizTypeOfQuiz = "REINFORCEMENT"
)

// QuizPromptInput is the union of fields any prompt kind may need.
// Each kind pulls only the subset it requires; unused fields are
// ignored. BuildQuizPrompt validates required-field presence per kind
// so the model never sees a partially-filled template.
type QuizPromptInput struct {
	// Language, Purpose, and TypeOfQuiz are accepted by every kind.
	// Purpose may be blank — when so, the builder falls back to a
	// generic elementary-level brief. TypeOfQuiz defaults to GENERAL
	// when blank.
	Language   QuizLanguage
	Purpose    QuizPurpose
	TypeOfQuiz QuizTypeOfQuiz

	// Curriculum context — required for Generate / Reinforce.
	Grade    string
	Semester string
	Program  string

	// Chapters scopes the generated quiz to specific curriculum units. The
	// caller supplies pre-formatted, ready-to-render descriptions (one per
	// chapter); the template renders them as a sub-list under the
	// curriculum context. Empty means "no chapter preference" — the prompt
	// falls back to whatever Grade/Semester/Program already imply.
	ChapterDescriptions []string

	// NumQuestions controls how many MCQ items Generate / Reinforce ask
	// for. Zero or negative falls back to defaultNumQuestions inside the
	// template builder so callers don't have to defend every entry point.
	NumQuestions int

	// Previous-round payloads — required for Reinforce.
	PreviousQuestions string
	PreviousAnswers   string
	PreviousAIReview  string

	// Current-round payloads — required for Grade / GradeReinforce.
	Questions string
	Answers   string

	// Student's currently configured grade — required for
	// GradeReinforce so the model can reason about progression.
	CurrentGrade string
}

// BuildQuizPrompt returns the (system, user) message contents for the
// requested prompt kind. The caller drops them into ChatRequest.Messages
// as Role=system and Role=user respectively, sets JSONMode=true and a
// low Temperature for determinism, and forwards through the adapter.
func BuildQuizPrompt(kind QuizPromptKind, in QuizPromptInput) (system string, user string, err error) {
	lang, err := normalizeLanguage(in.Language)
	if err != nil {
		return "", "", err
	}
	purpose, err := normalizeQuizPurpose(in.Purpose)
	if err != nil {
		return "", "", err
	}
	typeOfQuiz := normalizeQuizTypeOfQuiz(in.TypeOfQuiz)
	in.Purpose = purpose
	in.TypeOfQuiz = typeOfQuiz

	switch kind {
	case QuizPromptKindGenerate:
		// Curriculum context (Grade/Semester/Program) is OPTIONAL. The
		// user prompt adapts to render only the lines that are populated,
		// so the model still gets a coherent brief on partial / empty
		// context.
		return buildGeneratePrompt(lang, purpose, in), buildGenerateUser(lang, purpose, in), nil
	case QuizPromptKindReinforce:
		// Previous-round payloads are required (otherwise the bot can't
		// know what to reinforce); curriculum context stays optional.
		if err := requireFields(
			in.PreviousQuestions, "PreviousQuestions",
			in.PreviousAnswers, "PreviousAnswers",
			in.PreviousAIReview, "PreviousAIReview"); err != nil {
			return "", "", err
		}
		return buildReinforcePrompt(lang, purpose, in), buildReinforceUser(lang, purpose, in), nil
	case QuizPromptKindGrade:
		if err := requireFields(in.Questions, "Questions", in.Answers, "Answers"); err != nil {
			return "", "", err
		}
		return buildGradePrompt(lang, purpose), buildGradeUser(lang, purpose, in), nil
	case QuizPromptKindGradeReinforce:
		// CurrentGrade is optional — when unknown, the grading prompt
		// falls back to "unknown" so the model still produces a result.
		if err := requireFields(in.Questions, "Questions", in.Answers, "Answers"); err != nil {
			return "", "", err
		}
		return buildGradeReinforcePrompt(lang, purpose), buildGradeReinforceUser(lang, purpose, in), nil
	default:
		return "", "", fmt.Errorf("bot: unknown quiz prompt kind %d", kind)
	}
}

func normalizeLanguage(lang QuizLanguage) (QuizLanguage, error) {
	switch strings.ToLower(strings.TrimSpace(string(lang))) {
	case string(enum.LanguageTypeVietnamese), string(enum.LanguageTypeVietnameseV2):
		return QuizLanguageVietnamese, nil
	case string(enum.LanguageTypeEnglish), string(enum.LanguageTypeEnglishV2):
		return QuizLanguageEnglish, nil
	default:
		return "", fmt.Errorf("bot: unsupported language %q", string(lang))
	}
}

// normalizeQuizPurpose accepts an empty value (treated as ASSESSMENT so
// the existing prompt tone still applies) and rejects unknown values
// loudly — drift from the enum should not silently degrade to a default.
func normalizeQuizPurpose(p QuizPurpose) (QuizPurpose, error) {
	switch QuizPurpose(strings.ToUpper(strings.TrimSpace(string(p)))) {
	case "", QuizPurposeAssessment:
		return QuizPurposeAssessment, nil
	case QuizPurposePractice:
		return QuizPurposePractice, nil
	case QuizPurposeExam:
		// EXAM has no dedicated template yet; reuse the ASSESSMENT shape
		// (closest existing tone) so the request still produces a quiz.
		return QuizPurposeAssessment, nil
	default:
		return "", fmt.Errorf("bot: unsupported quiz purpose %q", string(p))
	}
}

// normalizeQuizTypeOfQuiz is tolerant: it never errors. Blank or unknown
// values fall back to GENERAL so a partially-populated request still
// produces a usable prompt — the column has a DB-level default of
// 'GENERAL' for the same reason.
func normalizeQuizTypeOfQuiz(t QuizTypeOfQuiz) QuizTypeOfQuiz {
	switch QuizTypeOfQuiz(strings.ToUpper(strings.TrimSpace(string(t)))) {
	case QuizTypeOfQuizReinforcement:
		return QuizTypeOfQuizReinforcement
	default:
		return QuizTypeOfQuizGeneral
	}
}

// requireFields takes alternating (value, name) pairs and returns the
// first missing-field error encountered. Keeps the per-kind validation
// readable at the call site.
func requireFields(pairs ...string) error {
	if len(pairs)%2 != 0 {
		return fmt.Errorf("bot: requireFields expects (value, name) pairs")
	}
	for i := 0; i < len(pairs); i += 2 {
		if strings.TrimSpace(pairs[i]) == "" {
			return fmt.Errorf("bot: %s is required", pairs[i+1])
		}
	}
	return nil
}

// defaultNumQuestions is the fallback used when callers leave
// QuizPromptInput.NumQuestions at zero. Kept in sync with the
// module/quiz validator's DefaultNumQuestions (5), but duplicated here
// so the domain prompt builder remains usable without going through the
// module-layer validator.
const defaultNumQuestions = 10

// maxPromptChapterLabelLen caps each rendered chapter line so an
// over-long descriptor cannot bloat the prompt. The cap is generous
// enough to fit the longest real-world chapter name we ship today and
// still leave headroom; anything past it is truncated with an ellipsis.
const maxPromptChapterLabelLen = 120

// cleanChapterLabels trims each label, drops empties, deduplicates by
// case-insensitive match (so EN/VN duplicates from a half-translated
// curriculum don't render twice), and truncates over-long entries. The
// callers (EN/VN template builders) treat an empty slice as "no chapter
// preference" and skip rendering, so this is the single normalization
// seam for both languages.
func cleanChapterLabels(chapters []string) []string {
	if len(chapters) == 0 {
		return nil
	}
	out := make([]string, 0, len(chapters))
	seen := make(map[string]struct{}, len(chapters))
	for _, raw := range chapters {
		label := strings.TrimSpace(raw)
		if label == "" {
			continue
		}
		if len(label) > maxPromptChapterLabelLen {
			label = label[:maxPromptChapterLabelLen-1] + "…"
		}
		key := strings.ToLower(label)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, label)
	}
	return out
}

func resolveNumQuestions(n int) int {
	if n <= 0 {
		return defaultNumQuestions
	}
	return n
}

func buildGeneratePrompt(lang QuizLanguage, _ QuizPurpose, in QuizPromptInput) string {
	n := resolveNumQuestions(in.NumQuestions)
	if lang == QuizLanguageEnglish {
		return buildSystemGenerateEN(n)
	}
	return buildSystemGenerateVN(n)
}

func buildGenerateUser(lang QuizLanguage, purpose QuizPurpose, in QuizPromptInput) string {
	if lang == QuizLanguageEnglish {
		return userGenerateEN(purpose, in)
	}
	return userGenerateVN(purpose, in)
}

func buildReinforcePrompt(lang QuizLanguage, _ QuizPurpose, in QuizPromptInput) string {
	n := resolveNumQuestions(in.NumQuestions)
	if lang == QuizLanguageEnglish {
		return buildSystemReinforceEN(n)
	}
	return buildSystemReinforceVN(n)
}

func buildReinforceUser(lang QuizLanguage, purpose QuizPurpose, in QuizPromptInput) string {
	if lang == QuizLanguageEnglish {
		return userReinforceEN(purpose, in)
	}
	return userReinforceVN(purpose, in)
}

func buildGradePrompt(lang QuizLanguage, purpose QuizPurpose) string {
	switch lang {
	case QuizLanguageEnglish:
		if purpose == QuizPurposePractice {
			return systemGradePracticeEN
		}
		return systemGradeAssessmentEN
	default:
		if purpose == QuizPurposePractice {
			return systemGradePracticeVN
		}
		return systemGradeAssessmentVN
	}
}

func buildGradeUser(lang QuizLanguage, purpose QuizPurpose, in QuizPromptInput) string {
	if lang == QuizLanguageEnglish {
		return userGradeEN(purpose, in)
	}
	return userGradeVN(purpose, in)
}

func buildGradeReinforcePrompt(lang QuizLanguage, purpose QuizPurpose) string {
	switch lang {
	case QuizLanguageEnglish:
		if purpose == QuizPurposePractice {
			return systemGradeReinforcePracticeEN
		}
		return systemGradeReinforceAssessmentEN
	default:
		if purpose == QuizPurposePractice {
			return systemGradeReinforcePracticeVN
		}
		return systemGradeReinforceAssessmentVN
	}
}

func buildGradeReinforceUser(lang QuizLanguage, purpose QuizPurpose, in QuizPromptInput) string {
	if lang == QuizLanguageEnglish {
		return userGradeReinforceEN(purpose, in)
	}
	return userGradeReinforceVN(purpose, in)
}
