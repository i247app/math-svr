package container

import (
	userDomain "math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/module/user"
)

type ServiceContainer struct {
	UserSvc *user.Service
}

type RepositoryContainer struct {
	UserRepository userDomain.IRepository
}
