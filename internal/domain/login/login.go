package login

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// Login is an account's password credential (ma_logins). It is optional:
// an account registered without a password has no row at all, and signs in
// through OTP alone. upass holds the hash, never the password itself.
type Login struct {
	loginId      int64
	userId       int64
	upass        *string
	loginsStatus *string
	rptFlg       *string
	kwords       *string
	note         *string
	status       string
	createId     *int64
	createDt     mtime.MathTime
	modifyId     *int64
	modifyDt     mtime.MathTime
}

func NewLogin() *Login {
	return &Login{}
}

func (l *Login) LoginId() int64 { return l.loginId }

func (l *Login) SetLoginId(loginId int64) { l.loginId = loginId }

func (l *Login) UserId() int64 { return l.userId }

func (l *Login) SetUserId(userId int64) { l.userId = userId }

// Upass is the password hash. Nil means the row carries no password.
func (l *Login) Upass() *string { return l.upass }

func (l *Login) SetUpass(upass *string) { l.upass = upass }

func (l *Login) LoginsStatus() *string { return l.loginsStatus }

func (l *Login) SetLoginsStatus(loginsStatus *string) { l.loginsStatus = loginsStatus }

func (l *Login) RptFlg() *string { return l.rptFlg }

func (l *Login) SetRptFlg(rptFlg *string) { l.rptFlg = rptFlg }

func (l *Login) Kwords() *string { return l.kwords }

func (l *Login) SetKwords(kwords *string) { l.kwords = kwords }

func (l *Login) Note() *string { return l.note }

func (l *Login) SetNote(note *string) { l.note = note }

func (l *Login) Status() string { return l.status }

func (l *Login) SetStatus(status string) { l.status = status }

func (l *Login) CreateId() *int64 { return l.createId }

func (l *Login) SetCreateId(createId *int64) { l.createId = createId }

func (l *Login) CreateDt() mtime.MathTime { return l.createDt }

func (l *Login) SetCreateDt(createDt mtime.MathTime) { l.createDt = createDt }

func (l *Login) ModifyId() *int64 { return l.modifyId }

func (l *Login) SetModifyId(modifyId *int64) { l.modifyId = modifyId }

func (l *Login) ModifyDt() mtime.MathTime { return l.modifyDt }

func (l *Login) SetModifyDt(modifyDt mtime.MathTime) { l.modifyDt = modifyDt }
