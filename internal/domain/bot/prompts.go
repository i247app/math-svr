package bot

import (
	"fmt"
	"strings"
)

// QuizPurpose discriminates the four supported quiz LLM calls. Each
// purpose maps to a distinct (system, user) template pair so the model
// is steered explicitly toward generation, grading, or remediation.
type QuizPurpose int

const (
	// QuizPurposeGenerate asks the model to produce a fresh quiz from
	// curriculum context (grade, semester, program).
	QuizPurposeGenerate QuizPurpose = iota + 1
	// QuizPurposeReinforce asks the model to produce a remedial quiz
	// targeting the student's weak spots from a prior quiz.
	QuizPurposeReinforce
	// QuizPurposeGrade asks the model to score the student's answers
	// to a fresh quiz and (for ASSESSMENT) predict ai_detect_grade.
	QuizPurposeGrade
	// QuizPurposeGradeReinforce grades a reinforce quiz; the prompt
	// carries the student's currently configured grade as context so
	// the prediction reflects progression, not just raw accuracy.
	QuizPurposeGradeReinforce
)

// QuizLanguage is the response language requested from the model. The
// model is instructed to emit ai_review in this language; JSON keys are
// always English regardless.
type QuizLanguage string

const (
	QuizLanguageVietnamese QuizLanguage = "vn"
	QuizLanguageEnglish    QuizLanguage = "en"
)

// QuizType narrows the prompt's tone and grading shape. EXAM is reserved
// for a future template set; today only ASSESSMENT and PRACTICE are
// honoured by the builder.
type QuizType string

const (
	QuizTypeAssessment QuizType = "ASSESSMENT"
	QuizTypePractice   QuizType = "PRACTICE"
)

// QuizPromptInput is the union of fields any quiz purpose may need.
// Each purpose pulls only the subset it requires; unused fields are
// ignored. BuildQuizPrompt validates required-field presence per
// purpose so the model never sees a partially-filled template.
type QuizPromptInput struct {
	// Language and Type are required for every purpose.
	Language QuizLanguage
	Type     QuizType

	// Curriculum context — required for Generate / Reinforce.
	Grade    string
	Semester string
	Program  string

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
// requested purpose. The caller drops them into ChatRequest.Messages
// as Role=system and Role=user respectively, sets JSONMode=true and a
// low Temperature for determinism, and forwards through the adapter.
func BuildQuizPrompt(purpose QuizPurpose, in QuizPromptInput) (system string, user string, err error) {
	lang, err := normalizeLanguage(in.Language)
	if err != nil {
		return "", "", err
	}
	typ, err := normalizeQuizType(in.Type)
	if err != nil {
		return "", "", err
	}

	fmt.Println("Language: ", lang)

	switch purpose {
	case QuizPurposeGenerate:
		// Curriculum context (Grade/Semester/Program) is OPTIONAL. The
		// user prompt adapts to render only the lines that are populated,
		// so the model still gets a coherent brief on partial / empty
		// context.
		return buildGeneratePrompt(lang, typ, in), buildGenerateUser(lang, typ, in), nil
	case QuizPurposeReinforce:
		// Previous-round payloads are required (otherwise the bot can't
		// know what to reinforce); curriculum context stays optional.
		if err := requireFields(
			in.PreviousQuestions, "PreviousQuestions",
			in.PreviousAnswers, "PreviousAnswers",
			in.PreviousAIReview, "PreviousAIReview"); err != nil {
			return "", "", err
		}
		return buildReinforcePrompt(lang, typ, in), buildReinforceUser(lang, typ, in), nil
	case QuizPurposeGrade:
		if err := requireFields(in.Questions, "Questions", in.Answers, "Answers"); err != nil {
			return "", "", err
		}
		return buildGradePrompt(lang, typ), buildGradeUser(lang, typ, in), nil
	case QuizPurposeGradeReinforce:
		// CurrentGrade is optional — when unknown, the grading prompt
		// falls back to "unknown" so the model still produces a result.
		if err := requireFields(in.Questions, "Questions", in.Answers, "Answers"); err != nil {
			return "", "", err
		}
		return buildGradeReinforcePrompt(lang, typ), buildGradeReinforceUser(lang, typ, in), nil
	default:
		return "", "", fmt.Errorf("bot: unknown quiz purpose %d", purpose)
	}
}

func normalizeLanguage(lang QuizLanguage) (QuizLanguage, error) {
	switch strings.ToLower(strings.TrimSpace(string(lang))) {
	case "vn":
		return QuizLanguageVietnamese, nil
	case "", "en":
		return QuizLanguageEnglish, nil
	default:
		return "", fmt.Errorf("bot: unsupported language %q", string(lang))
	}
}

func normalizeQuizType(t QuizType) (QuizType, error) {
	switch QuizType(strings.ToUpper(strings.TrimSpace(string(t)))) {
	case QuizTypeAssessment:
		return QuizTypeAssessment, nil
	case QuizTypePractice:
		return QuizTypePractice, nil
	default:
		return "", fmt.Errorf("bot: unsupported quiz type %q", string(t))
	}
}

// requireFields takes alternating (value, name) pairs and returns the
// first missing-field error encountered. Keeps the per-purpose validation
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

func buildGeneratePrompt(lang QuizLanguage, _ QuizType, _ QuizPromptInput) string {
	if lang == QuizLanguageEnglish {
		return systemGenerateEN
	}
	return systemGenerateVN
}

func buildGenerateUser(lang QuizLanguage, typ QuizType, in QuizPromptInput) string {
	if lang == QuizLanguageEnglish {
		return userGenerateEN(typ, in)
	}
	return userGenerateVN(typ, in)
}

func buildReinforcePrompt(lang QuizLanguage, _ QuizType, _ QuizPromptInput) string {
	if lang == QuizLanguageEnglish {
		return systemReinforceEN
	}
	return systemReinforceVN
}

func buildReinforceUser(lang QuizLanguage, typ QuizType, in QuizPromptInput) string {
	if lang == QuizLanguageEnglish {
		return userReinforceEN(typ, in)
	}
	return userReinforceVN(typ, in)
}

func buildGradePrompt(lang QuizLanguage, typ QuizType) string {
	switch lang {
	case QuizLanguageEnglish:
		if typ == QuizTypePractice {
			return systemGradePracticeEN
		}
		return systemGradeAssessmentEN
	default:
		if typ == QuizTypePractice {
			return systemGradePracticeVN
		}
		return systemGradeAssessmentVN
	}
}

func buildGradeUser(lang QuizLanguage, typ QuizType, in QuizPromptInput) string {
	if lang == QuizLanguageEnglish {
		return userGradeEN(typ, in)
	}
	return userGradeVN(typ, in)
}

func buildGradeReinforcePrompt(lang QuizLanguage, typ QuizType) string {
	switch lang {
	case QuizLanguageEnglish:
		if typ == QuizTypePractice {
			return systemGradeReinforcePracticeEN
		}
		return systemGradeReinforceAssessmentEN
	default:
		if typ == QuizTypePractice {
			return systemGradeReinforcePracticeVN
		}
		return systemGradeReinforceAssessmentVN
	}
}

func buildGradeReinforceUser(lang QuizLanguage, typ QuizType, in QuizPromptInput) string {
	if lang == QuizLanguageEnglish {
		return userGradeReinforceEN(typ, in)
	}
	return userGradeReinforceVN(typ, in)
}
