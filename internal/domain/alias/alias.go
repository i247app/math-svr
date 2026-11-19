package alias

import (
	"math-ai.com/math-ai/internal/domain/shared/audit"
)

type Alias struct {
	audit.Audit

	aid         int64
	uid         int64
	aka         string
	aliasStatus *string
}

func NewAlias() *Alias {
	return &Alias{}
}

func (a *Alias) AID() int64       { return a.aid }
func (a *Alias) SetAID(aid int64) { a.aid = aid }

func (a *Alias) UID() int64       { return a.uid }
func (a *Alias) SetUID(uid int64) { a.uid = uid }

func (a *Alias) Aka() string       { return a.aka }
func (a *Alias) SetAka(aka string) { a.aka = aka }

func (a *Alias) AliasStatus() *string               { return a.aliasStatus }
func (a *Alias) SetAliasStatus(aliasStatus *string) { a.aliasStatus = aliasStatus }
