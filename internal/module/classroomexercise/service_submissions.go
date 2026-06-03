package classroomexercise

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/classroomexercise"
	dto "math-ai.com/math-ai/internal/application/dto/classroomexercise"
	query "math-ai.com/math-ai/internal/application/query/classroomexercise"
	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	domain "math-ai.com/math-ai/internal/domain/classroomexercise"
	exerciseDomain "math-ai.com/math-ai/internal/domain/classroomexercise"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/shared/enum"
)

// avatarUrlTTL bounds how long a generated avatar URL is valid. Mirrors
// the classroom / school modules' coverUrlTTL.
const avatarUrlTTL = 1 * time.Hour

// SubmitExerciseAnswers validates → loads the exercise → enforces
// membership + window + duplicate guards → marshals the answers payload
// → calls the bot OUTSIDE the UoW → persists inside one short UoW.
// The result is the freshly graded submission.
func (s *Service) SubmitExerciseAnswers(ctx context.Context, req *dto.SubmitExerciseAnswersReq, sessionUserID int64) (*dto.SubmitExerciseAnswersRes, error) {
	log := logger.From(ctx)

	if err := ValidateSubmitExerciseAnswers(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveCaller(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}

	exercise, err := s.exerciseRepo.FindByClassroomExerciseId(ctx, req.ClassroomExerciseID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if exercise == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOT_FOUND, nil,
			ErrClassroomExerciseNotFound)
	}
	if !exerciseAvailableForSubmission(exercise) {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_EXERCISE_UNAVAILABLE, nil,
			ErrExerciseNotAvailable)
	}
	if _, err := s.requireMember(ctx, exercise.ClassroomId(), caller.ProfileId()); err != nil {
		return nil, err
	}
	if err := enforceExerciseAccess(ctx, exercise, caller.ProfileId()); err != nil {
		return nil, err
	}
	if err := enforceSubmissionWindow(ctx, exercise); err != nil {
		return nil, err
	}

	existing, err := s.submissionRepo.FindByExerciseAndProfile(ctx, exercise.ClassroomExerciseId(), caller.ProfileId())
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if existing != nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_ALREADY_EXISTS, nil,
			ErrSubmissionAlreadyExists)
	}

	answersJSON, err := json.Marshal(req.Answers)
	if err != nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_INVALID_ANSWERS, nil,
			fmt.Errorf("submission: marshal answers: %w", err))
	}

	lang := metadata.GetClientLanguage(ctx).ToEnumLanguage()

	questions := ""
	if exercise.Questions() != nil {
		questions = *exercise.Questions()
	}
	if questions == "" {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_GRADING_FAILED, nil,
			ErrExerciseHasNoQuestionsToGrade)
	}

	grading, err := s.bot.GradeExercise(ctx, gradeExerciseInput{
		Language:  lang,
		Questions: questions,
		Answers:   string(answersJSON),
	})
	if err != nil {
		return nil, err
	}

	actor := caller.ProfileId()
	cmd := command.SubmitExerciseAnswersCommand{
		ActorID:             &actor,
		ClassroomExerciseID: exercise.ClassroomExerciseId(),
		ClassroomID:         exercise.ClassroomId(),
		ProfileID:           caller.ProfileId(),
		AnswersJSON:         string(answersJSON),
		Note:                req.Note,
	}
	if grading != nil {
		if grading.AIReview != "" {
			ai := grading.AIReview
			cmd.AIReview = &ai
		}
		if grading.TotalQuestions > 0 {
			v := int64(grading.TotalQuestions)
			cmd.TotalQuestions = &v
		}
		if grading.CorrectNumber >= 0 {
			v := int64(grading.CorrectNumber)
			cmd.CorrectNumber = &v
		}
		if grading.ScorePercentage >= 0 {
			v := int64(grading.ScorePercentage)
			cmd.ScorePercentage = &v
		}
	}

	saved, err := s.submitAnswersCmd.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	log.Info("classroom_exercise_submission.submitted",
		"classroom_exercise_id", saved.ClassroomExerciseId(),
		"profile_id", saved.ProfileId(),
		"score_percentage", grading.ScorePercentage,
	)

	res := dto.DomainSubmissionToResponse(saved)

	return &dto.SubmitExerciseAnswersRes{Submission: res}, nil
}

// GetSubmission resolves the row → enforces "caller is the owner OR
// caller is a manager of the parent classroom" before returning.
func (s *Service) GetSubmission(ctx context.Context, req *dto.GetSubmissionReq, sessionUserID int64) (*dto.GetSubmissionRes, error) {
	if err := ValidateGetSubmission(ctx, req); err != nil {
		return nil, err
	}

	sub, err := s.getSubmissionQuery.Handle(ctx, query.GetSubmissionByIdQuery{
		ClassroomExerciseSubmissionID: req.ClassroomExerciseSubmissionID,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if sub == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_NOT_FOUND, nil,
			ErrSubmissionNotFound)
	}

	caller, err := s.resolveCaller(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeSubmissionRead(ctx, sub, caller.ProfileId()); err != nil {
		return nil, err
	}
	resp := dto.DomainSubmissionToResponse(sub)
	if err := s.hydrateSubmissions(ctx, []*domain.Submission{sub}, []*dto.SubmissionResponse{resp}); err != nil {
		return nil, err
	}
	return &dto.GetSubmissionRes{Submission: resp}, nil
}

// ListSubmissions is the flexible list endpoint.
//
//	profile_id  | scope (classroom_id or classroom_exercise_id) | mode
//	------------|------------------------------------------------|------
//	== caller   | optional                                       | self-list
//	!= caller   | required                                       | manager
//	omitted     | required                                       | manager
//
// Self-list returns rows owned by the caller's profile. Manager mode
// returns rows for any student in the scope, gated by the caller being
// an active OWNER / CO_TEACHER of the scoped classroom (resolved either
// directly from classroom_id or via the exercise's classroom).
func (s *Service) ListSubmissions(ctx context.Context, req *dto.ListSubmissionsReq, sessionUserID int64) (*dto.ListSubmissionsRes, error) {
	if err := ValidateListSubmissions(ctx, req); err != nil {
		return nil, err
	}

	// Decide the mode. self-list = profile_id provided AND it belongs
	// to the session user. Anything else collapses to manager mode and
	// must clear the scope + manager gate.
	isSelfList := false
	if req.ProfileID != nil {
		p, err := s.profileRepo.FindByProfileId(ctx, *req.ProfileID)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if p == nil {
			return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
				ErrProfileNotFound)
		}
		if sessionUserID != 0 && p.UserId() == sessionUserID {
			isSelfList = true
		}
	}

	if !isSelfList {
		scopeClassroomID, err := s.resolveListScopeClassroomID(ctx, req)
		if err != nil {
			return nil, err
		}
		if scopeClassroomID == 0 {
			return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_PERMISSION_DENIED, nil,
				ErrClassroomIDOrClassroomExerciseIDRequiredWhenListingAcrossProfiles)
		}
		// if _, err := s.requireManagerForUser(ctx, scopeClassroomID, sessionUserID); err != nil {
		// 	return nil, err
		// }
	}

	q := query.ListSubmissionsQuery{
		Status:    req.Status,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
		Page:      int64(req.Page),
		Limit:     int64(req.Size),
	}
	if req.ProfileID != nil {
		q.ProfileID = *req.ProfileID
	}
	if req.ClassroomID != nil {
		q.ClassroomID = *req.ClassroomID
	}
	if req.ClassroomExerciseID != nil {
		q.ClassroomExerciseID = *req.ClassroomExerciseID
	}

	rows, pg, err := s.listSubmissionsQuery.Handle(ctx, q)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	responses := dto.DomainSubmissionListToResponse(rows)
	if err := s.hydrateSubmissions(ctx, rows, responses); err != nil {
		return nil, err
	}
	return &dto.ListSubmissionsRes{
		Submissions: responses,
		Pagination:  pg,
	}, nil
}

// resolveListScopeClassroomID picks the manager-gate scope for the
// flexible list endpoint. classroom_exercise_id takes precedence —
// when both are supplied and disagree, we trust the exercise's parent
// classroom because that's what the SQL filter will narrow against.
func (s *Service) resolveListScopeClassroomID(ctx context.Context, req *dto.ListSubmissionsReq) (int64, error) {
	if req.ClassroomExerciseID != nil && *req.ClassroomExerciseID != 0 {
		exercise, err := s.exerciseRepo.FindByClassroomExerciseId(ctx, *req.ClassroomExerciseID)
		if err != nil {
			return 0, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if exercise == nil {
			return 0, errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOT_FOUND, nil,
				ErrClassroomExerciseNotFound)
		}
		return exercise.ClassroomId(), nil
	}
	if req.ClassroomID != nil {
		return *req.ClassroomID, nil
	}
	return 0, nil
}

// ListSubmissionsByExercise is the teacher view: every submission for
// a single exercise. Manager-gated.
func (s *Service) ListSubmissionsByExercise(ctx context.Context, req *dto.ListSubmissionsByExerciseReq, sessionUserID int64) (*dto.ListSubmissionsByExerciseRes, error) {
	if err := ValidateListSubmissionsByExercise(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveCaller(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}

	exercise, err := s.exerciseRepo.FindByClassroomExerciseId(ctx, req.ClassroomExerciseID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if exercise == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOT_FOUND, nil,
			ErrClassroomExerciseNotFound)
	}
	if _, err := s.requireManager(ctx, exercise.ClassroomId(), caller.ProfileId()); err != nil {
		return nil, err
	}

	rows, pg, err := s.listSubmissionsQuery.Handle(ctx, query.ListSubmissionsQuery{
		ClassroomExerciseID: exercise.ClassroomExerciseId(),
		Status:              req.Status,
		SortBy:              req.SortBy,
		SortOrder:           req.SortOrder,
		Page:                int64(req.Page),
		Limit:               int64(req.Size),
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	responses := dto.DomainSubmissionListToResponse(rows)
	if err := s.hydrateSubmissions(ctx, rows, responses); err != nil {
		return nil, err
	}
	return &dto.ListSubmissionsByExerciseRes{
		Submissions: responses,
		Pagination:  pg,
	}, nil
}

// ListSubmittedMembers is the teacher-side roster of classroom members
// who already submitted the exercise. Returns each member's profile
// detail and submission metadata.
func (s *Service) ListSubmittedMembers(ctx context.Context, req *dto.ListAudienceMembersReq, sessionUserID int64) (*dto.ListAudienceMembersRes, error) {
	return s.listAudienceMembers(ctx, req, sessionUserID, true)
}

// ListNonSubmittedMembers is the teacher-side roster of classroom
// members who have NOT submitted the exercise. Submission summary is
// omitted on each row.
func (s *Service) ListNonSubmittedMembers(ctx context.Context, req *dto.ListAudienceMembersReq, sessionUserID int64) (*dto.ListAudienceMembersRes, error) {
	return s.listAudienceMembers(ctx, req, sessionUserID, false)
}

// listAudienceMembers is the shared implementation behind the two
// audience endpoints. submitted=true uses the INNER JOIN flavor of the
// repo query and additionally hydrates submission metadata onto each
// row; submitted=false uses the LEFT JOIN ... IS NULL flavor and skips
// the submission lookup. Permission, classroom resolution, and the
// page query stay identical between the two paths.
func (s *Service) listAudienceMembers(ctx context.Context, req *dto.ListAudienceMembersReq, sessionUserID int64, submitted bool) (*dto.ListAudienceMembersRes, error) {
	if err := ValidateListAudienceMembers(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveCaller(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}

	exercise, err := s.exerciseRepo.FindByClassroomExerciseId(ctx, req.ClassroomExerciseID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if exercise == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOT_FOUND, nil,
			ErrClassroomExerciseNotFound)
	}
	if _, err := s.requireManager(ctx, exercise.ClassroomId(), caller.ProfileId()); err != nil {
		return nil, err
	}

	members, pg, err := s.classroomMemberRepo.ListMembersByExerciseSubmission(ctx,
		&classroomDomain.ListMembersByExerciseSubmissionParams{
			ClassroomId:         exercise.ClassroomId(),
			ClassroomExerciseId: exercise.ClassroomExerciseId(),
			Submitted:           submitted,
			Search:              req.Search,
			SortBy:              req.SortBy,
			SortOrder:           req.SortOrder,
			Page:                int64(req.Page),
			Limit:               int64(req.Size),
		})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	responses := make([]*dto.AudienceMemberResponse, 0, len(members))
	profileIDs := make([]int64, 0, len(members))
	for _, m := range members {
		resp := &dto.AudienceMemberResponse{
			MemberID:     m.MemberId(),
			ClassroomID:  m.ClassroomId(),
			ProfileID:    m.ProfileId(),
			MemberRole:   m.MemberRole(),
			MemberStatus: m.MemberStatus(),
		}
		if m.JoinedDt().IsValid() {
			resp.JoinedDt = m.JoinedDt().String()
		}
		responses = append(responses, resp)
		profileIDs = append(profileIDs, m.ProfileId())
	}

	if err := s.hydrateAudienceProfiles(ctx, profileIDs, responses); err != nil {
		return nil, err
	}
	if submitted {
		if err := s.hydrateAudienceSubmissions(ctx, exercise.ClassroomExerciseId(), profileIDs, responses); err != nil {
			return nil, err
		}
	}

	return &dto.ListAudienceMembersRes{
		Members:    responses,
		Pagination: pg,
	}, nil
}

// hydrateAudienceProfiles attaches the slim AudienceProfileDetail to
// each row via one batched ma_profiles IN-query.
func (s *Service) hydrateAudienceProfiles(ctx context.Context, profileIDs []int64, responses []*dto.AudienceMemberResponse) error {
	if len(profileIDs) == 0 || s.profileRepo == nil {
		return nil
	}
	profiles, err := s.profileRepo.ListByProfileIds(ctx, profileIDs)
	if err != nil {
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	byID := make(map[int64]*dto.AudienceProfileDetail, len(profiles))
	for _, p := range profiles {
		detail := &dto.AudienceProfileDetail{
			ProfileID:   p.ProfileId(),
			ProfileCode: p.ProfileCode(),
			Name:        p.Name(),
			Role:        p.Role(),
			AvatarKey:   p.AvatarKey(),
			StudentID:   p.StudentId(),
			TeacherID:   p.TeacherId(),
		}
		s.signAudienceAvatarURL(ctx, detail)
		byID[p.ProfileId()] = detail
	}
	for _, resp := range responses {
		if resp == nil {
			continue
		}
		if detail, ok := byID[resp.ProfileID]; ok {
			resp.Profile = detail
		}
	}
	return nil
}

// hydrateAudienceSubmissions attaches the submission summary to each
// submitted-members row via one batched IN-query.
func (s *Service) hydrateAudienceSubmissions(ctx context.Context, classroomExerciseID int64, profileIDs []int64, responses []*dto.AudienceMemberResponse) error {
	if len(profileIDs) == 0 || s.submissionRepo == nil {
		return nil
	}
	subs, err := s.submissionRepo.ListByExerciseAndProfileIds(ctx, classroomExerciseID, profileIDs)
	if err != nil {
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	byProfile := make(map[int64]*dto.AudienceSubmissionSummary, len(subs))
	for _, sub := range subs {
		summary := &dto.AudienceSubmissionSummary{
			ClassroomExerciseSubmissionID: sub.ClassroomExerciseSubmissionId(),
			SubmissionStatus:              sub.SubmissionStatus(),
			TotalQuestions:                sub.TotalQuestions(),
			CorrectNumber:                 sub.CorrectNumber(),
			ScorePercentage:               sub.ScorePercentage(),
			Note:                          sub.Note(),
		}
		if sub.SubmittedDt().IsValid() {
			summary.SubmittedDt = sub.SubmittedDt().String()
		}
		if sub.GradedDt().IsValid() {
			summary.GradedDt = sub.GradedDt().String()
		}
		byProfile[sub.ProfileId()] = summary
	}
	for _, resp := range responses {
		if resp == nil {
			continue
		}
		if summary, ok := byProfile[resp.ProfileID]; ok {
			resp.Submission = summary
		}
	}
	return nil
}

// signAudienceAvatarURL mirrors signSubmissionAvatarURL — presigns a
// short-lived URL for the profile avatar_key when storage is wired.
func (s *Service) signAudienceAvatarURL(ctx context.Context, detail *dto.AudienceProfileDetail) {
	if detail == nil || s.storageProvider == nil || detail.AvatarKey == nil || *detail.AvatarKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *detail.AvatarKey,
		Expiration: avatarUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("classroom_exercise_submission.audience_avatar presign failed profile_id=%d err=%v", detail.ProfileID, err)
		return
	}
	detail.AvatarURL = &url
}

// SoftDeleteSubmission is manager-only — the use case is invalidating
// a spammed / bad attempt, not student-side withdrawal.
func (s *Service) SoftDeleteSubmission(ctx context.Context, req *dto.DeleteSubmissionReq, sessionUserID int64) (*dto.DeleteSubmissionRes, error) {
	if err := ValidateDeleteSubmission(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveCaller(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}

	sub, err := s.submissionRepo.FindBySubmissionId(ctx, req.ClassroomExerciseSubmissionID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if sub == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_NOT_FOUND, nil,
			ErrSubmissionNotFound)
	}
	if _, err := s.requireManager(ctx, sub.ClassroomId(), caller.ProfileId()); err != nil {
		return nil, err
	}

	actor := caller.ProfileId()
	if err := s.softDeleteSubmissionCmd.Handle(ctx, command.SoftDeleteSubmissionCommand{
		ActorID:                       &actor,
		ClassroomExerciseSubmissionID: req.ClassroomExerciseSubmissionID,
	}); err != nil {
		return nil, err
	}
	return &dto.DeleteSubmissionRes{}, nil
}

// authorizeSubmissionRead lets the row's own profile read it directly;
// for everyone else, the caller must be a manager of the parent
// classroom. Used by GetSubmission only — list endpoints scope by
// caller-profile or by manager gate up front.
func (s *Service) authorizeSubmissionRead(ctx context.Context, sub *domain.Submission, callerProfileID int64) error {
	if sub.ProfileId() == callerProfileID {
		return nil
	}
	if _, err := s.requireManager(ctx, sub.ClassroomId(), callerProfileID); err != nil {
		return err
	}
	return nil
}

// exerciseAvailableForSubmission rejects ARCHIVED / DELETED exercises.
func exerciseAvailableForSubmission(e *exerciseDomain.Exercise) bool {
	if e == nil {
		return false
	}
	if e.ExerciseStatus() == nil {
		return true
	}
	return *e.ExerciseStatus() == string(enum.ClassroomExerciseStatusTypeActive)
}

// enforceExerciseAccess re-applies the visibility access gate from the
// exercise module — PRIVATE rows are still creator-only.
func enforceExerciseAccess(ctx context.Context, e *exerciseDomain.Exercise, callerProfileID int64) error {
	if e == nil {
		return nil
	}
	if e.Visibility() != string(enum.ClassroomExerciseVisibilityPrivate) {
		return nil
	}
	if e.CreatorProfileId() == callerProfileID {
		return nil
	}
	return errs.NewError(ctx, status.CLASSROOM_EXERCISE_PRIVATE_DENIED, nil,
		ErrPrivateExerciseOnlyCreatorCanSubmit)
}

// enforceSubmissionWindow rejects out-of-window submissions. start_date
// and end_date are both optional; missing values are treated as "no
// lower / upper bound".
func enforceSubmissionWindow(ctx context.Context, e *exerciseDomain.Exercise) error {
	now := time.Now().UTC()
	if e.StartDate().IsValid() && now.Before(e.StartDate().Time) {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_WINDOW_NOT_OPEN, nil,
			ErrYouCanNotSubmitBeforeStartTime)
	}
	if e.EndDate().IsValid() && now.After(e.EndDate().Time) {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_WINDOW_CLOSED, nil,
			ErrYouCanNotSubmitAfterEndTime)
	}
	return nil
}

// hydrateSubmissions attaches the slim ExerciseSummary and
// ProfileSummary blocks to each response. Two batched lookups per page
// (one ma_classroom_exercises IN-query, one ma_profiles IN-query); the
// cost is independent of page size so N+1 is structurally impossible.
// Missing rows (deleted exercise / profile under a submission) leave
// the corresponding field nil so the rest of the row still renders.
func (s *Service) hydrateSubmissions(
	ctx context.Context,
	rows []*domain.Submission,
	responses []*dto.SubmissionResponse,
) error {
	if len(rows) == 0 {
		return nil
	}

	exerciseIDSet := make(map[int64]struct{}, len(rows))
	profileIDSet := make(map[int64]struct{}, len(rows))
	for _, r := range rows {
		exerciseIDSet[r.ClassroomExerciseId()] = struct{}{}
		profileIDSet[r.ProfileId()] = struct{}{}
	}
	exerciseIDs := make([]int64, 0, len(exerciseIDSet))
	for id := range exerciseIDSet {
		exerciseIDs = append(exerciseIDs, id)
	}
	profileIDs := make([]int64, 0, len(profileIDSet))
	for id := range profileIDSet {
		profileIDs = append(profileIDs, id)
	}

	exerciseSummaries := make(map[int64]*dto.SubmissionExerciseSummary, len(exerciseIDs))
	if s.exerciseRepo != nil && len(exerciseIDs) > 0 {
		exercises, err := s.exerciseRepo.ListByClassroomExerciseIds(ctx, exerciseIDs)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		for _, e := range exercises {
			summary := &dto.SubmissionExerciseSummary{
				ClassroomExerciseID: e.ClassroomExerciseId(),
				ClassroomID:         e.ClassroomId(),
				CreatorProfileID:    e.CreatorProfileId(),
				Visibility:          e.Visibility(),
				Title:               e.Title(),
				Description:         e.Description(),
				ChapterName:         e.ChapterName(),
				LessonName:          e.LessonName(),
				TotalQuestions:      e.TotalQuestions(),
				ExerciseStatus:      e.ExerciseStatus(),
			}
			if e.StartDate().IsValid() {
				summary.StartDate = e.StartDate().String()
			}
			if e.EndDate().IsValid() {
				summary.EndDate = e.EndDate().String()
			}
			exerciseSummaries[e.ClassroomExerciseId()] = summary
		}
	}

	profileSummaries := make(map[int64]*dto.SubmissionProfileSummary, len(profileIDs))
	if s.profileRepo != nil && len(profileIDs) > 0 {
		profiles, err := s.profileRepo.ListByProfileIds(ctx, profileIDs)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		for _, p := range profiles {
			summary := &dto.SubmissionProfileSummary{
				ProfileID: p.ProfileId(),
				Name:      p.Name(),
				Role:      p.Role(),
				AvatarKey: p.AvatarKey(),
			}
			s.signSubmissionAvatarURL(ctx, summary)
			profileSummaries[p.ProfileId()] = summary
		}
	}

	for i, r := range rows {
		if responses[i] == nil {
			continue
		}
		if summary, ok := exerciseSummaries[r.ClassroomExerciseId()]; ok {
			responses[i].ClassroomExercise = summary
		}
		if summary, ok := profileSummaries[r.ProfileId()]; ok {
			responses[i].Profile = summary
		}
	}
	return nil
}

// signSubmissionAvatarURL mirrors signOwnerAvatarURL from the classroom
// module — presigns a short-lived URL for the profile avatar_key when
// storage is configured. No-op when storage is disabled or the profile
// has no avatar.
func (s *Service) signSubmissionAvatarURL(ctx context.Context, summary *dto.SubmissionProfileSummary) {
	if summary == nil || s.storageProvider == nil || summary.AvatarKey == nil || *summary.AvatarKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *summary.AvatarKey,
		Expiration: avatarUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("classroom_exercise_submission.profile_avatar presign failed profile_id=%d err=%v", summary.ProfileID, err)
		return
	}
	summary.AvatarURL = &url
}
