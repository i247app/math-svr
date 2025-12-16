package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type IContactService interface {
	SubmitContact(ctx context.Context, req *dto.CreateContactRequest, uid string) (status.Code, *dto.ContactResponse, error)
	GetContacts(ctx context.Context, req *dto.ListContactsRequest) (status.Code, []*dto.ContactResponse, *pagination.Pagination, error)
}
