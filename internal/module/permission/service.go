package permission

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/role"
	dto "math-ai.com/math-ai/internal/application/dto/role"
	query "math-ai.com/math-ai/internal/application/query/role"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/role"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// roleImageUrlTTL bounds how long a presigned role image URL stays valid.
const roleImageUrlTTL = 1 * time.Hour

// Service is the permission module's façade. Today it manages only the role
// registry (ma_roles); it is scaffolding for a future permission model and
// nothing else in the system consults it yet. storageProvider may be nil; in
// that case responses simply omit role_image_url.
type Service struct {
	getRoleQuery       *query.GetRoleByIdQueryHandler
	listRolesQuery     *query.ListRolesQueryHandler
	createRoleCmd      *command.CreateRoleCommandHandler
	updateRoleCmd      *command.UpdateRoleCommandHandler
	softDeleteRoleCmd  *command.SoftDeleteRoleCommandHandler
	forceDeleteRoleCmd *command.ForceDeleteRoleCommandHandler
	storageProvider    *storage.Adapter
}

func NewService(
	roleRepo domain.IRepository,
	uow transaction.UnitOfWork,
	storageProvider *storage.Adapter,
) *Service {
	return &Service{
		getRoleQuery:       query.NewGetRoleByIdQueryHandler(roleRepo),
		listRolesQuery:     query.NewListRolesQueryHandler(roleRepo),
		createRoleCmd:      command.NewCreateRoleCommandHandler(uow),
		updateRoleCmd:      command.NewUpdateRoleCommandHandler(uow),
		softDeleteRoleCmd:  command.NewSoftDeleteRoleCommandHandler(uow),
		forceDeleteRoleCmd: command.NewForceDeleteRoleCommandHandler(uow),
		storageProvider:    storageProvider,
	}
}

func (s *Service) CreateRole(ctx context.Context, req *dto.CreateRoleReq, actorID *int64) (*dto.CreateRoleRes, error) {
	if err := ValidateCreateRole(ctx, req); err != nil {
		return nil, err
	}

	imageKey, err := s.uploadRoleImage(ctx, roleImageFile{
		reader:      req.RoleImageFile,
		filename:    req.RoleImageFilename,
		contentType: req.RoleImageContentType,
	})
	if err != nil {
		return nil, err
	}

	created, err := s.createRoleCmd.Handle(ctx, command.CreateRoleCommand{
		ActorID:      actorID,
		RoleCode:     req.RoleCode,
		RoleName:     req.RoleName,
		Description:  req.Description,
		RoleImageKey: imageKey,
		Note:         req.Note,
		RoleStatus:   req.RoleStatus,
	})
	if err != nil {
		// Nothing references the object we just uploaded.
		s.deleteRoleImage(ctx, imageKey, "create orphan")
		return nil, err
	}
	return &dto.CreateRoleRes{Role: s.toResponse(ctx, created)}, nil
}

func (s *Service) UpdateRole(ctx context.Context, req *dto.UpdateRoleReq, actorID *int64) (*dto.UpdateRoleRes, error) {
	if err := ValidateUpdateRole(ctx, req); err != nil {
		return nil, err
	}

	// Read first: an unknown role must not cost an upload, and the current
	// key is what gets deleted once a new image replaces it.
	existing, err := s.getRoleQuery.Handle(ctx, query.GetRoleByIdQuery{RoleID: req.RoleID})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if existing == nil {
		return nil, errs.NewError(ctx, status.ROLE_NOT_FOUND, nil, ErrRoleNotFound)
	}

	imageKey, err := s.uploadRoleImage(ctx, roleImageFile{
		reader:      req.RoleImageFile,
		filename:    req.RoleImageFilename,
		contentType: req.RoleImageContentType,
	})
	if err != nil {
		return nil, err
	}

	updated, err := s.updateRoleCmd.Handle(ctx, command.UpdateRoleCommand{
		ActorID:         actorID,
		RoleID:          req.RoleID,
		RoleName:        req.RoleName,
		Description:     req.Description,
		RoleImageKey:    imageKey,
		Note:            req.Note,
		RoleStatus:      req.RoleStatus,
		RemoveRoleImage: req.RemoveRoleImage,
	})
	if err != nil {
		s.deleteRoleImage(ctx, imageKey, "update orphan")
		return nil, err
	}

	// The committed row no longer points at the old image (replaced or
	// removed), so the object is garbage.
	if imageKey != nil || req.RemoveRoleImage {
		s.deleteRoleImage(ctx, existing.RoleImageKey(), "update superseded")
	}
	return &dto.UpdateRoleRes{Role: s.toResponse(ctx, updated)}, nil
}

func (s *Service) SoftDeleteRole(ctx context.Context, req *dto.DeleteRoleReq) (*dto.DeleteRoleRes, error) {
	if err := ValidateDeleteRole(ctx, req); err != nil {
		return nil, err
	}
	if err := s.softDeleteRoleCmd.Handle(ctx, command.SoftDeleteRoleCommand{RoleID: req.RoleID}); err != nil {
		return nil, err
	}
	return &dto.DeleteRoleRes{}, nil
}

func (s *Service) ForceDeleteRole(ctx context.Context, req *dto.DeleteRoleReq) (*dto.DeleteRoleRes, error) {
	if err := ValidateDeleteRole(ctx, req); err != nil {
		return nil, err
	}
	if err := s.forceDeleteRoleCmd.Handle(ctx, command.ForceDeleteRoleCommand{RoleID: req.RoleID}); err != nil {
		return nil, err
	}
	return &dto.DeleteRoleRes{}, nil
}

func (s *Service) GetRole(ctx context.Context, req *dto.GetRoleReq) (*dto.GetRoleRes, error) {
	if err := ValidateGetRole(ctx, req); err != nil {
		return nil, err
	}
	found, err := s.getRoleQuery.Handle(ctx, query.GetRoleByIdQuery{RoleID: req.RoleID})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if found == nil {
		return nil, errs.NewError(ctx, status.ROLE_NOT_FOUND, nil, ErrRoleNotFound)
	}
	return &dto.GetRoleRes{Role: s.toResponse(ctx, found)}, nil
}

func (s *Service) ListRoles(ctx context.Context, req *dto.ListRolesReq) (*dto.ListRolesRes, error) {
	if err := ValidateListRoles(ctx, req); err != nil {
		return nil, err
	}
	roles, pg, err := s.listRolesQuery.Handle(ctx, query.ListRolesQuery{
		Search:     req.Search,
		RoleStatus: req.RoleStatus,
		RoleIDs:    req.RoleIDs,
		Page:       req.Page,
		Limit:      req.Size,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	responses := dto.DomainListToResponse(roles)
	for _, r := range responses {
		s.populateRoleImageUrl(ctx, r)
	}
	return &dto.ListRolesRes{
		Roles:      responses,
		Pagination: pg,
	}, nil
}

func (s *Service) toResponse(ctx context.Context, r *domain.Role) *dto.RoleResponse {
	resp := dto.DomainToResponse(r)
	s.populateRoleImageUrl(ctx, resp)
	return resp
}
