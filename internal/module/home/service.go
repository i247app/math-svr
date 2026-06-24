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
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
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
	storageProvider *storage.Adapter,
) *Service {
	return &Service{
		getHomeLayoutQuery: query.NewGetHomeLayoutQueryHandler(
			classroomRepo, memberRepo, exerciseRepo, submissionRepo, profileRepo,
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

// resolveActingProfile loads the profile and confirms it belongs to the
// session user — the same ownership contract the classroom module
// enforces. A zero sessionUserID skips the ownership check (internal
// callers / tests only).
func (s *Service) resolveActingProfile(ctx context.Context, profileID, sessionUserID int64) (*profileDomain.Profile, error) {
	p, err := s.profileRepo.FindByProfileId(ctx, profileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if p == nil {
		return nil, errs.NewError(ctx, status.HOME_PROFILE_NOT_FOUND, nil, ErrProfileNotFound)
	}
	if sessionUserID != 0 && sessionUserID != p.UserId() {
		return nil, errs.NewError(ctx, status.HOME_PROFILE_NOT_OWNED, nil, ErrProfileNotOwnedByUser)
	}
	return p, nil
}

// buildLayout maps the domain assembly into the wire DTO, populating only
// the role-specific section that matches the acting profile.
func (s *Service) buildLayout(ctx context.Context, caller *profileDomain.Profile, data *query.HomeLayoutData) *dto.HomeLayout {
	profileSummary := dto.ProfileToSummary(caller)
	s.signProfileAvatar(ctx, profileSummary)

	layout := &dto.HomeLayout{
		Profile: profileSummary,
		Role:    data.Role,
	}

	switch enum.RoleType(data.Role) {
	case enum.RoleTypeTeacher:
		layout.Teacher = &dto.TeacherLayout{
			Classrooms:        s.classroomCards(ctx, data),
			AssignedExercises: s.exerciseCards(data),
		}
	case enum.RoleTypeStudent:
		layout.Student = &dto.StudentLayout{
			Classrooms:       s.classroomCards(ctx, data),
			PendingExercises: s.exerciseCards(data),
		}
	case enum.RoleTypeParent:
		layout.Parent = &dto.ParentLayout{
			Children:          s.childSummaries(ctx, data),
			Classrooms:        s.classroomCards(ctx, data),
			PendingExercises:  s.parentPendingCards(ctx, data),
			RecentCompletions: s.completionCards(ctx, data),
		}
	}
	return layout
}

// classroomCards maps the loaded classrooms into dashboard tiles, tagging
// each with the relevant profile's member role (and, for parents, the
// enrolled child's profile_id) and signing the cover URL.
func (s *Service) classroomCards(ctx context.Context, data *query.HomeLayoutData) []*dto.ClassroomCard {
	cards := make([]*dto.ClassroomCard, 0, len(data.Classrooms))
	for _, c := range data.Classrooms {
		var myRole *string
		if role, ok := data.RoleByClassroom[c.ClassroomId()]; ok {
			r := role
			myRole = &r
		}
		var memberProfileID *int64
		if pid, ok := data.MemberProfileByClassroom[c.ClassroomId()]; ok {
			id := pid
			memberProfileID = &id
		}
		card := dto.ClassroomToCard(c, myRole, memberProfileID)
		s.signCoverURL(ctx, card)
		cards = append(cards, card)
	}
	return cards
}

// exerciseCards maps the teacher/student exercise list into slim cards,
// attaching each exercise's classroom ref from the batched classroom map.
func (s *Service) exerciseCards(data *query.HomeLayoutData) []*dto.ExerciseCard {
	cards := make([]*dto.ExerciseCard, 0, len(data.Exercises))
	for _, e := range data.Exercises {
		card := dto.ExerciseToCard(e)
		if c, ok := data.ClassroomByID[e.ClassroomId()]; ok {
			card.Classroom = dto.ClassroomToRef(c)
		}
		cards = append(cards, card)
	}
	return cards
}

// childSummaries maps the parent's children into slim profile cards and
// signs each avatar.
func (s *Service) childSummaries(ctx context.Context, data *query.HomeLayoutData) []*dto.ProfileSummary {
	cards := make([]*dto.ProfileSummary, 0, len(data.Children))
	for _, c := range data.Children {
		summary := dto.ProfileToSummary(c)
		s.signProfileAvatar(ctx, summary)
		cards = append(cards, summary)
	}
	return cards
}

// completionCards maps the parent's recent-completion feed, hydrating the
// child, exercise, and classroom refs from the batched lookup maps.
func (s *Service) completionCards(ctx context.Context, data *query.HomeLayoutData) []*dto.CompletedExerciseCard {
	cards := make([]*dto.CompletedExerciseCard, 0, len(data.Submissions))
	for _, sub := range data.Submissions {
		card := dto.SubmissionToCompletedCard(sub)
		if child, ok := data.ProfileByID[sub.ProfileId()]; ok {
			summary := dto.ProfileToSummary(child)
			s.signProfileAvatar(ctx, summary)
			card.Child = summary
		}
		if ex, ok := data.ExerciseByID[sub.ClassroomExerciseId()]; ok {
			card.Exercise = dto.ExerciseToCard(ex)
		}
		if c, ok := data.ClassroomByID[sub.ClassroomId()]; ok {
			card.Classroom = dto.ClassroomToRef(c)
		}
		cards = append(cards, card)
	}
	return cards
}

// parentPendingCards maps the parent's per-child "not completed yet" feed,
// hydrating the child profile and classroom refs from the batched lookup
// maps. Mirrors completionCards so the to-do and done feeds share shape.
func (s *Service) parentPendingCards(ctx context.Context, data *query.HomeLayoutData) []*dto.ParentPendingExerciseCard {
	cards := make([]*dto.ParentPendingExerciseCard, 0, len(data.PendingExercises))
	for _, item := range data.PendingExercises {
		if item.Exercise == nil {
			continue
		}
		card := &dto.ParentPendingExerciseCard{
			ClassroomExerciseID: item.Exercise.ClassroomExerciseId(),
			ClassroomID:         item.Exercise.ClassroomId(),
			Exercise:            dto.ExerciseToCard(item.Exercise),
		}
		if child, ok := data.ProfileByID[item.ChildProfileID]; ok {
			summary := dto.ProfileToSummary(child)
			s.signProfileAvatar(ctx, summary)
			card.Child = summary
		}
		if c, ok := data.ClassroomByID[item.Exercise.ClassroomId()]; ok {
			card.Classroom = dto.ClassroomToRef(c)
		}
		cards = append(cards, card)
	}
	return cards
}

// signProfileAvatar mutates the summary in place to add a short-lived
// presigned avatar URL. No-op when storage is disabled or no avatar_key.
func (s *Service) signProfileAvatar(ctx context.Context, summary *dto.ProfileSummary) {
	if summary == nil || s.storageProvider == nil || summary.AvatarKey == nil || *summary.AvatarKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *summary.AvatarKey,
		Expiration: presignTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("home.profile_avatar presign failed profile_id=%d err=%v", summary.ProfileID, err)
		return
	}
	summary.AvatarURL = &url
}

// signCoverURL mutates the classroom card in place to add a short-lived
// presigned cover URL. No-op when storage is disabled or no cover_key.
func (s *Service) signCoverURL(ctx context.Context, card *dto.ClassroomCard) {
	if card == nil || s.storageProvider == nil || card.CoverKey == nil || *card.CoverKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *card.CoverKey,
		Expiration: presignTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("home.classroom_cover presign failed classroom_id=%d err=%v", card.ClassroomID, err)
		return
	}
	card.CoverURL = &url
}

func isSupportedRole(role string) bool {
	switch enum.RoleType(role) {
	case enum.RoleTypeTeacher, enum.RoleTypeParent, enum.RoleTypeStudent:
		return true
	default:
		return false
	}
}
