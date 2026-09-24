package user

import (
	"io"

	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/pagination"
	"math-ai.com/math-ai/internal/shared/utils"
)

type UserResponse struct {
	ID              int64   `json:"id"`
	UserID          int64   `json:"uid"`
	Name            string  `json:"name"`
	Email           *string `json:"email,omitempty"`
	IsEmailVerified bool    `json:"is_email_verified"`
	Phone           *string `json:"phone,omitempty"`
	// Role and IdentityCode are both null for a guest — someone who has
	// reached the product without registering. A registered user always
	// carries a role, so their payload is unchanged.
	Role         *string `json:"role"`
	IdentityCode *string `json:"identity_code"`
	// AvatarKey is the raw S3 object key persisted on the user row.
	// AvatarUrl is a short-lived presigned URL the module layer fills
	// in on the way out (see populateImageUrl in module/user). Clients
	// should display AvatarUrl and ignore AvatarKey.
	AvatarKey *string `json:"avatar_key,omitempty"`
	AvatarUrl *string `json:"avatar_url,omitempty"`
	CreateDt  string  `json:"create_dt"`
	ModifyDt  string  `json:"modify_dt"`
}

type GetUserByPhoneReq struct {
	Phone string `json:"phone"`
}

type GetUserByPhoneRes struct {
	User *UserResponse `json:"user"`
}

type GetUserByEmailReq struct {
	Email string `json:"email"`
}

type GetUserByEmailRes struct {
	User *UserResponse `json:"user"`
}

type CreateUserReq struct {
	// UserName is the parent's display name — persisted to
	// ma_users.user_name (NOT NULL). Distinct from Name, which is the
	// child's name and lands in ma_profiles.name.
	// UserName string `json:"user_name"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email,omitempty"`
	Role  string `json:"role"`

	// Avatar is a client-supplied reference to an object already in our
	// storage. It can be either a bare S3 key (e.g.
	// "user-avatars/20260101-uuid.png") or a full URL pointing at the
	// bucket. The server normalizes it to the canonical key and persists
	// that to ma_users.avatar_key. Mutually exclusive with AvatarFile.
	Avatar string `json:"avatar,omitempty"`

	AvatarFile        io.Reader `json:"-"` // multipart file reader
	AvatarFilename    string    `json:"-"` // original filename
	AvatarContentType string    `json:"-"` // MIME type
}

// CreateUserRes carries both the freshly-created parent and their initial
// child profile so the client doesn't need a follow-up /profiles/list round
// trip during onboarding.
type CreateUserRes struct {
	User *UserResponse `json:"user"`
}

// CreateGuestReq carries nothing the client has to fill in: the body may
// be empty. Note that every call opens a NEW guest — see
// Service.CreateGuest — so the client should call it once and keep the
// token, not on every launch.
type CreateGuestReq struct {
	// ChildName names the profile opened alongside the account. Optional:
	// blank means the placeholder name, and it is only read when the
	// account is actually opened — a device coming back keeps the name it
	// already has.
	ChildName string `json:"child_name,omitempty"`
}

// CreateGuestRes hands back the guest account only. The child opened
// alongside it is read through /profiles/list like any other profile —
// the guest's session already authorises that call. The session token
// travels in the X-Auth-Token response header, exactly as it does for
// /auth/login.
type CreateGuestRes struct {
	User *UserResponse `json:"user"`
}

type UpdateUserReq struct {
	ID     int64   `json:"id"`
	UserID int64   `json:"uid"`
	Name   *string `json:"name,omitempty"`
	Email  *string `json:"email,omitempty"`
	Phone  *string `json:"phone,omitempty"`
	// Role patches ma_users.role. nil = leave unchanged; non-nil must be a
	// valid RoleType (STUDENT / TEACHER / PARENT). Mirrors the profile
	// update contract.
	Role *string `json:"role,omitempty"`

	// Avatar is a client-supplied reference to an object already in our
	// storage — either a bare S3 key or a URL pointing at the bucket.
	// Pointer semantics: nil = leave avatar_key untouched, non-nil =
	// replace (including the empty string, which the validator rejects).
	// Mutually exclusive with AvatarFile.
	Avatar *string `json:"avatar,omitempty"`

	AvatarFile        io.Reader `json:"-"` // multipart file reader
	AvatarFilename    string    `json:"-"` // original filename
	AvatarContentType string    `json:"-"` // MIME type
}

type UpdateUserRes struct {
	User *UserResponse `json:"user"`
}

type GetUserByUserIdReq struct {
	UserID int64 `json:"uid"`
}

type GetUserByUserIdRes struct {
	User *UserResponse `json:"user"`
}

type ListUsersReq struct {
	Page int64 `json:"page"`
	Size int64 `json:"size"`
}

type ListUsersRes struct {
	Users      []*UserResponse        `json:"users"`
	Pagination *pagination.Pagination `json:"pagination"`
}

type DeleteUserReq struct {
	UserID int64 `json:"uid"`
}

type DeleteUserRes struct {
}

// UploadAvatarRes mirrors profile.UploadAvatarRes — same shape so the
// mobile client can use one renderer for both. AvatarUrl is the
// presigned URL; AvatarKey is exposed so the caller can persist it for
// long-lived references (e.g. re-presigning later via /users/me).
type UploadAvatarRes struct {
	UserID    int64  `json:"uid"`
	AvatarKey string `json:"avatar_key"`
	AvatarUrl string `json:"avatar_url"`
}

func DomainToResponse(u *user.User) *UserResponse {
	if u == nil {
		return nil
	}

	phone := u.Phone()
	if phone != nil {
		normalized, _ := utils.NormalizePhone(*phone)
		if normalized != "" {
			phone = &normalized
		}
	}

	return &UserResponse{
		ID:              u.Id(),
		UserID:          u.UserId(),
		Name:            u.UserName(),
		Email:           u.Email(),
		IsEmailVerified: u.IsEmailVerified(),
		Phone:           phone,
		Role:            u.Role(),
		IdentityCode:    u.IdentityCode(),
		AvatarKey:       u.AvatarKey(),
		CreateDt:        u.CreateDt().String(),
		ModifyDt:        u.ModifyDt().String(),
	}
}

func DomainListToResponse(users []*user.User) []*UserResponse {
	result := make([]*UserResponse, len(users))
	for i, u := range users {
		result[i] = DomainToResponse(u)
	}
	return result
}

// LoginNameOf is the identifier a session records as the login name: the
// phone when the account has one, else the email. A registered account
// always has at least one of the two.
func LoginNameOf(u *UserResponse) string {
	if u.Phone != nil && *u.Phone != "" {
		return *u.Phone
	}
	return utils.DerefString(u.Email)
}
