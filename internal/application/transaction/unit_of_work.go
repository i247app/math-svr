package transaction

import (
	"context"

	"math-ai.com/math-ai/internal/domain/device"
	"math-ai.com/math-ai/internal/domain/loginlog"
	"math-ai.com/math-ai/internal/domain/otp"
	"math-ai.com/math-ai/internal/domain/profile"
	"math-ai.com/math-ai/internal/domain/user"
)

// Repositories is the bundle of aggregate repositories bound to a single
// transaction, handed to callers of UnitOfWork.Do.
type Repositories struct {
	User     user.IRepository
	Alias    user.IAliasRepository
	Profile  profile.IRepository
	LoginLog loginlog.IRepository
	Device   device.IRepository
	Otp      otp.IRepository
}

// UnitOfWork runs fn inside a transaction, committing on nil error and
// rolling back otherwise. Implementations live in the infrastructure layer.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}
