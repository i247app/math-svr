package user

import (
	"context"

	command "math-ai.com/math-ai/internal/application/command/user"
	dto "math-ai.com/math-ai/internal/application/dto/user"
	query "math-ai.com/math-ai/internal/application/query/user"
	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	domain "math-ai.com/math-ai/internal/domain/user"
)

type Service struct {
	getUserByUserIdQuery *query.GetUserByUserIdQueryHandler
	getUserByPhoneQuery  *query.GetUserByPhoneQueryHandler
	listUsersQuery       *query.ListUsersQueryHandler
	createUserCmd        *command.CreateUserCommandHandler
	updateUserCmd        *command.UpdateUserCommandHandler
	softDeleteUserCmd    *command.SoftDeleteUserCommandHandler
	forceDeleteUserCmd   *command.ForceDeleteUserCommandHandler
}

func NewService(
	repo domain.IRepository,
	uow transaction.UnitOfWork,
) *Service {
	return &Service{
		getUserByUserIdQuery: query.NewGetUserByUserIdQueryHandler(repo),
		getUserByPhoneQuery:  query.NewGetUserByPhoneQueryHandler(repo),
		listUsersQuery:       query.NewListUsersQueryHandler(repo),
		createUserCmd:        command.NewCreateUserCommandHandler(uow),
		updateUserCmd:        command.NewUpdateUserCommandHandler(uow),
		softDeleteUserCmd:    command.NewSoftDeleteUserCommandHandler(uow),
		forceDeleteUserCmd:   command.NewForceDeleteUserCommandHandler(uow),
	}
}

func (s *Service) GetUserById(ctx context.Context, req *dto.GetUserByUserIdReq) (*dto.GetUserByUserIdRes, error) {
	user, err := s.getUserByUserIdQuery.Handle(ctx, query.GetUserByUserIdQuery{UserId: req.UserID})
	if err != nil {
		return nil, err
	}

	res := &dto.GetUserByUserIdRes{
		User: dto.DomainToResponse(user),
	}

	return res, nil
}

func (s *Service) ListUsers(ctx context.Context, req *dto.ListUsersReq) (*dto.ListUsersRes, error) {
	users, pg, err := s.listUsersQuery.Handle(ctx, &query.ListUsersQuery{
		Page:  int64(req.Page),
		Limit: int64(req.Size),
	})
	if err != nil {
		return nil, err
	}

	return &dto.ListUsersRes{
		Users:      dto.DomainListToResponse(users),
		Pagination: pg,
	}, nil
}

func (s *Service) GetUserByPhone(ctx context.Context, req *dto.GetUserByPhoneReq) (*dto.GetUserByPhoneRes, error) {
	user, err := s.getUserByPhoneQuery.Handle(ctx, query.GetUserByPhoneQuery{Phone: req.Phone})
	if err != nil {
		return nil, err
	}

	res := &dto.GetUserByPhoneRes{
		User: dto.DomainToResponse(user),
	}

	return res, nil
}

func (s *Service) CreateUser(ctx context.Context, req *dto.CreateUserReq) (*dto.CreateUserRes, error) {
	if err := ValidateCreateUser(ctx, req); err != nil {
		return nil, err
	}

	created, err := s.createUserCmd.Handle(ctx, command.CreateUserCommand{
		Email: req.Email,
		Phone: req.Phone,
	})
	if err != nil {
		return nil, err
	}

	return &dto.CreateUserRes{
		User: dto.DomainToResponse(created),
	}, nil
}

func (s *Service) SoftDeleteUser(ctx context.Context, req *dto.DeleteUserReq) (*dto.DeleteUserRes, error) {
	if err := ValidateDeleteUser(ctx, req); err != nil {
		return nil, err
	}

	user, err := s.getUserByUserIdQuery.Handle(ctx, query.GetUserByUserIdQuery{UserId: req.UserID})
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errs.NewError(ctx, status.NOT_FOUND, nil, nil)
	}

	if err := s.softDeleteUserCmd.Handle(ctx, command.SoftDeleteUserCommand{
		UserID: req.UserID,
	}); err != nil {
		return nil, err
	}

	return &dto.DeleteUserRes{}, nil
}

func (s *Service) ForceDeleteUser(ctx context.Context, req *dto.DeleteUserReq) (*dto.DeleteUserRes, error) {
	if err := ValidateDeleteUser(ctx, req); err != nil {
		return nil, err
	}

	if err := s.forceDeleteUserCmd.Handle(ctx, command.ForceDeleteUserCommand{
		UserID: req.UserID,
	}); err != nil {
		return nil, err
	}

	return &dto.DeleteUserRes{}, nil
}

func (s *Service) UpdateUser(ctx context.Context, req *dto.UpdateUserReq) (*dto.UpdateUserRes, error) {
	if err := ValidateUpdateUser(ctx, req); err != nil {
		return nil, err
	}

	user, err := s.updateUserCmd.Handle(ctx, command.UpdateUserCommand{
		ID:     req.ID,
		UserID: req.UserID,
		Email:  req.Email,
		Phone:  req.Phone,
	})
	if err != nil {
		return nil, err
	}

	res := &dto.UpdateUserRes{
		User: dto.DomainToResponse(user),
	}

	return res, nil
}
