package classroomexercisesubmission

import (
	domainBot "math-ai.com/math-ai/internal/domain/bot"
	"math-ai.com/math-ai/internal/shared/enum"
)

// normalizeLanguage maps the enum.LanguageType to the bot domain's
// QuizLanguage. Default is Vietnamese to match the project-wide
// default in errs.NewError.
func normalizeLanguage(lang enum.LanguageType) string {
	if lang == enum.LanguageTypeEnglish {
		return string(domainBot.QuizLanguageEnglish)
	}
	return string(domainBot.QuizLanguageVietnamese)
}
