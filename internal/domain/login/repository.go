package login

import "context"

type IRepository interface {
	FindByLoginId(ctx context.Context, loginId int64) (*Login, error)
	// FindByUserId returns the account's credential, or (nil, nil) when the
	// account was registered without a password.
	FindByUserId(ctx context.Context, userId int64) (*Login, error)
	Create(ctx context.Context, l *Login) (*Login, error)
}

// PasswordHasher turns a password into a storable hash and checks one
// against it. Compare returns a non-nil error on any mismatch. The bcrypt
// implementation lives in infrastructure/security and is injected by
// bootstrap, so application code never imports the crypto package.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(password string, hash string) error
}
