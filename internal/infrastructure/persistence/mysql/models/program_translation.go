package models

import "time"

// ProgramTranslationModel mirrors ma_program_translations. PtStatus is
// scanned from the legacy column name `gt_status` (a copy-paste typo
// inherited from the grade migration that we honour at the SQL boundary
// so existing data keeps reading cleanly).
type ProgramTranslationModel struct {
	Id                   int64
	ProgramTranslationId string
	ProgramId            string
	Language             string
	Label                string
	Description          string
	Note                 *string
	PtStatus             *string
	Status               string
	CreateId             *string
	CreateDt             time.Time
	ModifyId             *string
	ModifyDt             time.Time
}
