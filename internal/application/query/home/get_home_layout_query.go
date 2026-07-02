package query

import (
	"context"

	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	exerciseDomain "math-ai.com/math-ai/internal/domain/exercise"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	quizDomain "math-ai.com/math-ai/internal/domain/quiz"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
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
	// homeQuizLimit caps the acting profile's standalone quiz history on
	// the home payload — most recent first, deep-link for older entries.
	homeQuizLimit = 20
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

	// Quizzes is the acting profile's standalone (out-of-classroom) quiz
	// history, most recent first — populated for every role.
	Quizzes []*quizDomain.Quiz

	Exercises   []*exerciseDomain.Exercise
	Submissions []*exerciseDomain.Submission

	// ExpiredExercises is the past-deadline counterpart of Exercises for
	// the teacher (assigned) and student (missed) layouts.
	ExpiredExercises []*exerciseDomain.Exercise

	// PendingExercises is the parent-only "children's to-do" feed: one
	// entry per (child, still-open exercise the child has not submitted).
	// ExpiredByChild is its past-deadline counterpart (missed exercises).
	PendingExercises []PendingExerciseForChild
	ExpiredByChild   []PendingExerciseForChild

	ExerciseByID  map[int64]*exerciseDomain.Exercise
	ClassroomByID map[int64]*classroomDomain.Classroom
	// ProfileByID is the layout's general profile cache keyed by
	// profile_id: classroom owners (teachers) for every role, plus the
	// parent's children. The module reads it to hydrate the teacher on each
	// classroom card and the child on each completion/pending row.
	ProfileByID map[int64]*profileDomain.Profile
}

// PendingExerciseForChild pairs a not-yet-completed exercise with the
// child it is pending for. The module maps each into a TaskCard (task_type
// pending or expired), hydrating the child / classroom refs from the
// HomeLayoutData lookup maps.
type PendingExerciseForChild struct {
	ChildProfileID int64
	Exercise       *exerciseDomain.Exercise
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
	quizRepo       quizDomain.IRepository
}

func NewGetHomeLayoutQueryHandler(
	classroomRepo classroomDomain.IRepository,
	memberRepo classroomDomain.IMemberRepository,
	exerciseRepo exerciseDomain.IRepository,
	submissionRepo exerciseDomain.ISubmissionRepository,
	profileRepo profileDomain.IRepository,
	quizRepo quizDomain.IRepository,
) *GetHomeLayoutQueryHandler {
	return &GetHomeLayoutQueryHandler{
		classroomRepo:  classroomRepo,
		memberRepo:     memberRepo,
		exerciseRepo:   exerciseRepo,
		submissionRepo: submissionRepo,
		profileRepo:    profileRepo,
		quizRepo:       quizRepo,
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

	// The acting profile's standalone quiz history is role-independent —
	// load it once before the role-specific assembly. The builders mutate
	// and return the same data pointer, so the quizzes ride through.
	if err := h.loadProfileQuizzes(ctx, p, data); err != nil {
		return nil, err
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

// loadProfileQuizzes fetches the acting profile's most-recent standalone
// quizzes (out-of-classroom history), mirroring the /quizzes/list filter
// scoped to a single profile_id.
func (h *GetHomeLayoutQueryHandler) loadProfileQuizzes(ctx context.Context, p *profileDomain.Profile, data *HomeLayoutData) error {
	profileID := p.ProfileId()
	quizzes, _, err := h.quizRepo.ListQuizzes(ctx, quizDomain.ListQuizzesFilter{
		ProfileID: &profileID,
	}, 1, homeQuizLimit)
	if err != nil {
		return err
	}
	data.Quizzes = quizzes
	return nil
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
	classroomIDsForExercises := activeClassroomIDs(data.Classrooms)
	exercises, err := h.exerciseRepo.ListByClassroomIds(ctx, exerciseDomain.ListByClassroomIdsParams{
		ClassroomIDs:     classroomIDsForExercises,
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

	expired, err := h.exerciseRepo.ListByClassroomIds(ctx, exerciseDomain.ListByClassroomIdsParams{
		ClassroomIDs:     classroomIDsForExercises,
		CallerProfileID:  p.ProfileId(),
		CreatorProfileID: &creator,
		OnlyActive:       true,
		ExpiredAsOf:      &now,
		Limit:            homeExerciseRenderCap,
	})
	if err != nil {
		return nil, err
	}
	data.ExpiredExercises = expired
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

	exerciseClassroomIDs := activeClassroomIDs(data.Classrooms)
	pending, err := h.listUnsubmittedExercises(ctx, p.ProfileId(), exerciseClassroomIDs, false)
	if err != nil {
		return nil, err
	}
	data.Exercises = pending

	expired, err := h.listUnsubmittedExercises(ctx, p.ProfileId(), exerciseClassroomIDs, true)
	if err != nil {
		return nil, err
	}
	data.ExpiredExercises = expired
	return data, nil
}

// listUnsubmittedExercises returns the still-open (expired=false) or
// past-deadline (expired=true) exercises in the given classrooms that the
// profile has NOT submitted, capped at homeExerciseRenderCap. Shared by
// the student layout's pending and expired buckets.
func (h *GetHomeLayoutQueryHandler) listUnsubmittedExercises(ctx context.Context, profileID int64, classroomIDs []int64, expired bool) ([]*exerciseDomain.Exercise, error) {
	now := mtime.Now()
	params := exerciseDomain.ListByClassroomIdsParams{
		ClassroomIDs:    classroomIDs,
		CallerProfileID: profileID,
		OnlyActive:      true,
		Limit:           homeExerciseFetchLimit,
	}
	if expired {
		params.ExpiredAsOf = &now
	} else {
		params.NotExpiredAsOf = &now
	}

	candidates, err := h.exerciseRepo.ListByClassroomIds(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	exerciseIDs := make([]int64, len(candidates))
	for i, e := range candidates {
		exerciseIDs[i] = e.ClassroomExerciseId()
	}
	submitted, err := h.submissionRepo.ListSubmittedExerciseIdsByProfile(ctx, profileID, exerciseIDs)
	if err != nil {
		return nil, err
	}
	out := make([]*exerciseDomain.Exercise, 0, len(candidates))
	for _, e := range candidates {
		if _, done := submitted[e.ClassroomExerciseId()]; done {
			continue
		}
		out = append(out, e)
		if len(out) >= homeExerciseRenderCap {
			break
		}
	}
	return out, nil
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

	if err := h.loadParentPendingExercises(ctx, childIDs, members, data); err != nil {
		return nil, err
	}
	return data, nil
}

// loadParentPendingExercises computes the parent's per-child "not
// completed yet" feeds, split into still-open (PendingExercises) and
// past-deadline (ExpiredByChild) buckets. Each is the set of PUBLIC
// exercises in the children's (non-archived) classrooms the child has not
// submitted; the same exercise yields one entry per child who still owes
// it.
func (h *GetHomeLayoutQueryHandler) loadParentPendingExercises(
	ctx context.Context,
	childIDs []int64,
	members []*classroomDomain.Member,
	data *HomeLayoutData,
) error {
	if len(data.Classrooms) == 0 {
		return nil
	}

	pending, err := h.collectParentChildExercises(ctx, childIDs, members, data, false)
	if err != nil {
		return err
	}
	data.PendingExercises = pending

	expired, err := h.collectParentChildExercises(ctx, childIDs, members, data, true)
	if err != nil {
		return err
	}
	data.ExpiredByChild = expired
	return nil
}

// collectParentChildExercises returns the per-child (child, exercise)
// pairs for either the still-open (expired=false) or past-deadline
// (expired=true) PUBLIC exercises the child has not submitted. Candidates
// are fetched once across every classroom; the submitted-set lookup is one
// read per child (children per parent is small).
func (h *GetHomeLayoutQueryHandler) collectParentChildExercises(
	ctx context.Context,
	childIDs []int64,
	members []*classroomDomain.Member,
	data *HomeLayoutData,
	expired bool,
) ([]PendingExerciseForChild, error) {
	// child -> the loaded classrooms they are an active member of.
	childClassrooms := make(map[int64][]int64, len(childIDs))
	for _, m := range members {
		if _, ok := data.ClassroomByID[m.ClassroomId()]; !ok {
			continue // archived / not loaded
		}
		childClassrooms[m.ProfileId()] = append(childClassrooms[m.ProfileId()], m.ClassroomId())
	}

	// CallerProfileID stays zero so only PUBLIC exercises surface, matching
	// the student-side view (children cannot see PRIVATE rows).
	now := mtime.Now()
	params := exerciseDomain.ListByClassroomIdsParams{
		ClassroomIDs: activeClassroomIDs(data.Classrooms),
		OnlyActive:   true,
		Limit:        homeExerciseFetchLimit,
	}
	if expired {
		params.ExpiredAsOf = &now
	} else {
		params.NotExpiredAsOf = &now
	}
	candidates, err := h.exerciseRepo.ListByClassroomIds(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	// Group candidates by classroom so each child only inspects the
	// exercises in their own classrooms.
	exByClassroom := make(map[int64][]*exerciseDomain.Exercise, len(data.Classrooms))
	for _, e := range candidates {
		exByClassroom[e.ClassroomId()] = append(exByClassroom[e.ClassroomId()], e)
	}

	out := make([]PendingExerciseForChild, 0)
	for _, childID := range childIDs {
		classroomIDs := childClassrooms[childID]
		if len(classroomIDs) == 0 {
			continue
		}
		childExercises := make([]*exerciseDomain.Exercise, 0)
		for _, cid := range classroomIDs {
			childExercises = append(childExercises, exByClassroom[cid]...)
		}
		if len(childExercises) == 0 {
			continue
		}
		exerciseIDs := make([]int64, len(childExercises))
		for i, e := range childExercises {
			exerciseIDs[i] = e.ClassroomExerciseId()
		}
		submitted, err := h.submissionRepo.ListSubmittedExerciseIdsByProfile(ctx, childID, exerciseIDs)
		if err != nil {
			return nil, err
		}
		for _, e := range childExercises {
			if _, done := submitted[e.ClassroomExerciseId()]; done {
				continue
			}
			out = append(out, PendingExerciseForChild{ChildProfileID: childID, Exercise: e})
			if len(out) >= homeExerciseRenderCap {
				break
			}
		}
		if len(out) >= homeExerciseRenderCap {
			break
		}
	}
	return out, nil
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
	ownerIDSet := make(map[int64]struct{}, len(classrooms))
	for _, c := range classrooms {
		if isArchived(c) {
			continue
		}
		data.Classrooms = append(data.Classrooms, c)
		data.ClassroomByID[c.ClassroomId()] = c
		ownerIDSet[c.OwnerProfileId()] = struct{}{}
	}
	// Batch-load the owner (teacher) profiles so each classroom card can
	// surface who runs the class without an N+1 per row.
	return h.hydrateProfiles(ctx, ownerIDSet, data)
}

// hydrateProfiles loads the profiles named by idSet into data.ProfileByID
// in one batched read, skipping ids already cached. Shared by the teacher
// hydration on every classroom card and reused for any other profile refs.
func (h *GetHomeLayoutQueryHandler) hydrateProfiles(ctx context.Context, idSet map[int64]struct{}, data *HomeLayoutData) error {
	ids := make([]int64, 0, len(idSet))
	for id := range idSet {
		if id == 0 {
			continue
		}
		if _, ok := data.ProfileByID[id]; ok {
			continue
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil
	}
	profiles, err := h.profileRepo.ListByProfileIds(ctx, ids)
	if err != nil {
		return err
	}
	for _, p := range profiles {
		data.ProfileByID[p.ProfileId()] = p
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
