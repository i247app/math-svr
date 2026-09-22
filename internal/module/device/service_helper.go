package device

import (
	userDomain "math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/utils"
)

func (s *Service) isDemoUser(user *userDomain.User) bool {
	for _, name := range s.demoNames {
		if utils.DerefString(user.Phone()) == name || (user.Email() != nil && *user.Email() == name) {
			return true
		}
	}
	return false
}
