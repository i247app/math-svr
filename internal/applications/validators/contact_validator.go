package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
	"math-ai.com/math-ai/internal/shared/utils/validate"
)

type IContactValidator interface {
	ValidateSubmitContactRequest(req *dto.CreateContactRequest) (status.Code, error)
}

type contactValidator struct{}

func NewContactValidator() *contactValidator {
	return &contactValidator{}
}

func (v *contactValidator) ValidateSubmitContactRequest(r *dto.CreateContactRequest) (status.Code, error) {
	if r.ContactPhone == nil && r.ContactEmail == nil {
		return status.CONTACT_METHOD_MISSING, err_svc.ErrContactEmailOrPhoneRequired
	}

	// Validate contact name
	if r.ContactName == "" {
		return status.CONTACT_NAME_MISSING, err_svc.ErrContactMissingName
	}
	if len(r.ContactName) > 200 {
		return status.CONTACT_NAME_TOO_LONG, err_svc.ErrContactTooLongName
	}

	// Validate contact email
	if len(*r.ContactEmail) > 200 {
		return status.CONTACT_EMAIL_TOO_LONG, err_svc.ErrContactTooLongEmail
	}
	if r.ContactEmail != nil && !validate.IsValidEmail(*r.ContactEmail) {
		return status.CONTACT_EMAIL_INVALID, err_svc.ErrContactInvalidEmail
	}

	// Validate contact message
	if r.ContactMessage == "" {
		return status.CONTACT_MESSAGE_MISSING, err_svc.ErrContactMissingMessage
	}
	if len(r.ContactMessage) > 200 {
		return status.CONTACT_MESSAGE_TOO_LONG, err_svc.ErrContactTooLongMessage
	}

	// Validate contact phone
	if r.ContactPhone != nil && !validate.IsValidPhoneNumber(*r.ContactPhone) {
		return status.CONTACT_PHONE_INVALID, err_svc.ErrContactInvalidPhone
	}

	return status.SUCCESS, nil
}

func (v *contactValidator) ValidateGetContactById(id string) (status.Code, error) {
	if id == "" {
		return status.CONTACT_MISSING_ID, err_svc.ErrContactMissingId
	}

	return status.SUCCESS, nil
}
