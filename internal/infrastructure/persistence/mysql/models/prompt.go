package models

import (
	"time"
)

type PromptModel struct {
	PromptId      int64
	PromptCode    string
	Grade         *int
	Language      *string
	SystemPrompt  string
	PromptVersion int
	Description   *string
	PromptStatus  *string
	RptFlg        *string
	Kwords        *string
	Note          *string
	Status        string
	CreateId      *int64
	CreateDt      time.Time
	ModifyId      *int64
	ModifyDt      time.Time
}
