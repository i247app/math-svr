package home

import (
	dto "math-ai.com/math-ai/internal/application/dto/home"
	query "math-ai.com/math-ai/internal/application/query/home"
	"math-ai.com/math-ai/internal/shared/enum"
)

// quizCards maps the acting profile's standalone quiz history into slim
// cards. No storage signing needed — quizzes carry no images.
func quizCards(data *query.HomeLayoutData) []*dto.QuizCard {
	cards := make([]*dto.QuizCard, 0, len(data.Quizzes))
	for _, q := range data.Quizzes {
		cards = append(cards, dto.QuizToCard(q))
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
