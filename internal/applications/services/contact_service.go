package services

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/applications/utils"
	"math-ai.com/math-ai/internal/applications/validators"
	diRepo "math-ai.com/math-ai/internal/core/di/repositories"
	diSvc "math-ai.com/math-ai/internal/core/di/services"
	domain "math-ai.com/math-ai/internal/core/domain/contact"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ContactService struct {
	validator       validators.IContactValidator
	repo            diRepo.IContactRepository
	responseBuilder *utils.ResponseBuilder
	storageService  diSvc.IStorageService
}

func NewContactService(
	validator validators.IContactValidator,
	repo diRepo.IContactRepository,
	storageService diSvc.IStorageService,
) diSvc.IContactService {
	responseBuilder := utils.NewResponseBuilder(storageService)

	return &ContactService{
		validator:       validator,
		repo:            repo,
		responseBuilder: responseBuilder,
		storageService:  storageService,
	}
}

func (s *ContactService) GetContacts(ctx context.Context, req *dto.ListContactsRequest) (status.Code, []*dto.ContactResponse, *pagination.Pagination, error) {
	// Get total count
	totalCount, err := s.repo.CountContacts(ctx)
	if err != nil {
		return status.FAIL, nil, nil, err
	}

	// Build pagination
	paginationInfo := pagination.NewPagination(req.Page, req.Size, totalCount)

	// Build query params
	params := &dto.ListContactsParams{
		Limit:  paginationInfo.Size,
		Offset: paginationInfo.Skip,
	}

	// Handle take_all
	if req.TakeAll {
		params.Limit = totalCount
		params.Offset = 0
	}

	// Get contacts with pagination
	contacts, err := s.repo.GetContacts(ctx, params)
	if err != nil {
		return status.FAIL, nil, nil, err
	}

	// Convert domain contacts to DTO
	response := s.responseBuilder.BuildContactUsResponses(ctx, contacts)

	return status.SUCCESS, response, paginationInfo, nil
}

func (s *ContactService) SubmitContact(ctx context.Context, req *dto.CreateContactRequest, uid string) (status.Code, *dto.ContactResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateSubmitContactRequest(req); err != nil {
		return statusCode, nil, err
	}

	// Create domain object
	contact := domain.NewContactDomain()
	contact.GenerateID()
	contact.SetUID(uid) // Will be set if user is authenticated
	contact.SetContactName(req.ContactName)
	contact.SetContactEmail(req.ContactEmail)
	contact.SetContactPhone(req.ContactPhone)
	contact.SetContactMessage(req.ContactMessage)
	fmt.Println("contact ", contact)
	// Save to database
	_, err := s.repo.CreateContact(ctx, nil, contact)
	if err != nil {
		return status.FAIL, nil, fmt.Errorf("failed to create contact: %v", err)
	}

	// Return response
	response := s.responseBuilder.BuildContactUsResponse(ctx, contact)

	return status.SUCCESS, response, nil
}
