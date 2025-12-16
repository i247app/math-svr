package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type IContactService interface {
	ListGrades(ctx context.Context, req *dto.ListContactsRequest) (status.Code, []*dto.ContactResponse, *pagination.Pagination, error)
	SubmitContact(ctx context.Context, req *dto.CreateContactRequest) (status.Code, *dto.ContactResponse, error)
	MarkReadContact(ctx context.Context, req *dto.MarkReadContactRequest) (status.Code, *dto.ContactResponse, error)
}
