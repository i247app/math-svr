package audit

import (
	"time"
)

// Audit holds the columns every ma_* table shares (rules/database.md §3).
// Entities embed it, so its getters and setters are called on the entity
// itself (x.Note(), x.SetCreateId(...)) and are written once, here.
// The business status (<entity>_status) is not part of it: its name and
// values differ per table, so each entity keeps its own.
type Audit struct {
	rptFlg   *string
	kwords   *string
	note     *string
	status   string
	createId *int64
	createDt time.Time
	modifyId *int64
	modifyDt time.Time
}

// RptFlg is the client-side report flag (rpt_flg). Nil means "not reported".
func (a *Audit) RptFlg() *string          { return a.rptFlg }
func (a *Audit) SetRptFlg(rptFlg *string) { a.rptFlg = rptFlg }

// Kwords holds search keywords (kwords) for a future text-search index.
func (a *Audit) Kwords() *string          { return a.kwords }
func (a *Audit) SetKwords(kwords *string) { a.kwords = kwords }

func (a *Audit) Note() *string        { return a.note }
func (a *Audit) SetNote(note *string) { a.note = note }

// Status is the system state (status): ACTIVE, or INACTIVE once hidden.
func (a *Audit) Status() string          { return a.status }
func (a *Audit) SetStatus(status string) { a.status = status }

func (a *Audit) CreateId() *int64            { return a.createId }
func (a *Audit) SetCreateId(createId *int64) { a.createId = createId }

func (a *Audit) CreateDt() time.Time            { return a.createDt }
func (a *Audit) SetCreateDt(createDt time.Time) { a.createDt = createDt }

func (a *Audit) ModifyId() *int64            { return a.modifyId }
func (a *Audit) SetModifyId(modifyId *int64) { a.modifyId = modifyId }

func (a *Audit) ModifyDt() time.Time            { return a.modifyDt }
func (a *Audit) SetModifyDt(modifyDt time.Time) { a.modifyDt = modifyDt }
