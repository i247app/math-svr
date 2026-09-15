package home

import (
	dto "math-ai.com/math-ai/internal/application/dto/home"
	query "math-ai.com/math-ai/internal/application/query/home"
	"math-ai.com/math-ai/internal/shared/enum"
)

// examCards maps the acting profile's exam history into slim cards. No
// storage signing needed — exams carry no images.
func examCards(data *query.HomeLayoutData) []*dto.ExamCard {
	cards := make([]*dto.ExamCard, 0, len(data.Exams))
	for _, a := range data.Exams {
		cards = append(cards, dto.ExamToCard(a, data.ExamAiExams[a.AiExamId()]))
	}
	return cards
}

func isSupportedRole(role string) bool {
	switch enum.RoleType(role) {
	case enum.RoleTypeTeacher, enum.RoleTypeParent, enum.RoleTypeStudent:
		return true
	default:
		return false
	}
}
