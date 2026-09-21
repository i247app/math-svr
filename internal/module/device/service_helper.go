package device

import userDomain "math-ai.com/math-ai/internal/domain/user"

func (s *Service) isDemoUser(user *userDomain.User) bool {
	for _, name := range s.demoNames {
		if user.Phone() == name || (user.Email() != nil && *user.Email() == name) {
			return true
		}
	}
	return false
}
