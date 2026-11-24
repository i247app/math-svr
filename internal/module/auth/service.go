package auth

import (
	"context"

	command "math-ai.com/math-ai/internal/application/command/auth"
	"math-ai.com/math-ai/internal/application/dto"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/login"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/module/device"
	"math-ai.com/math-ai/internal/module/user"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/errs"
	"math-ai.com/math-ai/internal/shared/utils"
)

type Service struct {
	userSvc   *user.Service
	deviceSvc *device.Service
	loginCmd  *command.LoginCommandHandler
	logoutCmd *command.LogoutCommandHandler
}

func NewService(
	userSvc *user.Service,
	uow transaction.UnitOfWork,
	hasher login.PasswordHasher,
	trustDeviceTTLDays int) *Service {
	return &Service{
		userSvc:   userSvc,
		loginCmd:  command.NewLoginCommandHandler(uow, hasher, trustDeviceTTLDays),
		logoutCmd: command.NewLogoutCommandHandler(uow),
	}
}

func (s *Service) Login(ctx context.Context, sess *session.AppSession, req *dto.LoginReq) (*dto.LoginRes, error) {
	log := logger.From(ctx)

	if err := ValidateLogin(ctx, req); err != nil {
		return nil, err
	}

	loginName := req.LoginName
	if utils.ValidatePhone(req.LoginName) {
		normalizePhone, err := utils.NormalizePhone(loginName)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		loginName = normalizePhone
	}

	// Whoever this session belonged to before the login — a guest, if the
	// visitor tried the product first. Captured before sess.Init below
	// overwrites it.
	previousUID, _ := sess.UID()

	result, err := s.loginCmd.Handle(ctx, command.LoginCommand{
		LoginName:       loginName,
		DeviceUUID:      metadata.GetDeviceUUID(ctx),
		DeviceName:      metadata.GetDeviceName(ctx),
		Platform:        metadata.GetPlatform(ctx),
		IPAddress:       metadata.GetIPAddress(ctx),
		DevicePushToken: metadata.GetDevicePushToken(ctx),
		Password:        req.Password,
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		res := &dto.LoginRes{User: nil}
		return res, errs.NewError(ctx, login.AuthLoginFailed, nil, ErrUserNotFound)
	}

	userRes, err := s.userSvc.GetUserById(ctx, &dto.GetUserByUidReq{UID: result.UID})
	if err != nil {
		return nil, err
	}
	if userRes == nil || userRes.User == nil {
		res := &dto.LoginRes{User: nil}
		return res, nil
	}

	if result.IsTrustedDevice {
		log.Info("Login successful, updating session data...")
		sessionData := session.InitData{
			Source:    "login",
			IsSecure:  true,
			UID:       userRes.User.UID,
			LoginName: loginName,
		}

		if userRes.User.Email != nil {
			sessionData.Email = *userRes.User.Email
		}

		sess.Init(sessionData)

		// The sign-in just proved this account is theirs, which is the
		// only thing that licenses writing to it. If they had been
		// working as a guest, their child and every exam that child sat
		// move across now.
		s.userSvc.AdoptGuestInto(ctx, previousUID, userRes.User.UID)
	}

	res := &dto.LoginRes{
		User:        userRes.User,
		IsTrusted:   result.IsTrustedDevice,
		RequiredOTP: !result.IsTrustedDevice,
	}
	return res, nil
}

// ResumeSession refreshes the identity of a session that is already signed
// in. It proves nothing about the caller, so it never makes a session secure:
// a guest gets their user back on a session that stays unsecure, and any
// other unsecure session (logged out, revoked) has to sign in again.
func (s *Service) ResumeSession(ctx context.Context, sess *session.AppSession) (*dto.LoginRes, error) {
	log := logger.From(ctx)

	if !sess.IsValid() {
		return nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrSessionNotValid)
	}

	uid, ok := sess.UID()
	if !ok {
		return nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrUIDNotFoundInSession)
	}

	userRes, err := s.userSvc.GetUserById(ctx, &dto.GetUserByUidReq{UID: uid})
	if err != nil {
		return nil, err
	}
	if userRes == nil || userRes.User == nil {
		res := &dto.LoginRes{
			User: nil,
		}
		return res, nil
	}

	if enum.IdentityCodeType(utils.DerefString(userRes.User.IdentityCode)).IsGuest() {
		// Also takes back what earlier versions granted: they made a
		// resumed guest session secure, and it lives for 14 days.
		sess.MarkNotSecure()
		return &dto.LoginRes{User: userRes.User}, nil
	}
	if !sess.IsSecure() {
		return nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrSessionNotSecure)
	}

	log.Info("Session resumed, updating session data...")
	sessionData := session.InitData{
		Source:    "login",
		IsSecure:  true,
		UID:       userRes.User.UID,
		LoginName: dto.LoginNameOf(userRes.User),
	}

	if userRes.User.Email != nil {
		sessionData.Email = *userRes.User.Email
	}

	sess.Init(sessionData)

	res := &dto.LoginRes{
		User: userRes.User,
	}
	return res, nil
}

func (s *Service) Logout(ctx context.Context, sess *session.AppSession, req *dto.LogoutReq) (*dto.LogoutRes, error) {
	if err := ValidateLogout(ctx, req); err != nil {
		return nil, err
	}

	if err := s.logoutCmd.Handle(ctx, command.LogoutCommand{
		UID:        *req.UID,
		DeviceUUID: req.DeviceUUID,
	}); err != nil {
		return nil, err
	}

	sess.MarkNotSecure()

	res := &dto.LogoutRes{}
	return res, nil
}
