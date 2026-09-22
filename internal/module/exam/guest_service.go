package exam

import (
	"context"

	userCommand "math-ai.com/math-ai/internal/application/command/user"
	"math-ai.com/math-ai/internal/application/transaction"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	userDomain "math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// guestService is the exam module's one point of contact with the user
// aggregate. Generating an exam is where a visitor first meets the
// product, so it is also where their account gets opened — but opening
// an account is a user-aggregate concern, so the work itself lives in
// command/user and only the call sits here. Keeping it out of service.go
// mirrors what bot_service.go does for the LLM adapter.
type guestService struct {
	createGuest *userCommand.CreateGuestCommandHandler
	userRepo    userDomain.IRepository
	profileRepo profileDomain.IRepository
}

func newGuestService(uow transaction.UnitOfWork, userRepo userDomain.IRepository, profileRepo profileDomain.IRepository) *guestService {
	return &guestService{
		createGuest: userCommand.NewCreateGuestCommandHandler(uow),
		userRepo:    userRepo,
		profileRepo: profileRepo,
	}
}

// GuestIdentity is what the caller needs after a guest is recognised or
// opened: the uid to put in the session, and the profile the exam will
// be generated for.
type GuestIdentity struct {
	UserID    int64
	ProfileID int64
	// Existing distinguishes a device coming back from a device seen for
	// the first time. The caller only logs it — the exam path is the same
	// either way, which is the point.
	Existing bool
}

// EnsureGuest returns the guest account this device belongs to, opening
// one on first sight. Idempotent by device_uuid, so a visitor who closes
// the app and comes back keeps the exams they already sat.
func (g *guestService) EnsureGuest(ctx context.Context, deviceUUID, childName string) (*GuestIdentity, error) {
	res, err := g.createGuest.Handle(ctx, userCommand.CreateGuestCommand{
		DeviceUUID: deviceUUID,
		ChildName:  childName,
	})
	if err != nil {
		return nil, err
	}

	identity := &GuestIdentity{
		UserID:    res.User.UserId(),
		ProfileID: res.Profile.ProfileId(),
		Existing:  res.Existing,
	}
	if res.Existing {
		logger.From(ctx).Info("exam.guest.recognised", "uid", identity.UserID, "profile_id", identity.ProfileID)
	} else {
		logger.From(ctx).Info("exam.guest.opened", "uid", identity.UserID, "profile_id", identity.ProfileID)
	}
	return identity, nil
}

// DefaultProfileOf resolves which child a GUEST means when they did not
// say. A guest has exactly one profile — the server opened it for them —
// so asking them to name it is friction with only one possible answer.
//
// It deliberately answers only for guests: a registered parent can have
// several children, and silently picking one of them would hand back the
// wrong child's exam. They still have to state profile_id.
//
// (0, false, nil) means "not a guest, or nothing to resolve" — the caller
// then falls through to the ordinary validation error.
func (g *guestService) DefaultProfileOf(ctx context.Context, userID int64) (int64, bool, error) {
	u, err := g.userRepo.FindByUserId(ctx, userID)
	if err != nil {
		return 0, false, err
	}
	if u == nil || !enum.IdentityCodeType(utils.DerefString(u.IdentityCode())).IsGuest() {
		return 0, false, nil
	}

	p, err := g.profileRepo.FindDefaultProfileByUserId(ctx, userID)
	if err != nil {
		return 0, false, err
	}
	if p == nil {
		return 0, false, nil
	}
	return p.ProfileId(), true, nil
}

// IsGuest reports whether a uid belongs to an account that has never
// registered. It is the gate every /exams/* route consults for a session
// that is not secure: a guest never gets a secure session — that would
// hand them every other auth-gated route in the product — so this is
// what distinguishes "a guest on their own session" from "a real user
// who has not finished logging in".
func (g *guestService) IsGuest(ctx context.Context, userID int64) (bool, error) {
	u, err := g.userRepo.FindByUserId(ctx, userID)
	if err != nil {
		return false, err
	}
	if u == nil {
		return false, nil
	}
	return enum.IdentityCodeType(utils.DerefString(u.IdentityCode())).IsGuest(), nil
}
