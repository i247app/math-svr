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
	getClassroomQuery                  *query.GetClassroomByIdQueryHandler
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
	storageProvider *storage.Adapter,
) *Service {
	return &Service{
		getClassroomQuery:                  query.NewGetClassroomByIdQueryHandler(classroomRepo, classroomProgramRepo),
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

func (s *Service) GetClassroom(ctx context.Context, req *dto.GetClassroomReq, sessionUserID int64) (*dto.GetClassroomRes, error) {
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
	if err := s.hydratePendingRequestCounts(ctx, []*classroomDomain.Classroom{found}, []*dto.ClassroomResponse{resp}); err != nil {
		return nil, err
	}
	return &dto.GetClassroomRes{Classroom: resp}, nil
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

// ForceDeleteClassroom is OWNER-gated like SoftDelete; admin-only
// gating would land here when RBAC arrives (known-issues.md §11).
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

// validateRefs makes sure the caller-supplied school / programs / grade
// references resolve to ACTIVE rows. Cheap existence check; only runs
// when the pointer is non-nil so PATCH payloads stay lean.
//
// programIDs is treated as already deduped/trimmed by the validator
// layer — we just walk and verify each one. An empty slice is allowed
// (a classroom can carry zero programs).
func (s *Service) validateRefs(ctx context.Context, schoolID *int64, programIDs []int64, gradeID *int64) error {
	if schoolID != nil && *schoolID != 0 {
		sc, err := s.schoolRepo.FindBySchoolId(ctx, *schoolID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if sc == nil {
			return errs.NewError(ctx, status.CLASSROOM_INVALID_SCHOOL, nil,
				errors.New("school not found"))
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
			return errs.NewError(ctx, status.CLASSROOM_INVALID_PROGRAM, nil,
				errors.New("program not found"))
		}
	}
	if gradeID != nil && *gradeID != 0 {
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

// hydrateOwnersAndRelationships fills the Owner and Relationship fields
// for a page of classroom responses using batched cross-aggregate
// queries — one ListByProfileIds against ma_profiles and one
// ListByProfileAndClassroomIds against ma_classroom_members. Total
// cross-aggregate cost is +2 round trips per list call, independent of
// page size, so N+1 is structurally impossible.
//
// callerProfileID is the acting profile id from the request; it drives
// the relationship lookup. When it is zero (caller did not supply it)
// every classroom defaults to Relationship=NONE so the field is still
// stable for the frontend.
func (s *Service) hydrateOwnersAndRelationships(
	ctx context.Context,
	classrooms []*classroomDomain.Classroom,
	responses []*dto.ClassroomResponse,
	callerProfileID int64,
) error {
	if len(classrooms) == 0 {
		return nil
	}

	ownerIDSet := make(map[int64]struct{}, len(classrooms))
	classroomIDs := make([]int64, 0, len(classrooms))
	for _, c := range classrooms {
		ownerIDSet[c.OwnerProfileId()] = struct{}{}
		classroomIDs = append(classroomIDs, c.ClassroomId())
	}
	ownerIDs := make([]int64, 0, len(ownerIDSet))
	for id := range ownerIDSet {
		ownerIDs = append(ownerIDs, id)
	}

	ownerMap := make(map[int64]*dto.ClassroomOwnerSummary, len(ownerIDs))
	if len(ownerIDs) > 0 && s.profileRepo != nil {
		owners, err := s.profileRepo.ListByProfileIds(ctx, ownerIDs)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		for _, p := range owners {
			summary := &dto.ClassroomOwnerSummary{
				ProfileID: p.ProfileId(),
				Name:      p.Name(),
				Role:      p.Role(),
				AvatarKey: p.AvatarKey(),
			}
			s.signOwnerAvatarURL(ctx, summary)
			ownerMap[p.ProfileId()] = summary
		}
	}

	memberMap := make(map[int64]*classroomDomain.Member, len(classroomIDs))
	if callerProfileID != 0 && s.classroomMemberRepo != nil {
		rows, err := s.classroomMemberRepo.ListByProfileAndClassroomIds(ctx, callerProfileID, classroomIDs)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		for _, m := range rows {
			memberMap[m.ClassroomId()] = m
		}
	}

	for i, c := range classrooms {
		resp := responses[i]
		if resp == nil {
			continue
		}
		if owner, ok := ownerMap[c.OwnerProfileId()]; ok {
			resp.Owner = owner
		}
		rel, role := relationshipFromMember(memberMap[c.ClassroomId()])
		resp.Relationship = string(rel)
		resp.MyRole = role
	}
	return nil
}

// relationshipFromMember maps a member row's status into the public
// relationship enum. Terminal states (REJECTED/LEFT/REMOVED) and a
// missing row both flatten to NONE — the UI treats them identically as
// "not currently participating". Returns the caller's member_role only
// when the relationship is MEMBER; nil otherwise so the JSON omits it.
func relationshipFromMember(m *classroomDomain.Member) (enum.ClassroomRelationshipType, *string) {
	if m == nil || m.MemberStatus() == nil {
		return enum.ClassroomRelationshipTypeNone, nil
	}
	switch enum.ClassroomMemberStatusType(*m.MemberStatus()) {
	case enum.ClassroomMemberStatusTypeActive:
		role := m.MemberRole()
		return enum.ClassroomRelationshipTypeMember, &role
	case enum.ClassroomMemberStatusTypePendingInvitation:
		return enum.ClassroomRelationshipTypePendingInvitation, nil
	case enum.ClassroomMemberStatusTypePendingRequest:
		return enum.ClassroomRelationshipTypePendingRequest, nil
	default:
		return enum.ClassroomRelationshipTypeNone, nil
	}
}

// hydratePendingRequestCounts groups the PENDING_REQUEST member rows
// for the given classroom ids in one round trip and writes the result
// onto each response's PendingRequestCount field. Classrooms with zero
// pending requests are absent from the repo map and default to 0.
func (s *Service) hydratePendingRequestCounts(
	ctx context.Context,
	classrooms []*classroomDomain.Classroom,
	responses []*dto.ClassroomResponse,
) error {
	if len(classrooms) == 0 || s.classroomMemberRepo == nil {
		return nil
	}
	ids := make([]int64, 0, len(classrooms))
	for _, c := range classrooms {
		ids = append(ids, c.ClassroomId())
	}
	counts, err := s.classroomMemberRepo.CountPendingRequestsByClassroomIds(ctx, ids)
	if err != nil {
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	for i, c := range classrooms {
		if responses[i] == nil {
			continue
		}
		responses[i].PendingRequestCount = counts[c.ClassroomId()]
	}
	return nil
}

// signOwnerAvatarURL mutates the owner summary in place to add a
// short-lived presigned URL for the avatar. No-op when storage is
// disabled or the owner has no avatar_key.
func (s *Service) signOwnerAvatarURL(ctx context.Context, summary *dto.ClassroomOwnerSummary) {
	if summary == nil || s.storageProvider == nil || summary.AvatarKey == nil || *summary.AvatarKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *summary.AvatarKey,
		Expiration: coverUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("classroom.owner_avatar presign failed profile_id=%d err=%v", summary.ProfileID, err)
		return
	}
	summary.AvatarURL = &url
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
