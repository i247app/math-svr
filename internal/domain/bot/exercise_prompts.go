package bot

import (
	"fmt"
	"strings"
)

// ExercisePromptKind enumerates the exercise-side LLM calls.
//
//	Generate — teacher-issued exercise generation.
//	Grade    — student-submission grading; reads `{questions, answers}`,
//	           emits the standard `{total_questions, correct_number,
//	           score_percentage, ai_review}` shape so the existing
//	           parseGradedQuiz parser handles the response unchanged.
type ExercisePromptKind int

const (
	ExercisePromptKindGenerate ExercisePromptKind = iota + 1
	ExercisePromptKindGrade
)

// ExercisePromptInput is the union of fields the exercise prompts may
// consume. Unlike quizzes (which derive their topic from curriculum
// rows), exercises are anchored on teacher-supplied chapter + lesson
// names, with grade + program added when the parent classroom has them
// pinned. Grading reuses Questions / Answers as the JSON blobs being
// scored.
type ExercisePromptInput struct {
	Language     QuizLanguage
	Grade        string
	Program      string
	ChapterName  string
	LessonName   string
	NumQuestions int
	// Questions / Answers carry the raw JSON the grading prompt operates
	// on. Both are required for ExercisePromptKindGrade and are ignored
	// for Generate.
	Questions string
	Answers   string
}

// BuildExercisePrompt returns the (system, user) message contents for
// the requested exercise prompt kind. ChapterName + LessonName are
// required because they are the teacher's explicit topic scope; without
// them the exercise has no anchor and the bot would degrade to a generic
// curriculum-only generation.
func BuildExercisePrompt(kind ExercisePromptKind, in ExercisePromptInput) (system string, user string, err error) {
	lang, err := normalizeLanguage(in.Language)
	if err != nil {
		return "", "", err
	}

	switch kind {
	case ExercisePromptKindGenerate:
		if err := requireFields(
			in.ChapterName, "ChapterName",
			in.LessonName, "LessonName"); err != nil {
			return "", "", err
		}
		return buildExerciseGeneratePrompt(lang, in), buildExerciseGenerateUser(lang, in), nil
	case ExercisePromptKindGrade:
		if err := requireFields(
			in.Questions, "Questions",
			in.Answers, "Answers"); err != nil {
			return "", "", err
		}
		return buildExerciseGradePrompt(lang), buildExerciseGradeUser(lang, in), nil
	default:
		return "", "", fmt.Errorf("bot: unknown exercise prompt kind %d", kind)
	}
}

func buildExerciseGradePrompt(lang QuizLanguage) string {
	if lang == QuizLanguageEnglish {
		return systemExerciseGradeEN
	}
	return systemExerciseGradeVN
}

func buildExerciseGradeUser(lang QuizLanguage, in ExercisePromptInput) string {
	if lang == QuizLanguageEnglish {
		return userExerciseGradeEN(in)
	}
	return userExerciseGradeVN(in)
}

func buildExerciseGeneratePrompt(lang QuizLanguage, in ExercisePromptInput) string {
	n := resolveNumQuestions(in.NumQuestions)
	if lang == QuizLanguageEnglish {
		return buildSystemExerciseGenerateEN(n)
	}
	return buildSystemExerciseGenerateVN(n)
}

func buildExerciseGenerateUser(lang QuizLanguage, in ExercisePromptInput) string {
	if lang == QuizLanguageEnglish {
		return userExerciseGenerateEN(in)
	}
	return userExerciseGenerateVN(in)
}

// buildExerciseContextEN emits the optional grade / program lines. The
// chapter and lesson go in the dedicated SCOPE block of the user prompt,
// not here.
func buildExerciseContextEN(in ExercisePromptInput) string {
	var b strings.Builder
	if v := strings.TrimSpace(in.Grade); v != "" {
		fmt.Fprintf(&b, "- Grade: %s\n", v)
	}
	if v := strings.TrimSpace(in.Program); v != "" {
		fmt.Fprintf(&b, "- Curriculum: %s\n", v)
	}
	return strings.TrimRight(b.String(), "\n")
}

func buildExerciseContextVN(in ExercisePromptInput) string {
	var b strings.Builder
	if v := strings.TrimSpace(in.Grade); v != "" {
		fmt.Fprintf(&b, "- Lớp: %s\n", v)
	}
	if v := strings.TrimSpace(in.Program); v != "" {
		fmt.Fprintf(&b, "- Chương trình học: %s\n", v)
	}
	return strings.TrimRight(b.String(), "\n")
}
