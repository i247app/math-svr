package models

import (
	"time"
)

type ExamPromptModel struct {
	PromptId      int64
	Grade         int
	SystemPrompt  string
	PromptVersion int
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
