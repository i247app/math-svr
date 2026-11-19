package models

type AliasModel struct {
	Audit

	Aid         int64
	Uid         int64
	Aka         string
	AliasStatus *string
}
