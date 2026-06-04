package exercise

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/exercise"
	domain "math-ai.com/math-ai/internal/domain/exercise"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/shared/enum"
)

// resolveCaller resolves ProfileID with a fallback: when the request
// omits it, the session's authenticated user must have exactly one
// profile (otherwise the caller has to disambiguate). Mirrors how the
// classroom module derives the acting profile.
func (s *Service) resolveCaller(ctx context.Context, profileID *int64, sessionUserID int64) (*profileDomain.Profile, error) {
	if profileID != nil && *profileID != 0 {
		return s.resolveActingProfile(ctx, *profileID, sessionUserID)
	}
	// No ProfileID supplied — require the session and look up by user.
	if sessionUserID == 0 {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_PERMISSION_DENIED, nil, ErrProfileIDRequired)
	}
	profiles, err := s.profileRepo.ListByUserId(ctx, sessionUserID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if len(profiles) == 1 {
		return profiles[0], nil
	}
	return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_PERMISSION_DENIED, nil, ErrProfileIDRequiredMultiProfile)
}

// resolveExerciseProgram verifies the supplied program belongs to the
// classroom. When the teacher omits ProgramID we leave it nil — the
// exercise will be persisted without program context and the prompt
// renders a curriculum-agnostic brief.
func (s *Service) resolveExerciseProgram(ctx context.Context, classroomID int64, programID *int64) (*int64, error) {
	if programID == nil || *programID == 0 {
		return nil, nil
	}
	ids, err := s.classroomProgramRepo.ListProgramIdsByClassroomId(ctx, classroomID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	for _, id := range ids {
		if id == *programID {
			return programID, nil
		}
	}
	return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_PROGRAM_NOT_IN_CLASSROOM, nil, ErrProgramNotAssociated)
}

// hydrateSubmissionStatus stamps the per-exercise SubmissionStatus
// field on each response. One IN-query against
// ma_exercise_submissions per page regardless of size —
// projecting only the exercise id so the index ix_exercise_profile_submitted
// covers the lookup. When profileID is nil/zero or the submission repo
// is unwired, every row stays at the DTO default ("NONE") and no
// round trip is made.
func (s *Service) hydrateSubmissionStatus(
	ctx context.Context,
	profileID *int64,
	exercises []*domain.Exercise,
	responses []*dto.ExerciseResponse,
) error {
	if len(exercises) == 0 || s.submissionRepo == nil {
		return nil
	}
	if profileID == nil || *profileID == 0 {
		return nil
	}
	idSet := make(map[int64]struct{}, len(exercises))
	for _, e := range exercises {
		idSet[e.ClassroomExerciseId()] = struct{}{}
	}
	ids := make([]int64, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	submitted, err := s.submissionRepo.ListSubmittedExerciseIdsByProfile(ctx, *profileID, ids)
	if err != nil {
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	for i, e := range exercises {
		if responses[i] == nil {
			continue
		}
		if _, ok := submitted[e.ClassroomExerciseId()]; ok {
			responses[i].SubmissionStatus = dto.ExerciseSubmissionStatusSubmitted
		}
	}
	return nil
}

// hydrateClassroomAndProgram attaches the slim Classroom and Program
// blocks to each response. Two batched IN-lookups per page (one against
// ma_classrooms, one against ma_programs); cost is independent of page
// size so N+1 is structurally impossible. Missing rows (deleted
// classroom / program under an exercise) leave the corresponding field
// nil so the rest of the row still renders.
func (s *Service) hydrateClassroomAndProgram(
	ctx context.Context,
	exercises []*domain.Exercise,
	responses []*dto.ExerciseResponse,
) error {
	if len(exercises) == 0 {
		return nil
	}

	classroomIDSet := make(map[int64]struct{}, len(exercises))
	programIDSet := make(map[int64]struct{}, len(exercises))
	for _, e := range exercises {
		classroomIDSet[e.ClassroomId()] = struct{}{}
		if pid := e.ProgramId(); pid != nil && *pid != 0 {
			programIDSet[*pid] = struct{}{}
		}
	}

	classroomSummaries := make(map[int64]*dto.ExerciseClassroomSummary, len(classroomIDSet))
	if s.classroomRepo != nil && len(classroomIDSet) > 0 {
		ids := make([]int64, 0, len(classroomIDSet))
		for id := range classroomIDSet {
			ids = append(ids, id)
		}
		rows, err := s.classroomRepo.ListClassroomsByIds(ctx, ids)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		for _, c := range rows {
			classroomSummaries[c.ClassroomId()] = &dto.ExerciseClassroomSummary{
				ClassroomID:    c.ClassroomId(),
				Name:           c.Name(),
				Description:    c.Description(),
				ClassroomCode:  c.ClassroomCode(),
				OwnerProfileID: c.OwnerProfileId(),
				SchoolID:       c.SchoolId(),
				GradeID:        c.GradeId(),
			}
		}
	}

	programSummaries := make(map[int64]*dto.ExerciseProgramSummary, len(programIDSet))
	if s.programRepo != nil && len(programIDSet) > 0 {
		ids := make([]int64, 0, len(programIDSet))
		for id := range programIDSet {
			ids = append(ids, id)
		}
		lang := metadata.GetClientLanguage(ctx).ToEnumLanguage()
		rows, err := s.programRepo.ListProgramsByIds(ctx, ids, lang)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		for _, p := range rows {
			programSummaries[p.ProgramId()] = &dto.ExerciseProgramSummary{
				ProgramID:   p.ProgramId(),
				Label:       p.Label(),
				Description: p.Description(),
			}
		}
	}

	for i, e := range exercises {
		if responses[i] == nil {
			continue
		}
		if summary, ok := classroomSummaries[e.ClassroomId()]; ok {
			responses[i].Classroom = summary
		}
		if pid := e.ProgramId(); pid != nil && *pid != 0 {
			if summary, ok := programSummaries[*pid]; ok {
				responses[i].Program = summary
			}
		}
	}
	return nil
}

// resolveCurriculumLabels best-effort hydrates grade + program labels
// for the bot prompt. Errors here are logged but do not fail the
// request — the prompt renders only the lines whose label is non-empty.
func (s *Service) resolveCurriculumLabels(ctx context.Context, gradeID, programID *int64, lang enum.LanguageType) (string, string) {
	log := logger.From(ctx)
	var gradeLabel, programLabel string

	if gradeID != nil && *gradeID != 0 {
		grades, err := s.gradeRepo.ListGradesByIds(ctx, []int64{*gradeID}, lang)
		if err != nil {
			log.Warnf("classroom_exercise.resolve_grade_failed grade_id=%d err=%v", *gradeID, err)
		} else if len(grades) > 0 {
			gradeLabel = grades[0].Label()
		}
	}
	if programID != nil && *programID != 0 {
		programs, err := s.programRepo.ListProgramsByIds(ctx, []int64{*programID}, lang)
		if err != nil {
			log.Warnf("classroom_exercise.resolve_program_failed program_id=%d err=%v", *programID, err)
		} else if len(programs) > 0 {
			programLabel = programs[0].Label()
		}
	}
	return gradeLabel, programLabel
}
