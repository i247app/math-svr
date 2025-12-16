package services

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/applications/utils"
	"math-ai.com/math-ai/internal/applications/validators"
	diRepo "math-ai.com/math-ai/internal/core/di/repositories"
	diSvc "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
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
	params := diRepo.ListContactsParams{
		Search:    req.Search,
		Page:      req.Page,
		Limit:     req.Limit,
		OrderBy:   req.OrderBy,
		OrderDesc: req.OrderDesc,
		TakeAll:   req.TakeAll,
	}

	
	// Get total count
	contacts, pagination, err := s.repo.List(ctx, params)
	if err != nil {
		return status.FAIL, nil, nil, err
	}	

	
	if len(contacts) == 0 {
		return status.SUCCESS, []*dto.ContactResponse{}, pagination, nil
	}
	
	// Convert domain contacts to DTO
	response := s.responseBuilder.BuildContactUsResponses(ctx, contacts)

	return status.SUCCESS, response, pagination, nil
}

func (s *ContactService) SubmitContact(ctx context.Context, req *dto.CreateContactRequest, uid string) (status.Code, *dto.ContactResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateSubmitContactRequest(req); err != nil {
		return statusCode, nil, err
	}

	// Create domain object
	contactDomain := dto.BuildContactDomainForSubmit(req, uid)

	// Save to database
	_, err := s.repo.CreateContact(ctx, nil, contactDomain)
	if err != nil {
		return status.FAIL, nil, fmt.Errorf("failed to create contact: %v", err)
	}

	// Return response
	response := s.responseBuilder.BuildContactUsResponse(ctx, contactDomain)

	return status.SUCCESS, response, nil
}

// CheckReadContact marks a contact as read and returns the updated contact.
func (s *ContactService) CheckReadContact(ctx context.Context, contactID string) (status.Code, *dto.ContactResponse, error) {
	// Validate contact exists
	contact, err := s.repo.FindByID(ctx, contactID)
	if err != nil {
		return status.FAIL, nil, fmt.Errorf("failed to find contact: %v", err)
	}
	if contact == nil {
		return status.NOT_FOUND, nil, err_svc.ErrContactNotFound
	}

	// Update is_read to true
	rowsAffected, err := s.repo.UpdateContactIsRead(ctx, contactID, true)
	if err != nil {
		return status.FAIL, nil, fmt.Errorf("failed to update contact is_read: %v", err)
	}

	if rowsAffected == 0 {
		return status.NOT_FOUND, nil, err_svc.ErrContactNotFound
	}

	// Update domain object and build response
	contact.SetIsRead(true)
	response := s.responseBuilder.BuildContactUsResponse(ctx, contact)

	return status.SUCCESS, response, nil
}
