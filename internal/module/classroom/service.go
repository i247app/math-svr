package classroom

import (
	"context"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/classroom"
	dto "math-ai.com/math-ai/internal/application/dto/classroom"
	query "math-ai.com/math-ai/internal/application/query/classroom"
	"math-ai.com/math-ai/internal/application/transaction"
	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	exerciseDomain "math-ai.com/math-ai/internal/domain/exercise"
	gradeDomain "math-ai.com/math-ai/internal/domain/grade"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	programDomain "math-ai.com/math-ai/internal/domain/program"
	schoolDomain "math-ai.com/math-ai/internal/domain/school"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// coverUrlTTL bounds how long a generated cover URL is valid. Mirrors
// the school module's imageUrlTTL so cross-aggregate UX stays consistent.
const coverUrlTTL = 1 * time.Hour

// Service is the classroom module's public façade. It composes the CQRS
// handlers behind the validators, holds the curriculum/profile repos
// needed for foreign-key validation, and exposes a sessionUserID-aware
// API so handlers can enforce the §0 Q1 contract.
type Service struct {
	getClassroomQuery                  *query.GetClassroomByIdQueryHandler
	findClassroomByCodeQuery           *query.FindClassroomByCodeQueryHandler
	listClassroomsQuery                *query.ListClassroomsQueryHandler
	listMembersQuery                   *query.ListMembersQueryHandler
	listMyPendingInvitationsQuery      *query.ListMyPendingInvitationsQueryHandler
	listPendingInvitationsByClassroomQ *query.ListPendingInvitationsByClassroomQueryHandler
	createClassroomCmd                 *command.CreateClassroomCommandHandler
	updateClassroomCmd                 *command.UpdateClassroomCommandHandler
	archiveClassroomCmd                *command.ArchiveClassroomCommandHandler
	restoreClassroomCmd                *command.RestoreClassroomCommandHandler
	softDeleteClassroomCmd             *command.SoftDeleteClassroomCommandHandler
	forceDeleteClassroomCmd            *command.ForceDeleteClassroomCommandHandler
	joinByCodeCmd                      *command.JoinByCodeCommandHandler
	leaveClassroomCmd                  *command.LeaveClassroomCommandHandler
	removeMemberCmd                    *command.RemoveMemberCommandHandler
	updateMemberRoleCmd                *command.UpdateMemberRoleCommandHandler
	transferOwnershipCmd               *command.TransferOwnershipCommandHandler
	sendInvitationCmd                  *command.SendInvitationCommandHandler
	acceptInvitationCmd                *command.AcceptInvitationCommandHandler
	rejectInvitationCmd                *command.RejectInvitationCommandHandler
	cancelInvitationCmd                *command.CancelInvitationCommandHandler
	approveJoinRequestCmd              *command.ApproveJoinRequestCommandHandler
	rejectJoinRequestCmd               *command.RejectJoinRequestCommandHandler
	cancelJoinRequestCmd               *command.CancelJoinRequestCommandHandler
	listJoinRequestsByClassroomQuery   *query.ListJoinRequestsByClassroomQueryHandler
	listMyJoinRequestsQuery            *query.ListMyJoinRequestsQueryHandler

	classroomRepo        classroomDomain.IRepository
	classroomMemberRepo  classroomDomain.IMemberRepository
	classroomProgramRepo classroomDomain.IClassroomProgramRepository
	profileRepo          profileDomain.IRepository
	programRepo          programDomain.IRepository
	gradeRepo            gradeDomain.IRepository
	schoolRepo           schoolDomain.IRepository
	storageProvider      *storage.Adapter
}

func NewService(
	classroomRepo classroomDomain.IRepository,
	classroomMemberRepo classroomDomain.IMemberRepository,
	classroomProgramRepo classroomDomain.IClassroomProgramRepository,
	uow transaction.UnitOfWork,
	profileRepo profileDomain.IRepository,
	programRepo programDomain.IRepository,
	gradeRepo gradeDomain.IRepository,
	schoolRepo schoolDomain.IRepository,
	submissionRepo exerciseDomain.ISubmissionRepository,
	exerciseRepo exerciseDomain.IRepository,
	storageProvider *storage.Adapter,
) *Service {
	return &Service{
		getClassroomQuery:                  query.NewGetClassroomByIdQueryHandler(classroomRepo, classroomProgramRepo),
		findClassroomByCodeQuery:           query.NewFindClassroomByCodeQueryHandler(classroomRepo, classroomProgramRepo),
		listClassroomsQuery:                query.NewListClassroomsQueryHandler(classroomRepo, classroomProgramRepo),
		listMembersQuery:                   query.NewListMembersQueryHandler(classroomMemberRepo),
		listMyPendingInvitationsQuery:      query.NewListMyPendingInvitationsQueryHandler(classroomMemberRepo),
		listPendingInvitationsByClassroomQ: query.NewListPendingInvitationsByClassroomQueryHandler(classroomMemberRepo),
		createClassroomCmd:                 command.NewCreateClassroomCommandHandler(uow),
		updateClassroomCmd:                 command.NewUpdateClassroomCommandHandler(uow),
		archiveClassroomCmd:                command.NewArchiveClassroomCommandHandler(uow),
		restoreClassroomCmd:                command.NewRestoreClassroomCommandHandler(uow),
		softDeleteClassroomCmd:             command.NewSoftDeleteClassroomCommandHandler(uow),
		forceDeleteClassroomCmd:            command.NewForceDeleteClassroomCommandHandler(uow),
		joinByCodeCmd:                      command.NewJoinByCodeCommandHandler(uow),
		leaveClassroomCmd:                  command.NewLeaveClassroomCommandHandler(uow),
		removeMemberCmd:                    command.NewRemoveMemberCommandHandler(uow),
		updateMemberRoleCmd:                command.NewUpdateMemberRoleCommandHandler(uow),
		transferOwnershipCmd:               command.NewTransferOwnershipCommandHandler(uow),
		sendInvitationCmd:                  command.NewSendInvitationCommandHandler(uow),
		acceptInvitationCmd:                command.NewAcceptInvitationCommandHandler(uow),
		rejectInvitationCmd:                command.NewRejectInvitationCommandHandler(uow),
		cancelInvitationCmd:                command.NewCancelInvitationCommandHandler(uow),
		approveJoinRequestCmd:              command.NewApproveJoinRequestCommandHandler(uow),
		rejectJoinRequestCmd:               command.NewRejectJoinRequestCommandHandler(uow),
		cancelJoinRequestCmd:               command.NewCancelJoinRequestCommandHandler(uow),
		listJoinRequestsByClassroomQuery:   query.NewListJoinRequestsByClassroomQueryHandler(classroomMemberRepo),
		listMyJoinRequestsQuery:            query.NewListMyJoinRequestsQueryHandler(classroomMemberRepo),
		classroomRepo:                      classroomRepo,
		classroomMemberRepo:                classroomMemberRepo,
		classroomProgramRepo:               classroomProgramRepo,
		profileRepo:                        profileRepo,
		programRepo:                        programRepo,
		gradeRepo:                          gradeRepo,
		schoolRepo:                         schoolRepo,
		storageProvider:                    storageProvider,
	}
}

// CreateClassroom is the only entry point that needs role-gating: only
// a TEACHER profile can become an OWNER. Curriculum tie validation runs
// here so a bad program_id / grade_id is rejected before we mint a
// classroom row inside UoW.
func (s *Service) CreateClassroom(ctx context.Context, req *dto.CreateClassroomReq, sessionUserID int64) (*dto.CreateClassroomRes, error) {
	if err := ValidateCreateClassroom(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if err := s.requireTeacherRole(ctx, caller); err != nil {
		return nil, err
	}
	if err := s.validateRefs(ctx, req.SchoolID, req.ProgramIDs, req.GradeID); err != nil {
		return nil, err
	}

	actor := caller.ProfileId()
	created, err := s.createClassroomCmd.Handle(ctx, command.CreateClassroomCommand{
		ActorID:                &actor,
		OwnerProfileID:         caller.ProfileId(),
		Name:                   strings.TrimSpace(req.Name),
		Description:            req.Description,
		SchoolID:               req.SchoolID,
		ProgramIDs:             req.ProgramIDs,
		GradeID:                req.GradeID,
		MaxMembers:             req.MaxMembers,
		CoverKey:               req.CoverKey,
		Note:                   req.Note,
		ClassroomCode:          req.ClassroomCode,
		ClassroomCodeExpiresDt: req.ClassroomCodeExpiresDt,
	})
	if err != nil {
		return nil, err
	}

	resp := dto.DomainToResponse(created)
	s.populateCoverUrl(ctx, resp)
	return &dto.CreateClassroomRes{Classroom: resp}, nil
}

func (s *Service) UpdateClassroom(ctx context.Context, req *dto.UpdateClassroomReq, sessionUserID int64) (*dto.UpdateClassroomRes, error) {
	if err := ValidateUpdateClassroom(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}

	existing, err := s.classroomRepo.FindByClassroomId(ctx, req.ClassroomID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if err := guardClassroomMutable(ctx, existing); err != nil {
		return nil, err
	}
	if _, err := s.requireManager(ctx, req.ClassroomID, caller.ProfileId()); err != nil {
		return nil, err
	}
	programIDsForValidate := []int64{}
	if req.ProgramIDs != nil {
		programIDsForValidate = *req.ProgramIDs
	}
	if err := s.validateRefs(ctx, req.SchoolID, programIDsForValidate, req.GradeID); err != nil {
		return nil, err
	}

	actor := caller.ProfileId()
	updated, err := s.updateClassroomCmd.Handle(ctx, command.UpdateClassroomCommand{
		ActorID:     &actor,
		ClassroomID: req.ClassroomID,
		Name:        req.Name,
		Description: req.Description,
		SchoolID:    req.SchoolID,
		ProgramIDs:  req.ProgramIDs,
		GradeID:     req.GradeID,
		MaxMembers:  req.MaxMembers,
		AvatarKey:   req.AvatarKey,
		Note:        req.Note,
	})
	if err != nil {
		return nil, err
	}

	resp := dto.DomainToResponse(updated)
	s.populateCoverUrl(ctx, resp)
	return &dto.UpdateClassroomRes{Classroom: resp}, nil
}

// GetClassroom returns the classroom by id. profile_id is OPTIONAL —
// when supplied, the service still session-validates it via
// resolveActingProfile and hydrates the caller's relationship + my_role
// against the row. When omitted, the classroom is still returned, with
// Relationship defaulting to NONE and my_role nil. Owner profile data
// is hydrated unconditionally so any caller (member or not) renders a
// complete card. The previous member-only gate is intentionally
// dropped: the Relationship field now carries the same information the
// gate used to enforce, in a form the client can branch on directly.
func (s *Service) GetClassroom(ctx context.Context, req *dto.GetClassroomReq, sessionUserID int64) (*dto.GetClassroomRes, error) {
	if err := ValidateGetClassroom(ctx, req); err != nil {
		return nil, err
	}

	var callerProfileID int64
	if req.ProfileID != 0 {
		caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
		if err != nil {
			return nil, err
		}
		callerProfileID = caller.ProfileId()
	}

	found, err := s.getClassroomQuery.Handle(ctx, query.GetClassroomByIdQuery{ClassroomID: req.ClassroomID})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if found == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil,
			ErrClassroomNotFound)
	}

	resp := dto.DomainToResponse(found)
	s.populateCoverUrl(ctx, resp)
	// hydrateOwnersAndRelationships handles callerProfileID == 0 — it
	// skips the membership lookup and leaves Relationship=NONE, but
	// always hydrates the owner summary, so the get-by-id response
	// shape matches the list response shape regardless of profile_id.
	if err := s.hydrateOwnersAndRelationships(
		ctx,
		[]*classroomDomain.Classroom{found},
		[]*dto.ClassroomResponse{resp},
		callerProfileID,
	); err != nil {
		return nil, err
	}
	if err := s.hydratePendingRequestCounts(ctx, []*classroomDomain.Classroom{found}, []*dto.ClassroomResponse{resp}); err != nil {
		return nil, err
	}
	return &dto.GetClassroomRes{Classroom: resp}, nil
}

// FindClassroomByCode resolves a classroom from its human-readable join
// code. Read-only preview path that does NOT mutate membership — the
// client typically calls this before /classrooms/join-by-code to show
// the user what they're about to join. Owner is hydrated for the card
// render; Relationship is left at NONE because the endpoint takes no
// profile_id (the caller may not even own a profile yet).
func (s *Service) FindClassroomByCode(ctx context.Context, req *dto.FindClassroomByCodeReq) (*dto.FindClassroomByCodeRes, error) {
	if err := ValidateFindClassroomByCode(ctx, req); err != nil {
		return nil, err
	}

	found, err := s.findClassroomByCodeQuery.Handle(ctx, query.FindClassroomByCodeQuery{ClassroomCode: req.ClassCode})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if found == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil, ErrClassroomNotFound)
	}

	resp := dto.DomainToResponse(found)
	s.populateCoverUrl(ctx, resp)
	if err := s.hydrateOwnersAndRelationships(
		ctx,
		[]*classroomDomain.Classroom{found},
		[]*dto.ClassroomResponse{resp},
		req.ProfileID,
	); err != nil {
		return nil, err
	}
	if err := s.hydratePendingRequestCounts(ctx, []*classroomDomain.Classroom{found}, []*dto.ClassroomResponse{resp}); err != nil {
		return nil, err
	}
	return &dto.FindClassroomByCodeRes{Classroom: resp}, nil
}

func (s *Service) ListClassrooms(ctx context.Context, req *dto.ListClassroomsReq, sessionUserID int64) (*dto.ListClassroomsRes, error) {
	if err := ValidateListClassrooms(ctx, req); err != nil {
		return nil, err
	}
	// ProfileID is intentionally NOT forwarded as a query filter — when
	// the repo sees it, it inner-joins ma_classroom_members and drops
	// every row where the profile isn't an ACTIVE member, which hides
	// classrooms the caller could still see via Relationship=NONE /
	// PENDING_*. The hydration step below uses req.ProfileID
	// independently to fill the per-row relationship column.
	classrooms, pg, err := s.listClassroomsQuery.Handle(ctx, query.ListClassroomsQuery{
		OwnerProfileID:  req.OwnerProfileID,
		SchoolID:        req.SchoolID,
		ProgramID:       req.ProgramID,
		ProgramIDs:      req.ProgramIDs,
		GradeID:         req.GradeID,
		Search:          req.Search,
		ClassCode:       req.ClassCode,
		IncludeArchived: req.IncludeArchived,
		Page:            req.Page,
		Limit:           req.Size,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	responses := dto.DomainListToResponse(classrooms)
	for _, r := range responses {
		s.populateCoverUrl(ctx, r)
	}
	if err := s.hydrateOwnersAndRelationships(ctx, classrooms, responses, req.ProfileID); err != nil {
		return nil, err
	}
	if err := s.hydratePendingRequestCounts(ctx, classrooms, responses); err != nil {
		return nil, err
	}
	return &dto.ListClassroomsRes{
		Classrooms: responses,
		Pagination: pg,
	}, nil
}

// ListMyJoinedClassrooms is the membership-scoped counterpart to
// ListClassrooms. Forwarding ProfileID into the query intentionally
// triggers the repo's ACTIVE-member inner-join — only classrooms the
// caller is currently an ACTIVE member of are returned. Relationship
// will hydrate to MEMBER for every row.
func (s *Service) ListMyJoinedClassrooms(ctx context.Context, req *dto.ListMyJoinedClassroomsReq, sessionUserID int64) (*dto.ListMyJoinedClassroomsRes, error) {
	if err := ValidateListMyJoinedClassrooms(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}

	profileID := caller.ProfileId()
	classrooms, pg, err := s.listClassroomsQuery.Handle(ctx, query.ListClassroomsQuery{
		ProfileID:       &profileID,
		Search:          req.Search,
		IncludeArchived: req.IncludeArchived,
		Page:            req.Page,
		Limit:           req.Size,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	responses := dto.DomainListToResponse(classrooms)
	for _, r := range responses {
		s.populateCoverUrl(ctx, r)
	}
	if err := s.hydrateOwnersAndRelationships(ctx, classrooms, responses, profileID); err != nil {
		return nil, err
	}
	if err := s.hydratePendingRequestCounts(ctx, classrooms, responses); err != nil {
		return nil, err
	}
	return &dto.ListMyJoinedClassroomsRes{
		Classrooms: responses,
		Pagination: pg,
	}, nil
}

func (s *Service) ArchiveClassroom(ctx context.Context, req *dto.ArchiveClassroomReq, sessionUserID int64) (*dto.ArchiveClassroomRes, error) {
	if err := ValidateArchiveClassroom(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireOwner(ctx, req.ClassroomID, caller.ProfileId()); err != nil {
		return nil, err
	}
	if err := s.archiveClassroomCmd.Handle(ctx, command.ArchiveClassroomCommand{ClassroomID: req.ClassroomID}); err != nil {
		return nil, err
	}
	return &dto.ArchiveClassroomRes{}, nil
}

func (s *Service) RestoreClassroom(ctx context.Context, req *dto.RestoreClassroomReq, sessionUserID int64) (*dto.RestoreClassroomRes, error) {
	if err := ValidateRestoreClassroom(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireOwner(ctx, req.ClassroomID, caller.ProfileId()); err != nil {
		return nil, err
	}
	if err := s.restoreClassroomCmd.Handle(ctx, command.RestoreClassroomCommand{ClassroomID: req.ClassroomID}); err != nil {
		return nil, err
	}
	return &dto.RestoreClassroomRes{}, nil
}

func (s *Service) SoftDeleteClassroom(ctx context.Context, req *dto.DeleteClassroomReq, sessionUserID int64) (*dto.DeleteClassroomRes, error) {
	if err := ValidateDeleteClassroom(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireOwner(ctx, req.ClassroomID, caller.ProfileId()); err != nil {
		return nil, err
	}
	if err := s.softDeleteClassroomCmd.Handle(ctx, command.SoftDeleteClassroomCommand{ClassroomID: req.ClassroomID}); err != nil {
		return nil, err
	}
	return &dto.DeleteClassroomRes{}, nil
}

func (s *Service) ForceDeleteClassroom(ctx context.Context, req *dto.DeleteClassroomReq, sessionUserID int64) (*dto.DeleteClassroomRes, error) {
	if err := ValidateDeleteClassroom(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireOwner(ctx, req.ClassroomID, caller.ProfileId()); err != nil {
		return nil, err
	}
	if err := s.forceDeleteClassroomCmd.Handle(ctx, command.ForceDeleteClassroomCommand{ClassroomID: req.ClassroomID}); err != nil {
		return nil, err
	}
	return &dto.DeleteClassroomRes{}, nil
}

func (s *Service) validateRefs(ctx context.Context, schoolID *int64, programIDs []int64, gradeID *int64) error {
	if schoolID != nil && *schoolID != 0 {
		sc, err := s.schoolRepo.FindBySchoolId(ctx, *schoolID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if sc == nil {
			return errs.NewError(ctx, status.CLASSROOM_INVALID_SCHOOL, nil, ErrSchoolNotFound)
		}
	}
	for _, pid := range programIDs {
		if pid == 0 {
			continue
		}
		p, err := s.programRepo.FindByProgramId(ctx, pid, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if p == nil {
			return errs.NewError(ctx, status.CLASSROOM_INVALID_PROGRAM, nil, ErrProgramNotFound)
		}
	}
	if gradeID != nil && *gradeID != 0 {
		g, err := s.gradeRepo.FindByGradeId(ctx, *gradeID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if g == nil {
			return errs.NewError(ctx, status.CLASSROOM_INVALID_GRADE, nil, ErrGradeNotFound)
		}
	}
	return nil
}
