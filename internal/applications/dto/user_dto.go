package dto

import (
	"io"
	"time"

	domain "math-ai.com/math-ai/internal/core/domain/user"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type UserResponse struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	Phone     string     `json:"phone"`
	Avatar    *string    `json:"-"`          // S3 key/path
	AvatarURL *string    `json:"avatar_url"` // Temporary presigned URL for access
	Dob       *time.Time `json:"dob"`
	Role      string     `json:"role"`
	CreateAt  time.Time  `json:"created_at"`
	ModifyAt  time.Time  `json:"modified_at"`
}

type GetUserResponse struct {
	User *UserResponse `json:"user"`
}

type ListUserRequest struct {
	Search    string `json:"search,omitempty" form:"search"`
	Page      int64  `json:"page" form:"page"`
	Limit     int64  `json:"size" form:"size"`
	OrderBy   string `json:"order_by" form:"order_by"`
	OrderDesc bool   `json:"order_desc" form:"order_desc"`
	TakeAll   bool   `json:"take_all" form:"take_all"`
}

type ListUserResponse struct {
	Items      []*UserResponse        `json:"items"`
	Pagination *pagination.Pagination `json:"metadata"`
}

type CreateUserRequest struct {
	Name     string     `json:"name"`
	Phone    string     `json:"phone"`
	Email    string     `json:"email"`
	Role     enum.ERole `json:"role,omitempty"`
	Password string     `json:"password"`
	Dob      *string    `json:"dob"`
	GradeID  string     `json:"grade_id"`
	TermID   string     `json:"term_id"`

	// Avatar upload fields (for multipart form)
	AvatarFile        io.Reader `json:"avatar_file"`         // File reader
	AvatarFilename    string    `json:"avatar_file_name"`    // Original filename
	AvatarContentType string    `json:"avatar_content_type"` // MIME type

	DeviceUUID string `json:"device_uuid,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
}

type CreateUserResponse struct {
	User *UserResponse `json:"user"`
}

type UpdateUserRequest struct {
	UID     string        `json:"uid"`
	Name    *string       `json:"name,omitempty"`
	Phone   *string       `json:"phone,omitempty"`
	Email   *string       `json:"email,omitempty"`
	Dob     *string       `json:"dob,omitempty"`
	Role    *enum.ERole   `json:"role,omitempty"`
	Status  *enum.EStatus `json:"status,omitempty"`
	GradeID *string       `json:"grade_id,omitempty"`
	TermID  *string       `json:"term_id,omitempty"`

	// Avatar upload fields (for multipart form)
	AvatarFile        io.Reader `json:"-"`                       // File reader
	AvatarFilename    string    `json:"-"`                       // Original filename
	AvatarContentType string    `json:"-"`                       // MIME type
	DeleteAvatar      bool      `json:"delete_avatar,omitempty"` // Flag to remove avatar
}

type UpdateUserResponse struct {
	User *UserResponse `json:"user"`
}

type DeleteUserRequest struct {
	UID string `json:"uid"`
}

func BuildUserDomainForCreate(req *CreateUserRequest) *domain.User {
	userDomain := domain.NewUserDomain()
	userDomain.GenerateID()
	userDomain.SetEmail(req.Email)
	userDomain.SetName(req.Name)
	userDomain.SetPhone(req.Phone)
	userDomain.SetPassword(req.Password)
	userDomain.SetGradeID(req.GradeID)
	userDomain.SetTermID(req.TermID)

	if req.Dob != nil {
		parsedDob, err := time.Parse(time.DateOnly, *req.Dob)
		if err == nil {
			userDomain.SetDOB(&parsedDob)
		}
	}

	userDomain.SetRole(string(req.Role))

	return userDomain
}

func BuildUserDomainForUpdate(req *UpdateUserRequest) *domain.User {
	userDomain := domain.NewUserDomain()
	userDomain.SetID(req.UID)

	if req.Email != nil {
		userDomain.SetEmail(*req.Email)
	}

	if req.Name != nil {
		userDomain.SetName(*req.Name)
	}

	if req.Phone != nil {
		userDomain.SetPhone(*req.Phone)
	}

	if req.Role != nil {
		userDomain.SetRole(string(*req.Role))
	}

	if req.Dob != nil {
		parsedDob, err := time.Parse(time.DateOnly, *req.Dob)
		if err == nil {
			userDomain.SetDOB(&parsedDob)
		}
	}

	return userDomain
}

func BuildAliasDomain(uid string, aka string) *domain.Alias {
	aliasDomain := domain.NewAliasDomain()
	aliasDomain.GenerateID()
	aliasDomain.SetUID(uid)
	aliasDomain.SetAka(aka)

	return aliasDomain
}

func UserResponseFromDomain(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID(),
		Email:     u.Email(),
		Name:      u.Name(),
		Phone:     u.Phone(),
		AvatarURL: u.AvatarKey(),
		Dob:       u.DOB(),
		Role:      u.Role(),
		CreateAt:  u.CreatedAt(),
		ModifyAt:  u.ModifyAt(),
	}
}

func UserResponseListFromDomain(users []*domain.User) []UserResponse {
	responses := make([]UserResponse, len(users))
	for i, u := range users {
		responses[i] = UserResponseFromDomain(u)
	}
	return responses
}
