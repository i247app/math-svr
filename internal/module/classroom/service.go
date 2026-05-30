package classroom

import (
	"context"
	"errors"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/classroom"
	dto "math-ai.com/math-ai/internal/application/dto/classroom"
	query "math-ai.com/math-ai/internal/application/query/classroom"
	"math-ai.com/math-ai/internal/application/transaction"
	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	gradeDomain "math-ai.com/math-ai/internal/domain/grade"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	programDomain "math-ai.com/math-ai/internal/domain/program"
	schoolDomain "math-ai.com/math-ai/internal/domain/school"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
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
	getClassroomQuery       *query.GetClassroomByIdQueryHandler
	listClassroomsQuery     *query.ListClassroomsQueryHandler
	listMembersQuery        *query.ListMembersQueryHandler
	createClassroomCmd      *command.CreateClassroomCommandHandler
	updateClassroomCmd      *command.UpdateClassroomCommandHandler
	archiveClassroomCmd     *command.ArchiveClassroomCommandHandler
	restoreClassroomCmd     *command.RestoreClassroomCommandHandler
	softDeleteClassroomCmd  *command.SoftDeleteClassroomCommandHandler
	forceDeleteClassroomCmd *command.ForceDeleteClassroomCommandHandler
	joinByCodeCmd           *command.JoinByCodeCommandHandler
	leaveClassroomCmd       *command.LeaveClassroomCommandHandler
	removeMemberCmd         *command.RemoveMemberCommandHandler
	updateMemberRoleCmd     *command.UpdateMemberRoleCommandHandler
	transferOwnershipCmd    *command.TransferOwnershipCommandHandler

	classroomRepo       classroomDomain.IRepository
	classroomMemberRepo classroomDomain.IMemberRepository
	profileRepo         profileDomain.IRepository
	programRepo         programDomain.IRepository
	gradeRepo           gradeDomain.IRepository
	schoolRepo          schoolDomain.IRepository
	storageProvider     *storage.Adapter
}

func NewService(
	classroomRepo classroomDomain.IRepository,
	classroomMemberRepo classroomDomain.IMemberRepository,
	uow transaction.UnitOfWork,
	profileRepo profileDomain.IRepository,
	programRepo programDomain.IRepository,
	gradeRepo gradeDomain.IRepository,
	schoolRepo schoolDomain.IRepository,
	storageProvider *storage.Adapter,
) *Service {
	return &Service{
		getClassroomQuery:       query.NewGetClassroomByIdQueryHandler(classroomRepo),
		listClassroomsQuery:     query.NewListClassroomsQueryHandler(classroomRepo),
		listMembersQuery:        query.NewListMembersQueryHandler(classroomMemberRepo),
		createClassroomCmd:      command.NewCreateClassroomCommandHandler(uow),
		updateClassroomCmd:      command.NewUpdateClassroomCommandHandler(uow),
		archiveClassroomCmd:     command.NewArchiveClassroomCommandHandler(uow),
		restoreClassroomCmd:     command.NewRestoreClassroomCommandHandler(uow),
		softDeleteClassroomCmd:  command.NewSoftDeleteClassroomCommandHandler(uow),
		forceDeleteClassroomCmd: command.NewForceDeleteClassroomCommandHandler(uow),
		joinByCodeCmd:           command.NewJoinByCodeCommandHandler(uow),
		leaveClassroomCmd:       command.NewLeaveClassroomCommandHandler(uow),
		removeMemberCmd:         command.NewRemoveMemberCommandHandler(uow),
		updateMemberRoleCmd:     command.NewUpdateMemberRoleCommandHandler(uow),
		transferOwnershipCmd:    command.NewTransferOwnershipCommandHandler(uow),
		classroomRepo:           classroomRepo,
		classroomMemberRepo:     classroomMemberRepo,
		profileRepo:             profileRepo,
		programRepo:             programRepo,
		gradeRepo:               gradeRepo,
		schoolRepo:              schoolRepo,
		storageProvider:         storageProvider,
	}
}

// CreateClassroom is the only entry point that needs role-gating: only
// a TEACHER profile can become an OWNER. Curriculum tie validation runs
// here so a bad program_id / grade_id is rejected before we mint a
// classroom row inside UoW.
func (s *Service) CreateClassroom(ctx context.Context, req *dto.CreateClassroomReq, sessionUserID string) (*dto.CreateClassroomRes, error) {
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
	if err := s.validateRefs(ctx, req.SchoolID, req.ProgramID, req.GradeID); err != nil {
		return nil, err
	}

	actor := caller.ProfileId()
	created, err := s.createClassroomCmd.Handle(ctx, command.CreateClassroomCommand{
		ActorID:             &actor,
		OwnerProfileID:      caller.ProfileId(),
		Name:                strings.TrimSpace(req.Name),
		Description:         req.Description,
		SchoolID:            req.SchoolID,
		ProgramID:           req.ProgramID,
		GradeID:             req.GradeID,
		MaxMembers:          req.MaxMembers,
		CoverKey:            req.CoverKey,
		Note:                req.Note,
		InviteCode:          req.InviteCode,
		InviteCodeExpiresDt: req.InviteCodeExpiresDt,
	})
	if err != nil {
		return nil, err
	}

	resp := dto.DomainToResponse(created)
	s.populateCoverUrl(ctx, resp)
	return &dto.CreateClassroomRes{Classroom: resp}, nil
}

func (s *Service) UpdateClassroom(ctx context.Context, req *dto.UpdateClassroomReq, sessionUserID string) (*dto.UpdateClassroomRes, error) {
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
	if err := s.validateRefs(ctx, req.SchoolID, req.ProgramID, req.GradeID); err != nil {
		return nil, err
	}

	actor := caller.ProfileId()
	updated, err := s.updateClassroomCmd.Handle(ctx, command.UpdateClassroomCommand{
		ActorID:     &actor,
		ClassroomID: req.ClassroomID,
		Name:        req.Name,
		Description: req.Description,
		SchoolID:    req.SchoolID,
		ProgramID:   req.ProgramID,
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

func (s *Service) GetClassroom(ctx context.Context, req *dto.GetClassroomReq, sessionUserID string) (*dto.GetClassroomRes, error) {
	if err := ValidateGetClassroom(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}

	found, err := s.getClassroomQuery.Handle(ctx, query.GetClassroomByIdQuery{ClassroomID: req.ClassroomID})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if found == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil,
			errors.New("classroom not found"))
	}
	if _, err := s.requireMember(ctx, found.ClassroomId(), caller.ProfileId()); err != nil {
		return nil, err
	}

	resp := dto.DomainToResponse(found)
	s.populateCoverUrl(ctx, resp)
	return &dto.GetClassroomRes{Classroom: resp}, nil
}

func (s *Service) ListClassrooms(ctx context.Context, req *dto.ListClassroomsReq, sessionUserID string) (*dto.ListClassroomsRes, error) {
	if err := ValidateListClassrooms(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}

	// Default "my classrooms": filter by caller's profile membership.
	// Explicit OwnerProfileID overrides; we don't apply both because the
	// AND semantics rarely match what the caller wants.
	callerProfileID := caller.ProfileId()
	var profileFilter *string
	if req.OwnerProfileID == nil {
		profileFilter = &callerProfileID
	}

	classrooms, pg, err := s.listClassroomsQuery.Handle(ctx, query.ListClassroomsQuery{
		ProfileID:       profileFilter,
		OwnerProfileID:  req.OwnerProfileID,
		SchoolID:        req.SchoolID,
		ProgramID:       req.ProgramID,
		GradeID:         req.GradeID,
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
	return &dto.ListClassroomsRes{
		Classrooms: responses,
		Pagination: pg,
	}, nil
}

func (s *Service) ArchiveClassroom(ctx context.Context, req *dto.ArchiveClassroomReq, sessionUserID string) (*dto.ArchiveClassroomRes, error) {
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

func (s *Service) RestoreClassroom(ctx context.Context, req *dto.RestoreClassroomReq, sessionUserID string) (*dto.RestoreClassroomRes, error) {
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

func (s *Service) SoftDeleteClassroom(ctx context.Context, req *dto.DeleteClassroomReq, sessionUserID string) (*dto.DeleteClassroomRes, error) {
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

// ForceDeleteClassroom is OWNER-gated like SoftDelete; admin-only
// gating would land here when RBAC arrives (known-issues.md §11).
func (s *Service) ForceDeleteClassroom(ctx context.Context, req *dto.DeleteClassroomReq, sessionUserID string) (*dto.DeleteClassroomRes, error) {
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

// validateRefs makes sure the caller-supplied school / program / grade
// references resolve to ACTIVE rows. Cheap existence check; only runs
// when the pointer is non-nil so PATCH payloads stay lean.
func (s *Service) validateRefs(ctx context.Context, schoolID, programID, gradeID *string) error {
	if schoolID != nil && strings.TrimSpace(*schoolID) != "" {
		sc, err := s.schoolRepo.FindBySchoolId(ctx, *schoolID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if sc == nil {
			return errs.NewError(ctx, status.CLASSROOM_INVALID_SCHOOL, nil,
				errors.New("school not found"))
		}
	}
	if programID != nil && strings.TrimSpace(*programID) != "" {
		p, err := s.programRepo.FindByProgramId(ctx, *programID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if p == nil {
			return errs.NewError(ctx, status.CLASSROOM_INVALID_PROGRAM, nil,
				errors.New("program not found"))
		}
	}
	if gradeID != nil && strings.TrimSpace(*gradeID) != "" {
		g, err := s.gradeRepo.FindByGradeId(ctx, *gradeID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if g == nil {
			return errs.NewError(ctx, status.CLASSROOM_INVALID_GRADE, nil,
				errors.New("grade not found"))
		}
	}
	return nil
}

// populateCoverUrl mutates resp in place to add a short-lived presigned
// URL when the classroom carries a cover_key. No-op if storage is
// disabled or the classroom has no cover.
func (s *Service) populateCoverUrl(ctx context.Context, resp *dto.ClassroomResponse) {
	if resp == nil || s.storageProvider == nil || resp.CoverKey == nil || *resp.CoverKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *resp.CoverKey,
		Expiration: coverUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("classroom.cover presign failed classroom_id=%s err=%v", resp.ClassroomID, err)
		return
	}
	resp.CoverURL = &url
}
