package query

import (
	"context"

	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	exerciseDomain "math-ai.com/math-ai/internal/domain/exercise"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/enum"
)

// homeExerciseLimit / homeCompletionLimit cap the size of the per-role
// feeds so the home payload stays small regardless of how prolific a
// classroom is. The student pending list is fetched slightly larger than
// it is rendered because the already-submitted rows are subtracted in
// application code after the DB read.
const (
	homeExerciseFetchLimit = 200
	homeExerciseRenderCap  = 50
	homeCompletionLimit    = 20
)

// HomeLayoutData is the domain-level assembly the module maps into the
// role-discriminated DTO. The acting profile and (for parents) the child
// profiles are carried alongside the classrooms/exercises/submissions so
// the module can render slim cards and sign avatar/cover URLs without
// re-querying. The *ByID maps exist purely to hydrate embedded refs on
// the parent completion feed in O(1).
type HomeLayoutData struct {
	Role     string
	Children []*profileDomain.Profile

	Classrooms               []*classroomDomain.Classroom
	RoleByClassroom          map[int64]string
	MemberProfileByClassroom map[int64]int64

	Exercises   []*exerciseDomain.Exercise
	Submissions []*exerciseDomain.Submission

	ExerciseByID  map[int64]*exerciseDomain.Exercise
	ClassroomByID map[int64]*classroomDomain.Classroom
	ProfileByID   map[int64]*profileDomain.Profile
}

// GetHomeLayoutQuery carries the already-resolved acting profile. The
// module layer is responsible for loading it and confirming it belongs to
// the session user before handing it here.
type GetHomeLayoutQuery struct {
	Profile *profileDomain.Profile
}

// GetHomeLayoutQueryHandler composes the cross-aggregate read for the
// home dashboard. It only reads — no UoW — and every fan-out is batched
// so a home render is a small, fixed number of round trips regardless of
// how many classrooms / exercises are in play.
type GetHomeLayoutQueryHandler struct {
	classroomRepo  classroomDomain.IRepository
	memberRepo     classroomDomain.IMemberRepository
	exerciseRepo   exerciseDomain.IRepository
	submissionRepo exerciseDomain.ISubmissionRepository
	profileRepo    profileDomain.IRepository
}

func NewGetHomeLayoutQueryHandler(
	classroomRepo classroomDomain.IRepository,
	memberRepo classroomDomain.IMemberRepository,
	exerciseRepo exerciseDomain.IRepository,
	submissionRepo exerciseDomain.ISubmissionRepository,
	profileRepo profileDomain.IRepository,
) *GetHomeLayoutQueryHandler {
	return &GetHomeLayoutQueryHandler{
		classroomRepo:  classroomRepo,
		memberRepo:     memberRepo,
		exerciseRepo:   exerciseRepo,
		submissionRepo: submissionRepo,
		profileRepo:    profileRepo,
	}
}

func (h *GetHomeLayoutQueryHandler) Handle(ctx context.Context, q GetHomeLayoutQuery) (*HomeLayoutData, error) {
	p := q.Profile
	data := &HomeLayoutData{
		Role:                     p.Role(),
		RoleByClassroom:          map[int64]string{},
		MemberProfileByClassroom: map[int64]int64{},
		ExerciseByID:             map[int64]*exerciseDomain.Exercise{},
		ClassroomByID:            map[int64]*classroomDomain.Classroom{},
		ProfileByID:              map[int64]*profileDomain.Profile{},
	}

	switch enum.RoleType(p.Role()) {
	case enum.RoleTypeTeacher:
		return h.buildTeacher(ctx, p, data)
	case enum.RoleTypeParent:
		return h.buildParent(ctx, p, data)
	case enum.RoleTypeStudent:
		return h.buildStudent(ctx, p, data)
	default:
		// Unknown role — the module guards against this, but return an
		// empty (role-only) payload defensively instead of erroring.
		return data, nil
	}
}

// buildTeacher: classrooms the teacher manages (OWNER / CO_TEACHER) plus
// the still-open exercises they personally assigned in those classrooms.
func (h *GetHomeLayoutQueryHandler) buildTeacher(ctx context.Context, p *profileDomain.Profile, data *HomeLayoutData) (*HomeLayoutData, error) {
	members, err := h.memberRepo.ListActiveByProfileIds(ctx, []int64{p.ProfileId()})
	if err != nil {
		return nil, err
	}
	classroomIDs := make([]int64, 0, len(members))
	for _, m := range members {
		if !isManagerRole(m.MemberRole()) {
			continue
		}
		if _, seen := data.RoleByClassroom[m.ClassroomId()]; seen {
			continue
		}
		data.RoleByClassroom[m.ClassroomId()] = m.MemberRole()
		classroomIDs = append(classroomIDs, m.ClassroomId())
	}

	if err := h.loadClassrooms(ctx, classroomIDs, data); err != nil {
		return nil, err
	}
	if len(data.Classrooms) == 0 {
		return data, nil
	}

	now := mtime.Now()
	creator := p.ProfileId()
	exercises, err := h.exerciseRepo.ListByClassroomIds(ctx, exerciseDomain.ListByClassroomIdsParams{
		ClassroomIDs:     activeClassroomIDs(data.Classrooms),
		CallerProfileID:  p.ProfileId(),
		CreatorProfileID: &creator,
		OnlyActive:       true,
		NotExpiredAsOf:   &now,
		Limit:            homeExerciseRenderCap,
	})
	if err != nil {
		return nil, err
	}
	data.Exercises = exercises
	return data, nil
}

// buildStudent: classrooms the student has joined plus the still-open
// exercises in those classrooms they have not yet submitted.
func (h *GetHomeLayoutQueryHandler) buildStudent(ctx context.Context, p *profileDomain.Profile, data *HomeLayoutData) (*HomeLayoutData, error) {
	members, err := h.memberRepo.ListActiveByProfileIds(ctx, []int64{p.ProfileId()})
	if err != nil {
		return nil, err
	}
	classroomIDs := make([]int64, 0, len(members))
	for _, m := range members {
		if _, seen := data.RoleByClassroom[m.ClassroomId()]; seen {
			continue
		}
		data.RoleByClassroom[m.ClassroomId()] = m.MemberRole()
		classroomIDs = append(classroomIDs, m.ClassroomId())
	}

	if err := h.loadClassrooms(ctx, classroomIDs, data); err != nil {
		return nil, err
	}
	if len(data.Classrooms) == 0 {
		return data, nil
	}

	now := mtime.Now()
	candidates, err := h.exerciseRepo.ListByClassroomIds(ctx, exerciseDomain.ListByClassroomIdsParams{
		ClassroomIDs:    activeClassroomIDs(data.Classrooms),
		CallerProfileID: p.ProfileId(),
		OnlyActive:      true,
		NotExpiredAsOf:  &now,
		Limit:           homeExerciseFetchLimit,
	})
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return data, nil
	}

	exerciseIDs := make([]int64, len(candidates))
	for i, e := range candidates {
		exerciseIDs[i] = e.ClassroomExerciseId()
	}
	submitted, err := h.submissionRepo.ListSubmittedExerciseIdsByProfile(ctx, p.ProfileId(), exerciseIDs)
	if err != nil {
		return nil, err
	}
	pending := make([]*exerciseDomain.Exercise, 0, len(candidates))
	for _, e := range candidates {
		if _, done := submitted[e.ClassroomExerciseId()]; done {
			continue
		}
		pending = append(pending, e)
		if len(pending) >= homeExerciseRenderCap {
			break
		}
	}
	data.Exercises = pending
	return data, nil
}

// buildParent: every classroom the parent's children are enrolled in plus
// a recent feed of exercises those children just completed.
func (h *GetHomeLayoutQueryHandler) buildParent(ctx context.Context, p *profileDomain.Profile, data *HomeLayoutData) (*HomeLayoutData, error) {
	siblings, err := h.profileRepo.ListByUserId(ctx, p.UserId())
	if err != nil {
		return nil, err
	}
	childIDs := make([]int64, 0, len(siblings))
	for _, c := range siblings {
		// Children are the STUDENT profiles under the same user — the
		// acting parent profile itself is excluded.
		if c.ProfileId() == p.ProfileId() {
			continue
		}
		if c.Role() != string(enum.RoleTypeStudent) {
			continue
		}
		data.Children = append(data.Children, c)
		data.ProfileByID[c.ProfileId()] = c
		childIDs = append(childIDs, c.ProfileId())
	}
	if len(childIDs) == 0 {
		return data, nil
	}

	members, err := h.memberRepo.ListActiveByProfileIds(ctx, childIDs)
	if err != nil {
		return nil, err
	}
	classroomIDs := make([]int64, 0, len(members))
	for _, m := range members {
		if _, seen := data.RoleByClassroom[m.ClassroomId()]; seen {
			continue
		}
		data.RoleByClassroom[m.ClassroomId()] = m.MemberRole()
		data.MemberProfileByClassroom[m.ClassroomId()] = m.ProfileId()
		classroomIDs = append(classroomIDs, m.ClassroomId())
	}
	if err := h.loadClassrooms(ctx, classroomIDs, data); err != nil {
		return nil, err
	}

	// Recent completions across all children, with the referenced
	// exercises hydrated in one batched read.
	submissions, err := h.submissionRepo.ListRecentByProfileIds(ctx, childIDs, homeCompletionLimit)
	if err != nil {
		return nil, err
	}
	data.Submissions = submissions
	if len(submissions) > 0 {
		exerciseIDSet := make(map[int64]struct{}, len(submissions))
		exerciseIDs := make([]int64, 0, len(submissions))
		for _, s := range submissions {
			if _, ok := exerciseIDSet[s.ClassroomExerciseId()]; ok {
				continue
			}
			exerciseIDSet[s.ClassroomExerciseId()] = struct{}{}
			exerciseIDs = append(exerciseIDs, s.ClassroomExerciseId())
		}
		exercises, err := h.exerciseRepo.ListByClassroomExerciseIds(ctx, exerciseIDs)
		if err != nil {
			return nil, err
		}
		for _, e := range exercises {
			data.ExerciseByID[e.ClassroomExerciseId()] = e
		}
	}
	return data, nil
}

// loadClassrooms fetches the classroom rows for the given ids (skipping
// ARCHIVED — the home dashboard surfaces live classrooms only) and
// records them on data both as a slice and in the ClassroomByID map.
func (h *GetHomeLayoutQueryHandler) loadClassrooms(ctx context.Context, ids []int64, data *HomeLayoutData) error {
	if len(ids) == 0 {
		return nil
	}
	classrooms, err := h.classroomRepo.ListClassroomsByIds(ctx, ids)
	if err != nil {
		return err
	}
	for _, c := range classrooms {
		if isArchived(c) {
			continue
		}
		data.Classrooms = append(data.Classrooms, c)
		data.ClassroomByID[c.ClassroomId()] = c
	}
	return nil
}

func isManagerRole(role string) bool {
	return role == string(enum.ClassroomMemberRoleTypeOwner) ||
		role == string(enum.ClassroomMemberRoleTypeCoTeacher)
}

func isArchived(c *classroomDomain.Classroom) bool {
	return c.ClassroomStatus() != nil &&
		*c.ClassroomStatus() == string(enum.ClassroomStatusTypeArchived)
}

// activeClassroomIDs returns the ids of the (already ARCHIVED-filtered)
// classrooms so the exercise read targets only the live set.
func activeClassroomIDs(classrooms []*classroomDomain.Classroom) []int64 {
	ids := make([]int64, len(classrooms))
	for i, c := range classrooms {
		ids[i] = c.ClassroomId()
	}
	return ids
}
