package models

import (
	"strings"
	"time"

	"math-ai.com/math-ai/internal/domain/shared/audit"
)

// auditColumnNames are the audit columns every ma_* table shares, in the
// order of Audit.Dest. Keep the two in lockstep.
var auditColumnNames = []string{"rpt_flg", "kwords", "note", "status", "create_id", "create_dt", "modify_id", "modify_dt"}

// Audit is the row form of audit.Audit. Models embed it.
type Audit struct {
	RptFlg   *string
	Kwords   *string
	Note     *string
	Status   string
	CreateId *int64
	CreateDt time.Time
	ModifyId *int64
	ModifyDt time.Time
}

// AuditColumns is the SELECT list of the audit columns, each qualified with
// the table alias (prefix "u" → "u.rpt_flg, …"; "" for none). It goes LAST
// in a repository's column list, matching Audit.Dest last in its Scan.
func AuditColumns(prefix string) string {
	if prefix == "" {
		return strings.Join(auditColumnNames, ", ")
	}
	return prefix + "." + strings.Join(auditColumnNames, ", "+prefix+".")
}

// Dest are the Scan destinations, in AuditColumns order.
func (m *Audit) Dest() []any {
	return []any{&m.RptFlg, &m.Kwords, &m.Note, &m.Status, &m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt}
}

// ApplyTo copies the row into a through its setters.
func (m *Audit) ApplyTo(a *audit.Audit) {
	a.SetRptFlg(m.RptFlg)
	a.SetKwords(m.Kwords)
	a.SetNote(m.Note)
	a.SetStatus(m.Status)
	a.SetCreateId(m.CreateId)
	a.SetCreateDt(m.CreateDt)
	a.SetModifyId(m.ModifyId)
	a.SetModifyDt(m.ModifyDt)
}
