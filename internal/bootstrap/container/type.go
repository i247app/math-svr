package container

import (
	gradeDomain "math-ai.com/math-ai/internal/domain/grade"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	programDomain "math-ai.com/math-ai/internal/domain/program"
	semesterDomain "math-ai.com/math-ai/internal/domain/semester"
	userDomain "math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/module/grade"
	"math-ai.com/math-ai/internal/module/profile"
	"math-ai.com/math-ai/internal/module/program"
	"math-ai.com/math-ai/internal/module/semester"
	"math-ai.com/math-ai/internal/module/user"
)

type ServiceContainer struct {
	UserSvc     *user.Service
	ProgramSvc  *program.Service
	GradeSvc    *grade.Service
	SemesterSvc *semester.Service
	ProfileSvc  *profile.Service
}

type RepositoryContainer struct {
	UserRepository     userDomain.IRepository
	ProgramRepository  programDomain.IRepository
	GradeRepository    gradeDomain.IRepository
	SemesterRepository semesterDomain.IRepository
	ProfileRepository  profileDomain.IRepository
}
