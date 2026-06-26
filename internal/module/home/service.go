package home

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	dto "math-ai.com/math-ai/internal/application/dto/home"
	query "math-ai.com/math-ai/internal/application/query/home"
	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	exerciseDomain "math-ai.com/math-ai/internal/domain/exercise"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	quizDomain "math-ai.com/math-ai/internal/domain/quiz"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// presignTTL bounds avatar / cover URL validity. Mirrors the classroom
// module's coverUrlTTL so cross-aggregate UX stays consistent.
const presignTTL = 1 * time.Hour

// Service is the home module's public façade. It resolves and
// session-validates the acting profile, delegates the cross-aggregate
// read to the query handler, and maps the domain assembly into the
// role-discriminated DTO (signing avatar / cover URLs along the way).
type Service struct {
	getHomeLayoutQuery *query.GetHomeLayoutQueryHandler
	profileRepo        profileDomain.IRepository
	storageProvider    *storage.Adapter
}

func NewService(
	classroomRepo classroomDomain.IRepository,
	memberRepo classroomDomain.IMemberRepository,
	exerciseRepo exerciseDomain.IRepository,
	submissionRepo exerciseDomain.ISubmissionRepository,
	profileRepo profileDomain.IRepository,
	quizRepo quizDomain.IRepository,
	storageProvider *storage.Adapter,
) *Service {
	return &Service{
		getHomeLayoutQuery: query.NewGetHomeLayoutQueryHandler(
			classroomRepo, memberRepo, exerciseRepo, submissionRepo, profileRepo, quizRepo,
		),
		profileRepo:     profileRepo,
		storageProvider: storageProvider,
	}
}

// GetHomeLayout builds the home dashboard for the profile named in the
// request, after confirming it belongs to the authenticated user.
func (s *Service) GetHomeLayout(ctx context.Context, req *dto.HomeLayoutReq, sessionUserID int64) (*dto.HomeLayoutRes, error) {
	if err := ValidateHomeLayout(ctx, req); err != nil {
		return nil, err
	}

	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if !isSupportedRole(caller.Role()) {
		return nil, errs.NewError(ctx, status.HOME_UNSUPPORTED_ROLE, nil, ErrUnsupportedRole)
	}

	data, err := s.getHomeLayoutQuery.Handle(ctx, query.GetHomeLayoutQuery{Profile: caller})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	layout := s.buildLayout(ctx, caller, data)
	return &dto.HomeLayoutRes{Home: layout}, nil
}
