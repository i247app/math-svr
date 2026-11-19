package login

import (
	"math-ai.com/math-ai/internal/domain/shared/audit"
)

// Login is an account's password credential (ma_logins). It is optional:
// an account registered without a password has no row at all, and signs in
// through OTP alone. upw holds the hash, never the password itself.
type Login struct {
	audit.Audit

	loginId      int64
	uid          int64
	upw          *string
	loginsStatus *string
}

func NewLogin() *Login {
	return &Login{}
}

func (l *Login) LoginId() int64           { return l.loginId }
func (l *Login) SetLoginId(loginId int64) { l.loginId = loginId }

func (l *Login) UID() int64       { return l.uid }
func (l *Login) SetUID(uid int64) { l.uid = uid }

// Upw is the password hash. Nil means the row carries no password.
func (l *Login) Upw() *string       { return l.upw }
func (l *Login) SetUpw(upw *string) { l.upw = upw }

func (l *Login) LoginsStatus() *string                { return l.loginsStatus }
func (l *Login) SetLoginsStatus(loginsStatus *string) { l.loginsStatus = loginsStatus }
